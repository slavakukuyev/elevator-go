package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/slavakukuyev/elevator-go/internal/constants"
	"github.com/slavakukuyev/elevator-go/internal/domain"
	"github.com/slavakukuyev/elevator-go/internal/infra/config"
	"github.com/slavakukuyev/elevator-go/internal/infra/logging"
	"github.com/slavakukuyev/elevator-go/internal/manager"
)

// V1Handlers contains all v1 API handlers
type V1Handlers struct {
	manager *manager.Manager
	cfg     *config.Config
	logger  *slog.Logger
}

// NewV1Handlers creates a new V1Handlers instance
func NewV1Handlers(manager *manager.Manager, cfg *config.Config, logger *slog.Logger) *V1Handlers {
	return &V1Handlers{
		manager: manager,
		cfg:     cfg,
		logger:  logger,
	}
}

// FloorRequestResponse represents the response for floor requests
type FloorRequestResponse struct {
	ElevatorName string `json:"elevator_name"`
	FromFloor    int    `json:"from_floor"`
	ToFloor      int    `json:"to_floor"`
	Direction    string `json:"direction"`
	Message      string `json:"message"`
}

// ElevatorCreateResponse represents the response for elevator creation
type ElevatorCreateResponse struct {
	Name     string `json:"name"`
	MinFloor int    `json:"min_floor"`
	MaxFloor int    `json:"max_floor"`
	Message  string `json:"message"`
}

// ElevatorDeleteRequest represents the request for elevator deletion
type ElevatorDeleteRequest struct {
	Name string `json:"name"`
}

// ElevatorDeleteResponse represents the response for elevator deletion
type ElevatorDeleteResponse struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]interface{} `json:"checks"`
}

// MetricsResponse represents the metrics response
type MetricsResponse struct {
	Timestamp time.Time              `json:"timestamp"`
	Metrics   map[string]interface{} `json:"metrics"`
}

// APIInfoResponse represents API information
type APIInfoResponse struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Endpoints   map[string]string `json:"endpoints"`
}

// FloorRequestHandler handles v1 floor requests (POST /v1/floors/request)
func (h *V1Handlers) FloorRequestHandler(w http.ResponseWriter, r *http.Request) {
	requestID := logging.GetRequestID(r.Context())
	rw := NewResponseWriter(w, h.logger, requestID)

	if r.Method != http.MethodPost {
		h.logger.WarnContext(r.Context(), "invalid request method for floor endpoint",
			slog.String("method", r.Method),
			slog.String("expected", "POST"),
			slog.String("request_id", requestID))
		rw.WriteError(http.StatusMethodNotAllowed, ErrorCodeMethodNotAllowed,
			"Method not allowed", "Only POST method is supported")
		return
	}

	var requestBody FloorRequestBody
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestBody); err != nil {
		h.logger.ErrorContext(r.Context(), "failed to decode floor request",
			slog.String("error", err.Error()),
			slog.String("request_id", requestID))
		rw.WriteError(http.StatusBadRequest, ErrorCodeInvalidJSON,
			"Invalid JSON", "Request body contains invalid JSON")
		return
	}

	// Validate client input floors before processing
	if _, err := domain.NewFloorWithValidation(requestBody.From); err != nil {
		h.logger.ErrorContext(r.Context(), "invalid from floor in client request",
			slog.Int("from_floor", requestBody.From),
			slog.String("error", err.Error()),
			slog.String("request_id", requestID))
		rw.WriteDomainError(err)
		return
	}

	if _, err := domain.NewFloorWithValidation(requestBody.To); err != nil {
		h.logger.ErrorContext(r.Context(), "invalid to floor in client request",
			slog.Int("to_floor", requestBody.To),
			slog.String("error", err.Error()),
			slog.String("request_id", requestID))
		rw.WriteDomainError(err)
		return
	}

	// Request an elevator
	elevator, err := h.manager.RequestElevator(r.Context(), requestBody.From, requestBody.To)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "elevator request failed",
			slog.Int("from_floor", requestBody.From),
			slog.Int("to_floor", requestBody.To),
			slog.String("error", err.Error()),
			slog.String("request_id", requestID))
		rw.WriteDomainError(err)
		return
	}

	var elevatorName string
	if elevator != nil {
		elevatorName = elevator.Name()
	}

	response := FloorRequestResponse{
		ElevatorName: elevatorName,
		FromFloor:    requestBody.From,
		ToFloor:      requestBody.To,
		Direction:    determineDirection(requestBody.From, requestBody.To),
		Message:      "Floor request processed successfully",
	}

	h.logger.InfoContext(r.Context(), "floor request processed successfully",
		slog.String("elevator_name", elevatorName),
		slog.Int("from_floor", requestBody.From),
		slog.Int("to_floor", requestBody.To),
		slog.String("direction", response.Direction),
		slog.String("request_id", requestID),
		slog.String("component", constants.ComponentHTTPHandler))

	rw.WriteJSON(http.StatusOK, response)
}

// ElevatorCreateHandler handles v1 elevator creation (POST /v1/elevators)
func (h *V1Handlers) ElevatorCreateHandler(w http.ResponseWriter, r *http.Request) {
	requestID := logging.GetRequestID(r.Context())
	rw := NewResponseWriter(w, h.logger, requestID)

	if r.Method != http.MethodPost {
		h.logger.WarnContext(r.Context(), "invalid request method for elevator endpoint",
			slog.String("method", r.Method),
			slog.String("expected", "POST"),
			slog.String("request_id", requestID))
		rw.WriteError(http.StatusMethodNotAllowed, ErrorCodeMethodNotAllowed,
			"Method not allowed", "Only POST method is supported")
		return
	}

	var requestBody ElevatorRequestBody
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestBody); err != nil {
		h.logger.ErrorContext(r.Context(), "failed to decode elevator request",
			slog.String("error", err.Error()),
			slog.String("request_id", requestID))
		rw.WriteError(http.StatusBadRequest, ErrorCodeInvalidJSON,
			"Invalid JSON", "Request body contains invalid JSON")
		return
	}

	// Validate client input floors for elevator creation
	if _, err := domain.NewFloorWithValidation(requestBody.MinFloor); err != nil {
		h.logger.ErrorContext(r.Context(), "invalid min floor in elevator creation request",
			slog.Int("min_floor", requestBody.MinFloor),
			slog.String("error", err.Error()),
			slog.String("request_id", requestID))
		rw.WriteDomainError(err)
		return
	}

	if _, err := domain.NewFloorWithValidation(requestBody.MaxFloor); err != nil {
		h.logger.ErrorContext(r.Context(), "invalid max floor in elevator creation request",
			slog.Int("max_floor", requestBody.MaxFloor),
			slog.String("error", err.Error()),
			slog.String("request_id", requestID))
		rw.WriteDomainError(err)
		return
	}

	// Set default overload threshold if not provided
	overloadThreshold := h.cfg.DefaultOverloadThreshold // use configured default value
	if requestBody.OverloadThreshold != nil {
		overloadThreshold = *requestBody.OverloadThreshold
	}

	err := h.manager.AddElevator(r.Context(), h.cfg, requestBody.Name, requestBody.MinFloor, requestBody.MaxFloor, h.cfg.EachFloorDuration, h.cfg.OpenDoorDuration, overloadThreshold)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to create elevator",
			slog.String("elevator_name", requestBody.Name),
			slog.Int("min_floor", requestBody.MinFloor),
			slog.Int("max_floor", requestBody.MaxFloor),
			slog.String("error", err.Error()),
			slog.String("request_id", requestID))
		rw.WriteDomainError(err)
		return
	}

	response := ElevatorCreateResponse{
		Name:     requestBody.Name,
		MinFloor: requestBody.MinFloor,
		MaxFloor: requestBody.MaxFloor,
		Message:  "Elevator created successfully",
	}

	h.logger.InfoContext(r.Context(), "elevator created successfully",
		slog.String("elevator_name", requestBody.Name),
		slog.Int("min_floor", requestBody.MinFloor),
		slog.Int("max_floor", requestBody.MaxFloor),
		slog.String("request_id", requestID),
		slog.String("component", constants.ComponentHTTPHandler))

	rw.WriteJSON(http.StatusCreated, response)
}

// ElevatorDeleteHandler handles v1 elevator deletion (DELETE /v1/elevators)
func (h *V1Handlers) ElevatorDeleteHandler(w http.ResponseWriter, r *http.Request) {
	requestID := logging.GetRequestID(r.Context())
	rw := NewResponseWriter(w, h.logger, requestID)

	if r.Method != http.MethodDelete {
		h.logger.WarnContext(r.Context(), "invalid request method for elevator delete endpoint",
			slog.String("method", r.Method),
			slog.String("expected", "DELETE"),
			slog.String("request_id", requestID))
		rw.WriteError(http.StatusMethodNotAllowed, ErrorCodeMethodNotAllowed,
			"Method not allowed", "Only DELETE method is supported")
		return
	}

	var requestBody ElevatorDeleteRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestBody); err != nil {
		h.logger.ErrorContext(r.Context(), "failed to decode elevator delete request",
			slog.String("error", err.Error()),
			slog.String("request_id", requestID))
		rw.WriteError(http.StatusBadRequest, ErrorCodeInvalidJSON,
			"Invalid JSON", "Request body contains invalid JSON")
		return
	}

	// Validate elevator name
	if requestBody.Name == "" {
		h.logger.ErrorContext(r.Context(), "elevator name is required",
			slog.String("request_id", requestID))
		rw.WriteError(http.StatusBadRequest, ErrorCodeValidation,
			"Validation Failed", "Elevator name is required")
		return
	}

	err := h.manager.DeleteElevator(r.Context(), requestBody.Name)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to delete elevator",
			slog.String("elevator_name", requestBody.Name),
			slog.String("error", err.Error()),
			slog.String("request_id", requestID))
		rw.WriteDomainError(err)
		return
	}

	response := ElevatorDeleteResponse{
		Name:    requestBody.Name,
		Message: "Elevator deleted successfully",
	}

	h.logger.InfoContext(r.Context(), "elevator deleted successfully",
		slog.String("elevator_name", requestBody.Name),
		slog.String("request_id", requestID),
		slog.String("component", constants.ComponentHTTPHandler))

	rw.WriteJSON(http.StatusOK, response)
}

// HealthHandler handles v1 health checks (GET /v1/health)
func (h *V1Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	requestID := logging.GetRequestID(r.Context())
	rw := NewResponseWriter(w, h.logger, requestID)

	if r.Method != http.MethodGet {
		rw.WriteError(http.StatusMethodNotAllowed, ErrorCodeMethodNotAllowed,
			"Method not allowed", "Only GET method is supported")
		return
	}

	health, err := h.manager.GetHealthStatus()
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get health status",
			slog.String("error", err.Error()),
			slog.String("request_id", requestID))
		rw.WriteError(http.StatusInternalServerError, ErrorCodeInternal,
			"Health check failed", err.Error())
		return
	}

	// Determine overall health status
	status := "healthy"
	statusCode := http.StatusOK
	if systemHealthy, ok := health["system_healthy"].(bool); ok && !systemHealthy {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Checks:    health,
	}

	h.logger.InfoContext(r.Context(), "health check request processed",
		slog.Int("status_code", statusCode),
		slog.String("request_id", requestID))

	rw.WriteJSON(statusCode, response)
}

// MetricsHandler handles v1 system metrics (GET /v1/metrics)
func (h *V1Handlers) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	requestID := logging.GetRequestID(r.Context())
	rw := NewResponseWriter(w, h.logger, requestID)

	if r.Method != http.MethodGet {
		rw.WriteError(http.StatusMethodNotAllowed, ErrorCodeMethodNotAllowed,
			"Method not allowed", "Only GET method is supported")
		return
	}

	metrics := h.manager.GetMetrics()

	response := MetricsResponse{
		Timestamp: time.Now(),
		Metrics:   metrics,
	}

	h.logger.InfoContext(r.Context(), "metrics request processed",
		slog.String("request_id", requestID))

	rw.WriteJSON(http.StatusOK, response)
}

// APIInfoHandler provides information about available API endpoints (GET /v1)
func (h *V1Handlers) APIInfoHandler(w http.ResponseWriter, r *http.Request) {
	requestID := logging.GetRequestID(r.Context())
	rw := NewResponseWriter(w, h.logger, requestID)

	if r.Method != http.MethodGet {
		rw.WriteError(http.StatusMethodNotAllowed, ErrorCodeMethodNotAllowed,
			"Method not allowed", "Only GET method is supported")
		return
	}

	response := APIInfoResponse{
		Name:        "Elevator Control System API",
		Version:     "v1",
		Description: "RESTful API for managing elevator systems",
		Endpoints: map[string]string{
			"POST /v1/floors/request": "Request elevator from one floor to another",
			"POST /v1/elevators":      "Create a new elevator in the system",
			"DELETE /v1/elevators":    "Delete an elevator from the system",
			"GET /v1/health":          "Check system health status",
			"GET /v1/metrics":         "Get system metrics",
			"GET /v1":                 "Get API information",
			"GET /metrics":            "Prometheus metrics endpoint",
			"WebSocket /ws/status":    "Real-time elevator status updates",
		},
	}

	rw.WriteJSON(http.StatusOK, response)
}
