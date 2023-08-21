package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/slavakukuyev/elevator-go/metrics"
	"go.uber.org/zap"
)

// Server represents the HTTP server.
type Server struct {
	manager    *Manager
	httpServer *http.Server
}

// RequestBody represents the JSON request body.
type RequestBody struct {
	From int `json:"from"`
	To   int `json:"to"`
}

// NewServer creates a new instance of Server.
func NewServer(port int, manager *Manager) *Server {
	s := &Server{
		manager: manager,
	}

	addr := fmt.Sprintf(":%d", port)

	// Create a new ServeMux to handle different routes
	mux := http.NewServeMux()

	// Register your custom handler for the "/elevator" route
	mux.HandleFunc("/elevator", s.handler)

	// Register Prometheus metrics handler for the "/metrics" route
	mux.Handle("/metrics", promhttp.Handler())

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return s
}

// handler is a method of Server that handles incoming requests.
func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	var elevatorName string
	defer func() {
		requestDuration(elevatorName, startTime)
	}()

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var requestBody RequestBody
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Request an elevator going from floor 0 to floor 9
	elevator, err := s.manager.RequestElevator(requestBody.From, requestBody.To)
	if err != nil {
		logger.Error("request elevator error",
			zap.Error(err),
			zap.Int("from", requestBody.From),
			zap.Int("to", requestBody.To),
		)

		http.Error(w, "request elevator error", http.StatusInternalServerError)
		return
	}

	if elevator != nil {
		elevatorName = elevator.name
	}

	response := fmt.Sprintf("Elevator %s received request: from %d to %d", elevatorName, requestBody.From, requestBody.To)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

func (s *Server) Start() {
	logger.Info("Server started", zap.String("Addr", s.httpServer.Addr))
	err := s.httpServer.ListenAndServe()
	if err != http.ErrServerClosed {
		logger.Error("Server error on start", zap.Error(err))
	}
}

func (s *Server) Shutdown() {
	logger.Info("Shutting down the server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown error:", zap.Error(err))
	}

	os.Exit(0)
}

func requestDuration(elevatorName string, start time.Time) {
	end := time.Now()
	durationIMilliseconds := end.Sub(start).Milliseconds()
	durationInSeconds := float64(durationIMilliseconds) / 1000.0
	metrics.RequestDurationHistogram(elevatorName, durationInSeconds)
}
