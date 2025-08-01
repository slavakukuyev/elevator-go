package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/slavakukuyev/elevator-go/internal/manager"
)

// WebSocketServer is a separate server just for WebSocket connections
type WebSocketServer struct {
	manager     *manager.Manager
	server      *http.Server
	logger      *slog.Logger
	ctx         context.Context
	cancel      context.CancelFunc
	connections map[*websocket.Conn]context.CancelFunc
	connMutex   sync.RWMutex
}

// Simple upgrader without any special configuration
var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
	// Set buffer sizes for better performance
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Enable compression for better performance
	EnableCompression: true,
}

// NewWebSocketServer creates a new WebSocket-only server
func NewWebSocketServer(port int, manager *manager.Manager, logger *slog.Logger) *WebSocketServer {
	ctx, cancel := context.WithCancel(context.Background())
	mux := http.NewServeMux()

	ws := &WebSocketServer{
		manager:     manager,
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
		connections: make(map[*websocket.Conn]context.CancelFunc),
	}

	// Add CORS headers manually for WebSocket endpoint
	mux.HandleFunc("/ws/status", func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Upgrade, Connection, Sec-WebSocket-Key, Sec-WebSocket-Version")

		ws.statusHandler(w, r)
	})

	ws.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return ws
}

// addConnection adds a connection to the tracking map
func (ws *WebSocketServer) addConnection(conn *websocket.Conn, cancel context.CancelFunc) {
	ws.connMutex.Lock()
	defer ws.connMutex.Unlock()
	ws.connections[conn] = cancel
}

// removeConnection removes a connection from the tracking map
func (ws *WebSocketServer) removeConnection(conn *websocket.Conn) {
	ws.connMutex.Lock()
	defer ws.connMutex.Unlock()
	if cancel, exists := ws.connections[conn]; exists {
		cancel()
		delete(ws.connections, conn)
	}
}

// closeAllConnections gracefully closes all active WebSocket connections
func (ws *WebSocketServer) closeAllConnections() {
	ws.connMutex.Lock()
	defer ws.connMutex.Unlock()

	for conn, cancel := range ws.connections {
		// Send close message
		if err := conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Server shutdown"),
			time.Now().Add(1*time.Second)); err != nil {
			ws.logger.Error("failed to send close message", slog.String("error", err.Error()))
		}
		cancel()
		if err := conn.Close(); err != nil {
			ws.logger.Error("failed to close WebSocket connection", slog.String("error", err.Error()))
		}
	}
	// Clear the map
	ws.connections = make(map[*websocket.Conn]context.CancelFunc)
}

// statusHandler handles WebSocket connections with proper connection management
func (ws *WebSocketServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	// WebSocket upgrade
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		ws.logger.Error("WebSocket upgrade failed", slog.String("error", err.Error()))
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			ws.logger.Error("failed to close WebSocket connection", slog.String("error", err.Error()))
		}
	}()

	// Create a context for this connection that cancels when the connection closes
	ctx, cancel := context.WithCancel(ws.ctx)
	ws.addConnection(conn, cancel)
	defer ws.removeConnection(conn)

	ws.logger.Info("WebSocket connection established", slog.String("component", "websocket-server"))

	// Set up connection timeouts
	const (
		writeWait      = 10 * time.Second
		pongWait       = 60 * time.Second
		pingPeriod     = (pongWait * 9) / 10
		statusInterval = 100 * time.Millisecond // Update every 100ms for real-time movement
	)

	// Set read deadline and pong handler for keep-alive
	if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		ws.logger.Error("failed to set read deadline", slog.String("error", err.Error()))
		return
	}
	conn.SetPongHandler(func(string) error {
		if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			ws.logger.Error("failed to set read deadline in pong handler", slog.String("error", err.Error()))
		}
		return nil
	})

	// Send initial status
	status, err := ws.manager.GetStatus()
	if err != nil {
		ws.logger.Error("Failed to get initial status", slog.String("component", "websocket-server"), slog.String("error", err.Error()))
		return
	}

	if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		ws.logger.Error("failed to set write deadline for initial status", slog.String("error", err.Error()))
		return
	}
	if err := conn.WriteJSON(status); err != nil {
		ws.logger.Error("Failed to send initial status", slog.String("component", "websocket-server"), slog.String("error", err.Error()))
		return
	}

	// Create tickers for status updates and ping messages
	statusTicker := time.NewTicker(statusInterval)
	defer statusTicker.Stop()

	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	// Channel to signal when connection is closed
	done := make(chan struct{})

	// Start a goroutine to handle incoming messages (mainly pong responses)
	go func() {
		defer close(done)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					ws.logger.Warn("WebSocket connection closed unexpectedly", slog.String("component", "websocket-server"), slog.String("error", err.Error()))
				}
				return
			}
		}
	}()

	// Main message loop
	for {
		select {
		case <-done:
			ws.logger.Info("WebSocket connection closed by client", slog.String("component", "websocket-server"))
			return

		case <-ctx.Done():
			ws.logger.Info("WebSocket connection context cancelled", slog.String("component", "websocket-server"))
			// Send close message to client
			if err := conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Server shutdown"), time.Now().Add(writeWait)); err != nil {
				ws.logger.Error("failed to send close message", slog.String("error", err.Error()))
			}
			return

		case <-pingTicker.C:
			if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				ws.logger.Error("failed to set write deadline for ping", slog.String("error", err.Error()))
				return
			}
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				ws.logger.Error("Failed to send ping message", slog.String("component", "websocket-server"), slog.String("error", err.Error()))
				return
			}

		case <-statusTicker.C:
			// Get status with timeout
			type wsStatusResult struct {
				status map[string]interface{}
				err    error
			}
			statusCh := make(chan wsStatusResult, 1)
			go func() {
				st, errS := ws.manager.GetStatus()
				statusCh <- wsStatusResult{status: st, err: errS}
			}()

			var st map[string]interface{}
			var errS error

			select {
			case <-time.After(5 * time.Second):
				ws.logger.Error("Failed to get status", slog.String("component", "websocket-server"), slog.String("error", "internal: status collection timed out: context canceled"))
				continue
			case result := <-statusCh:
				st = result.status
				errS = result.err
			}

			if errS != nil {
				ws.logger.Error("Failed to get status", slog.String("component", "websocket-server"), slog.String("error", errS.Error()))
				continue
			}

			// Send status update with timeout
			if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				ws.logger.Error("failed to set write deadline for status update", slog.String("error", err.Error()))
				return
			}
			if err := conn.WriteJSON(st); err != nil {
				ws.logger.Error("Failed to send status", slog.String("component", "websocket-server"), slog.String("error", err.Error()))
				return
			}
		}
	}
}

// Start starts the WebSocket server
func (ws *WebSocketServer) Start() error {
	ws.logger.Info("Starting WebSocket server", slog.String("addr", ws.server.Addr))
	return ws.server.ListenAndServe()
}

// Shutdown gracefully shuts down the WebSocket server
func (ws *WebSocketServer) Shutdown(ctx context.Context) error {
	ws.cancel()
	ws.closeAllConnections()
	return ws.server.Shutdown(ctx)
}
