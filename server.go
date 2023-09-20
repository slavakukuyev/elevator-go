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
	logger     *zap.Logger
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
// - logger (*zap.Logger): A pointer to the logger instance.
//
// Returns:
// - A pointer to the new Server instance.
//
// Example Usage:
//
//	manager := NewManager(zap.NewNop())
//	logger, _ := zap.NewDevelopment()
//	server := NewServer(8080, manager, logger)
func NewServer(port int, manager *Manager, logger *zap.Logger) *Server {
	s := &Server{
		manager: manager,
		logger:  logger.With(zap.String("module", "server")),
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
		s.logger.Error("request floor error",
			zap.Error(err),
			zap.Int("from", requestBody.From),
			zap.Int("to", requestBody.To),
		)

		http.Error(w, "request floor error", http.StatusInternalServerError)
		return
	}

	if elevator != nil {
		elevatorName = elevator.name
	}

	response := fmt.Sprintf("elevator %s received request: from %d to %d", elevatorName, requestBody.From, requestBody.To)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(response)); err != nil {
		s.logger.Error("response write error", zap.Error(err))
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

	// Request an elevator going from floor 0 to floor 9
	elevator, err := NewElevator(requestBody.Name, requestBody.MinFloor, requestBody.MaxFloor, cfg.EachFloorDuration, cfg.OpenDoorDuration, s.logger)
	if err != nil {
		s.logger.Error("request elevator error",
			zap.Error(err),
			zap.Int("from", requestBody.MinFloor),
			zap.Int("to", requestBody.MaxFloor),
		)

		http.Error(w, "request elevator error", http.StatusInternalServerError)
		return
	}

	s.manager.AddElevator(elevator)

	response := fmt.Sprintf("elevator %s has been created successfully", elevator.name)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(response)); err != nil {
		s.logger.Error("response write error", zap.Error(err))
	}
}

func (s *Server) Start() {
	s.logger.Info("Server started", zap.String("Addr", s.httpServer.Addr))
	err := s.httpServer.ListenAndServe()
	if err != http.ErrServerClosed {
		s.logger.Error("Server error on start", zap.Error(err))
	}
}

func (s *Server) Shutdown() {
	s.logger.Info("Shutting down the server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Server shutdown error:", zap.Error(err))
	}

	os.Exit(0)
}

func requestDuration(elevatorName string, start time.Time) {
	end := time.Now()
	durationIMilliseconds := end.Sub(start).Milliseconds()
	durationInSeconds := float64(durationIMilliseconds) / 1000.0
	metrics.RequestDurationHistogram(elevatorName, durationInSeconds)
}
