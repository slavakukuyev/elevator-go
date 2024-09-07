package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/slavakukuyev/elevator-go/internal/infra/config"
	"github.com/slavakukuyev/elevator-go/internal/manager"
	"github.com/slavakukuyev/elevator-go/metrics"
)

// Server represents the HTTP server.
type Server struct {
	manager    *manager.T
	httpServer *http.Server
	cfg        *config.Config
}

// FloorRequestBody represents the JSON request body.
type FloorRequestBody struct {
	From int `json:"from"`
	To   int `json:"to"`
}

// ElevatorRequestBody - represents the JSON request body.
type ElevatorRequestBody struct {
	Name     string `json:"name"`
	MinFloor int    `json:"min_floor"`
	MaxFloor int    `json:"max_floor"`
}

// NewServer creates a new instance of Server.
//
// Parameters:
// - port (int): The port number to listen on.
// - manager (*Manager): A pointer to the Manager instance.
//
// Returns:
// - A pointer to the new Server instance.
//
// Example Usage:
//
//	manager := NewManager(slog.NewNop())
//	server := NewServer(8080, manager)
func NewServer(cfg *config.Config, port int, manager *manager.T) *Server {
	s := &Server{
		manager: manager,
		cfg:     cfg,
	}

	addr := fmt.Sprintf(":%d", port)

	// Create a new ServeMux to handle different routes
	mux := http.NewServeMux()

	// Register your custom handler for the "/elevator" route
	mux.HandleFunc("/floor", s.floorHandler)
	mux.HandleFunc("/elevator", s.elevatorHandler)

	// Register Prometheus metrics handler for the "/metrics" route
	mux.Handle("/metrics", promhttp.Handler())

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return s
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
	startTime := time.Now()
	var elevatorName string
	defer func() {
		requestDuration(elevatorName, startTime)
	}()

	if r.Method != http.MethodPost {
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var requestBody FloorRequestBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestBody)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Request an elevator going from floor 0 to floor 9
	elevator, err := s.manager.RequestElevator(requestBody.From, requestBody.To)
	if err != nil {
		slog.Error("request floor error",
			slog.String("err", err.Error()),
			slog.Int("from", requestBody.From),
			slog.Int("to", requestBody.To),
		)

		http.Error(w, "request floor error", http.StatusInternalServerError)
		return
	}

	if elevator != nil {
		elevatorName = elevator.Name()
	}

	response := fmt.Sprintf("elevator %s received request: from %d to %d", elevatorName, requestBody.From, requestBody.To)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(response)); err != nil {
		slog.Error("response write error", slog.String("err", err.Error()))
	}
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
	if r.Method != http.MethodPost {
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var requestBody ElevatorRequestBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestBody)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err = s.manager.AddElevator(s.cfg, requestBody.Name, requestBody.MinFloor, requestBody.MaxFloor, s.cfg.EachFloorDuration, s.cfg.OpenDoorDuration)
	if err != nil {
		slog.Error("request elevator error",
			slog.String("err", err.Error()),
			slog.Int("from", requestBody.MinFloor),
			slog.Int("to", requestBody.MaxFloor),
		)

		http.Error(w, "request elevator error", http.StatusInternalServerError)
		return
	}

	response := fmt.Sprintf("elevator %s has been created successfully", requestBody.Name)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(response)); err != nil {
		slog.Error("response write error", slog.String("err", err.Error()))
	}
}

func (s *Server) Start() {
	slog.Info("Server started", slog.String("Addr", s.httpServer.Addr))
	err := s.httpServer.ListenAndServe()
	if err != http.ErrServerClosed {
		slog.Error("Server error on start", slog.String("err", err.Error()))
	}
}

func (s *Server) Shutdown() {
	slog.Info("Shutting down the server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown error:", slog.String("err", err.Error()))
	}

	os.Exit(0)
}

func requestDuration(elevatorName string, start time.Time) {
	end := time.Now()
	durationIMilliseconds := end.Sub(start).Milliseconds()
	durationInSeconds := float64(durationIMilliseconds) / 1000.0
	metrics.RequestDurationHistogram(elevatorName, durationInSeconds)
}
