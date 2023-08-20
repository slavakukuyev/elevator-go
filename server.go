package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

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
	s.httpServer = &http.Server{

		Addr:    addr,
		Handler: http.DefaultServeMux,
	}

	s.httpServer.Handler = http.HandlerFunc(s.handler)
	return s
}

// handler is a method of Server that handles incoming requests.
func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
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

	// Request an elevator going from floor 1 to floor 9
	if err := s.manager.RequestElevator(requestBody.From, requestBody.To); err != nil {
		logger.Error("request elevator error",
			zap.Error(err),
			zap.Int("from", requestBody.From),
			zap.Int("to", requestBody.To),
		)

		http.Error(w, "request elevator error", http.StatusInternalServerError)
		return
	}

	response := fmt.Sprintf("Received request: from %d to %d", requestBody.From, requestBody.To)
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
