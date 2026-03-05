package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/slavakukuyev/elevator-go/internal/constants"
	"github.com/slavakukuyev/elevator-go/internal/domain"
	"github.com/slavakukuyev/elevator-go/internal/infra/config"
	"github.com/slavakukuyev/elevator-go/internal/infra/health"
	"github.com/slavakukuyev/elevator-go/internal/infra/logging"
	"github.com/slavakukuyev/elevator-go/internal/manager"
	"github.com/slavakukuyev/elevator-go/metrics"
)

// Server represents the HTTP server.
type Server struct {
	manager       *manager.Manager
	httpServer    *http.Server
	cfg           *config.Config
	logger        *slog.Logger
	healthService *health.HealthService
}

// FloorRequestBody represents the JSON request body.
type FloorRequestBody struct {
	From int `json:"from"`
	To   int `json:"to"`
}

// ElevatorRequestBody - represents the JSON request body.
type ElevatorRequestBody struct {
	Name              string `json:"name"`
	MinFloor          int    `json:"min_floor"`
	MaxFloor          int    `json:"max_floor"`
	OverloadThreshold *int   `json:"overload_threshold,omitempty"` // Optional: defaults to 12 if not provided
}

// upgrader is used to upgrade HTTP connections to WebSocket connections.
var upgrader = websocket.Upgrader{
	// Allow all origins for demonstration purposes.
	CheckOrigin: func(r *http.Request) bool { return true },
	// Set buffer sizes for better performance
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Disable compression as it can cause issues with some proxies
	EnableCompression: false,
	// Add error handler to get more details about upgrade failures
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		// Log the error for debugging
		fmt.Printf("WebSocket upgrade error: %v (status: %d)\n", reason, status)
		http.Error(w, reason.Error(), status)
	},
}

// NewServer creates a new instance of Server with versioned API and middleware.
//
// Parameters:
// - cfg (*config.Config): The configuration instance.
// - port (int): The port number to listen on.
// - manager (*manager.Manager): A pointer to the Manager instance.
//
// Returns:
// - A pointer to the new Server instance.
//
// Example Usage:
//
//	cfg := &config.Config{}
//	manager := manager.New(cfg, factory)
//	server := NewServer(cfg, 8080, manager)
func NewServer(cfg *config.Config, port int, manager *manager.Manager) *Server {
	s := &Server{
		manager:       manager,
		cfg:           cfg,
		logger:        slog.With(slog.String("component", constants.ComponentHTTPServer)),
		healthService: health.NewHealthService(30 * time.Second), // 30 second cache TTL
	}

	// Initialize health checks
	s.setupHealthChecks(manager)

	addr := fmt.Sprintf(":%d", port)

	// Create versioned handlers
	v1Handlers := NewV1Handlers(manager, cfg, s.logger)

	// Create rate limiter using configuration
	rateLimiter := NewRateLimitMiddleware(cfg.RateLimitRPM, s.logger)

	// Create middleware chain
	middlewareChain := ChainMiddleware(
		RequestIDMiddleware(),
		LoggingMiddleware(s.logger),
		RecoveryMiddleware(s.logger),
		CORSMiddleware(),
		SecurityHeadersMiddleware(),
		rateLimiter.Handler(),
	)

	// Create a new ServeMux to handle different routes
	mux := http.NewServeMux()

	// === V1 API ROUTES (New versioned API) ===
	mux.HandleFunc("/v1", v1Handlers.APIInfoHandler)
	mux.HandleFunc("/v1/floors/request", v1Handlers.FloorRequestHandler)
	mux.HandleFunc("/v1/elevators", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			v1Handlers.ElevatorCreateHandler(w, r)
		case http.MethodDelete:
			v1Handlers.ElevatorDeleteHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/v1/health", v1Handlers.HealthHandler)
	mux.HandleFunc("/v1/metrics", v1Handlers.MetricsHandler)

	// Enhanced health endpoints
	mux.HandleFunc("/v1/health/live", s.livenessHandler)
	mux.HandleFunc("/v1/health/ready", s.readinessHandler)
	mux.HandleFunc("/v1/health/detailed", s.detailedHealthHandler)

	// === LEGACY ROUTES (Backward compatibility) ===
	mux.HandleFunc("/floor", s.floorHandler)
	mux.HandleFunc("/elevator", s.elevatorHandler)
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/metrics/system", s.systemMetricsHandler)

	// === MONITORING ROUTES ===
	// Prometheus metrics handler
	mux.Handle("/metrics", promhttp.Handler())

	// Add WebSocket routes directly to main mux
	mux.HandleFunc("/ws/status", s.statusWebSocketHandler)

	// Apply middleware chain to all routes
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      middlewareChain(mux),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return s
}

// setupHealthChecks initializes and registers health check components
func (s *Server) setupHealthChecks(manager *manager.Manager) {
	// System resource checker
	s.healthService.Register(health.NewSystemResourceChecker(85.0, 1000))

	// Liveness checker
	s.healthService.Register(health.NewLivenessChecker())

	// Manager health checker
	managerHealthChecker := health.NewComponentHealthChecker("manager", func(ctx context.Context) (bool, string, map[string]interface{}) {
		elevators := manager.GetElevators()

		// System is healthy when there are 0 elevators - this is a valid initial state
		if len(elevators) == 0 {
			return true, "System ready for elevator creation", map[string]interface{}{
				"elevator_count": 0,
				"system_state":   "initial_setup",
			}
		}

		healthyCount := 0
		for _, e := range elevators {
			healthMetrics := e.GetHealthMetrics()
			if isHealthy, ok := healthMetrics["is_healthy"].(bool); ok && isHealthy {
				healthyCount++
			}
		}

		details := map[string]interface{}{
			"total_elevators":   len(elevators),
			"healthy_elevators": healthyCount,
			"health_ratio":      float64(healthyCount) / float64(len(elevators)),
		}

		if healthyCount == 0 {
			return false, "No healthy elevators", details
		}

		if float64(healthyCount)/float64(len(elevators)) < 0.5 {
			return false, "Less than 50% of elevators are healthy", details
		}

		return true, "Manager and elevators are healthy", details
	})
	s.healthService.Register(managerHealthChecker)

	// Readiness checker (depends on manager)
	readinessChecker := health.NewReadinessChecker(managerHealthChecker)
	s.healthService.Register(readinessChecker)

	s.logger.Info("health checks initialized",
		slog.Int("registered_checkers", 4))
}

// livenessHandler handles liveness probe requests
func (s *Server) livenessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := s.healthService.Check(r.Context(), "liveness")
	if err != nil {
		http.Error(w, "Liveness check failed", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if result.Status == health.StatusHealthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		// Log error but can't write to response as headers may already be sent
		log.Printf("failed to encode response: %v", err)
	}
}

// readinessHandler handles readiness probe requests
func (s *Server) readinessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := s.healthService.Check(r.Context(), "readiness")
	if err != nil {
		http.Error(w, "Readiness check failed", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if result.Status == health.StatusHealthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		// Log error but can't write to response as headers may already be sent
		log.Printf("failed to encode response: %v", err)
	}
}

// detailedHealthHandler provides comprehensive health status
func (s *Server) detailedHealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	overallStatus, results := s.healthService.GetOverallStatus(r.Context())

	response := map[string]interface{}{
		"status":    string(overallStatus),
		"timestamp": time.Now(),
		"checks":    results,
		"summary": map[string]interface{}{
			"total_checks":     len(results),
			"healthy_checks":   countChecksWithStatus(results, health.StatusHealthy),
			"degraded_checks":  countChecksWithStatus(results, health.StatusDegraded),
			"unhealthy_checks": countChecksWithStatus(results, health.StatusUnhealthy),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	var statusCode int
	switch overallStatus {
	case health.StatusUnhealthy:
		statusCode = http.StatusServiceUnavailable
	case health.StatusDegraded:
		statusCode = http.StatusOK // Still serving traffic but degraded
	default:
		statusCode = http.StatusOK
	}

	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Log error but can't write to response as headers may already be sent
		log.Printf("failed to encode response: %v", err)
	}
}

// countChecksWithStatus counts health checks with a specific status
func countChecksWithStatus(results map[string]health.CheckResult, status health.Status) int {
	count := 0
	for _, result := range results {
		if result.Status == status {
			count++
		}
	}
	return count
}

// floorHandler is a method of Server that handles incoming floor requests.
//
// Parameters:
// - w (http.ResponseWriter): The response writer to send the HTTP response.
// - r (*http.Request): The HTTP request received.
//
// Returns:
// - None.
//
// Example Usage:
//
//	requestBody := FloorRequestBody{
//	  From: 0,
//	  To:   9,
//	}
//	requestBodyBytes, _ := json.Marshal(requestBody)
//	req, _ := http.NewRequest("POST", "/floor", bytes.NewBuffer(requestBodyBytes))
//	resp, _ := http.DefaultClient.Do(req)
//
//	// Check the response
//	body, _ := ioutil.ReadAll(resp.Body)
//	fmt.Println(string(body))
func (s *Server) floorHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logging.NewContextWithCorrelation(r.Context())
	startTime := time.Now()
	var elevatorName string
	defer func() {
		requestDuration(elevatorName, startTime)
	}()

	if r.Method != http.MethodPost {
		s.logger.WarnContext(ctx, "invalid request method for floor endpoint",
			slog.String("method", r.Method),
			slog.String("expected", "POST"))
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var requestBody FloorRequestBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestBody)
	if err != nil {
		validationErr := domain.NewValidationError("invalid request body format", err)
		s.logger.ErrorContext(ctx, "failed to decode floor request",
			slog.String("error", validationErr.Error()))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate client input floors before processing
	_, err = domain.NewFloorWithValidation(requestBody.From)
	if err != nil {
		s.logger.ErrorContext(ctx, "invalid from floor in client request",
			slog.Int("from_floor", requestBody.From),
			slog.String("error", err.Error()))
		http.Error(w, "invalid from floor: "+err.Error(), http.StatusBadRequest)
		return
	}

	_, err = domain.NewFloorWithValidation(requestBody.To)
	if err != nil {
		s.logger.ErrorContext(ctx, "invalid to floor in client request",
			slog.Int("to_floor", requestBody.To),
			slog.String("error", err.Error()))
		http.Error(w, "invalid to floor: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Request an elevator going from floor requestBody.From to floor requestBody.To
	elevator, err := s.manager.RequestElevator(ctx, requestBody.From, requestBody.To)
	if err != nil {
		s.logger.ErrorContext(ctx, "elevator request failed",
			slog.Int("from_floor", requestBody.From),
			slog.Int("to_floor", requestBody.To),
			slog.String("error", err.Error()))

		// Determine HTTP status based on error type
		statusCode := http.StatusInternalServerError
		if domainErr, ok := err.(*domain.DomainError); ok {
			switch domainErr.Type {
			case domain.ErrTypeValidation:
				statusCode = http.StatusBadRequest
			case domain.ErrTypeNotFound:
				statusCode = http.StatusNotFound
			case domain.ErrTypeConflict:
				statusCode = http.StatusConflict
			}
		}

		http.Error(w, "elevator request failed", statusCode)
		return
	}

	if elevator != nil {
		elevatorName = elevator.Name()
	}

	response := fmt.Sprintf("elevator %s received request: from %d to %d", elevatorName, requestBody.From, requestBody.To)
	w.Header().Set("Content-Type", constants.ContentTypeTextPlain)
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(response)); err != nil {
		s.logger.ErrorContext(ctx, "failed to write response",
			slog.String("error", err.Error()))
	}

	s.logger.InfoContext(ctx, "floor request processed successfully",
		slog.String("elevator_name", elevatorName),
		slog.Int("from_floor", requestBody.From),
		slog.Int("to_floor", requestBody.To),
		slog.String("direction", determineDirection(requestBody.From, requestBody.To)),
		slog.String("component", constants.ComponentHTTPHandler))
}

// determineDirection returns the direction string based on floor movement
func determineDirection(from, to int) string {
	if to > from {
		return string(domain.DirectionUp)
	}
	return string(domain.DirectionDown)
}

// elevatorHandler is a method of Server that handles incoming elevator requests.
//
// Parameters:
// - w (http.ResponseWriter): The response writer to send the HTTP response.
// - r (*http.Request): The HTTP request received.
//
// Returns:
// - None.
//
// Example Usage:
//
//	requestBody := ElevatorRequestBody{
//	  Name:      "Elevator1",
//	  MinFloor:  0,
//	  MaxFloor:  9,
//	}
//	requestBodyBytes, _ := json.Marshal(requestBody)
//	req, _ := http.NewRequest("POST", "/elevator", bytes.NewBuffer(requestBodyBytes))
//	resp, _ := http.DefaultClient.Do(req)
//
//	// Check the response
//	body, _ := ioutil.ReadAll(resp.Body)
//	fmt.Println(string(body))
func (s *Server) elevatorHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logging.NewContextWithCorrelation(r.Context())

	if r.Method != http.MethodPost {
		s.logger.WarnContext(ctx, "invalid request method for elevator endpoint",
			slog.String("method", r.Method),
			slog.String("expected", "POST"))
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var requestBody ElevatorRequestBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestBody)
	if err != nil {
		validationErr := domain.NewValidationError("invalid request body format", err)
		s.logger.ErrorContext(ctx, "failed to decode elevator request",
			slog.String("error", validationErr.Error()))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate client input floors for elevator creation
	_, err = domain.NewFloorWithValidation(requestBody.MinFloor)
	if err != nil {
		s.logger.ErrorContext(ctx, "invalid min floor in elevator creation request",
			slog.Int("min_floor", requestBody.MinFloor),
			slog.String("error", err.Error()))
		http.Error(w, "invalid min floor: "+err.Error(), http.StatusBadRequest)
		return
	}

	_, err = domain.NewFloorWithValidation(requestBody.MaxFloor)
	if err != nil {
		s.logger.ErrorContext(ctx, "invalid max floor in elevator creation request",
			slog.Int("max_floor", requestBody.MaxFloor),
			slog.String("error", err.Error()))
		http.Error(w, "invalid max floor: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = s.manager.AddElevator(ctx, s.cfg, requestBody.Name, requestBody.MinFloor, requestBody.MaxFloor, s.cfg.EachFloorDuration, s.cfg.OpenDoorDuration, s.cfg.DefaultOverloadThreshold)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create elevator",
			slog.String("elevator_name", requestBody.Name),
			slog.Int("min_floor", requestBody.MinFloor),
			slog.Int("max_floor", requestBody.MaxFloor),
			slog.String("error", err.Error()))

		// Determine HTTP status based on error type
		statusCode := http.StatusInternalServerError
		if domainErr, ok := err.(*domain.DomainError); ok {
			switch domainErr.Type {
			case domain.ErrTypeValidation:
				statusCode = http.StatusBadRequest
			case domain.ErrTypeConflict:
				statusCode = http.StatusConflict
			}
		}

		http.Error(w, "elevator creation failed", statusCode)
		return
	}

	response := fmt.Sprintf("elevator %s has been created successfully", requestBody.Name)
	w.Header().Set("Content-Type", constants.ContentTypeTextPlain)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(response)); err != nil {
		s.logger.ErrorContext(ctx, "failed to write response",
			slog.String("error", err.Error()))
	}

	s.logger.InfoContext(ctx, "elevator created successfully",
		slog.String("elevator_name", requestBody.Name),
		slog.Int("min_floor", requestBody.MinFloor),
		slog.Int("max_floor", requestBody.MaxFloor),
		slog.String("component", constants.ComponentHTTPHandler))
}

// GetHandler returns the HTTP handler for testing purposes
func (s *Server) GetHandler() http.Handler {
	return s.httpServer.Handler
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

func requestDuration(elevatorName string, start time.Time) {
	end := time.Now()
	durationIMilliseconds := end.Sub(start).Milliseconds()
	durationInSeconds := float64(durationIMilliseconds) / 1000.0
	metrics.RequestDurationHistogram(elevatorName, durationInSeconds)
}

// statusWebSocketHandler handles WebSocket connections for elevator status updates.
// It periodically sends the current status (retrieved from the manager) to the connected client.
func (s *Server) statusWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logging.NewContextWithCorrelation(r.Context())

	// Upgrade the connection to WebSocket.
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to upgrade connection to WebSocket",
			slog.String("error", err.Error()))
		return
	}
	defer func(ws *websocket.Conn) {
		errOnClose := ws.Close()
		if errOnClose != nil {
			s.logger.ErrorContext(ctx, "failed to close WebSocket connection",
				slog.String("error", errOnClose.Error()))
		}
	}(ws)

	s.logger.InfoContext(ctx, "WebSocket connection established")

	// Send an initial status immediately upon connection.
	status, err := s.manager.GetStatus()
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get initial elevator status",
			slog.String("error", err.Error()))
		return
	}
	if err := ws.WriteJSON(status); err != nil {
		s.logger.ErrorContext(ctx, "failed to send initial status via WebSocket",
			slog.String("error", err.Error()))
		return
	}

	// Create tickers for status updates and ping/pong keep-alive
	statusTicker := time.NewTicker(s.cfg.StatusUpdateInterval)
	defer statusTicker.Stop()

	pingTicker := time.NewTicker(s.cfg.WebSocketPingInterval)
	defer pingTicker.Stop()

	// Use request context without timeout to allow long-running connections
	wsCtx := ctx

	// Set up pong handler for keep-alive
	if err := ws.SetReadDeadline(time.Now().Add(s.cfg.WebSocketReadTimeout)); err != nil {
		s.logger.ErrorContext(ctx, "failed to set read deadline",
			slog.String("error", err.Error()))
		return
	}
	ws.SetPongHandler(func(string) error {
		if err := ws.SetReadDeadline(time.Now().Add(s.cfg.WebSocketReadTimeout)); err != nil {
			s.logger.ErrorContext(ctx, "failed to set read deadline in pong handler",
				slog.String("error", err.Error()))
		}
		return nil
	})

	// Start a goroutine to handle incoming messages (mainly pong responses)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					s.logger.WarnContext(ctx, "WebSocket connection closed unexpectedly",
						slog.String("error", err.Error()))
				}
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			s.logger.InfoContext(ctx, "WebSocket connection closed by client")
			return

		case <-wsCtx.Done():
			s.logger.InfoContext(ctx, "WebSocket connection context cancelled")
			// Send close message to client
			if err := ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Server shutdown"), time.Now().Add(s.cfg.WebSocketWriteTimeout)); err != nil {
				s.logger.ErrorContext(ctx, "failed to send close message",
					slog.String("error", err.Error()))
			}
			return

		case <-pingTicker.C:
			// Send ping message to keep connection alive
			if err := ws.SetWriteDeadline(time.Now().Add(s.cfg.WebSocketWriteTimeout)); err != nil {
				s.logger.ErrorContext(ctx, "failed to set write deadline for ping",
					slog.String("error", err.Error()))
				return
			}
			if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				s.logger.ErrorContext(ctx, "failed to send ping message",
					slog.String("error", err.Error()))
				return
			}

		case <-statusTicker.C:
			// Create a timeout context for this status update using configuration
			updateCtx, updateCancel := context.WithTimeout(wsCtx, s.cfg.StatusUpdateTimeout)

			// Retrieve the latest elevator status with timeout
			statusCh := make(chan statusResult, 1)
			go func() {
				st, errS := s.manager.GetStatus()
				statusCh <- statusResult{status: st, err: errS}
			}()

			var st map[string]interface{}
			var errS error

			select {
			case <-updateCtx.Done():
				s.logger.WarnContext(ctx, "status update timed out")
				updateCancel()
				continue
			case result := <-statusCh:
				st = result.status
				errS = result.err
			}
			updateCancel()

			if errS != nil {
				s.logger.ErrorContext(ctx, "failed to get status update",
					slog.String("error", errS.Error()))
				continue
			}

			// Set write deadline and send the updated status to the client
			if err := ws.SetWriteDeadline(time.Now().Add(s.cfg.WebSocketWriteTimeout)); err != nil {
				s.logger.ErrorContext(ctx, "failed to set write deadline for status update",
					slog.String("error", err.Error()))
				return
			}
			err = ws.WriteJSON(st)
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to send status update via WebSocket",
					slog.String("error", err.Error()))
				return
			}
		}
	}
}

// statusResult is a helper struct for handling status updates with timeouts
type statusResult struct {
	status map[string]interface{}
	err    error
}

// healthHandler handles health check requests
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logging.NewContextWithCorrelation(r.Context())

	if r.Method != http.MethodGet {
		s.logger.WarnContext(ctx, "invalid request method for health endpoint",
			slog.String("method", r.Method),
			slog.String("expected", "GET"))
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}

	health, err := s.manager.GetHealthStatus()
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get health status",
			slog.String("error", err.Error()))
		http.Error(w, "health check failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Determine overall health status code
	statusCode := http.StatusOK
	if systemHealthy, ok := health["system_healthy"].(bool); ok && !systemHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(health); err != nil {
		s.logger.ErrorContext(ctx, "failed to encode health response",
			slog.String("error", err.Error()))
	}

	s.logger.InfoContext(ctx, "health check request processed",
		slog.Int("status_code", statusCode))
}

// systemMetricsHandler handles system metrics requests
func (s *Server) systemMetricsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := logging.NewContextWithCorrelation(r.Context())

	if r.Method != http.MethodGet {
		s.logger.WarnContext(ctx, "invalid request method for metrics endpoint",
			slog.String("method", r.Method),
			slog.String("expected", "GET"))
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}

	metrics := s.manager.GetMetrics()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		s.logger.ErrorContext(ctx, "failed to encode metrics response",
			slog.String("error", err.Error()))
	}

	s.logger.InfoContext(ctx, "metrics request processed")
}
