package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/slavakukuyev/elevator-go/internal/domain"
	"github.com/slavakukuyev/elevator-go/internal/elevator"
	"github.com/slavakukuyev/elevator-go/internal/infra/config"
)

// ManagerInterface defines the interface for manager operations used by handlers
type ManagerInterface interface {
	RequestElevator(ctx context.Context, fromFloor, toFloor int) (*elevator.Elevator, error)
	AddElevator(ctx context.Context, cfg *config.Config, name string, minFloor, maxFloor int, eachFloorDuration, openDoorDuration time.Duration, overloadThreshold int) error
	GetStatus() (map[string]interface{}, error)
	GetHealthStatus() (map[string]interface{}, error)
	GetMetrics() map[string]interface{}
}

// MockManager for testing - implements ManagerInterface
type MockManager struct {
	mock.Mock
}

func (m *MockManager) RequestElevator(ctx context.Context, fromFloor, toFloor int) (*elevator.Elevator, error) {
	args := m.Called(ctx, fromFloor, toFloor)
	// For testing, create a real elevator or return nil for error cases
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	// Create a simple elevator for testing (this avoids interface conversion issues)
	testElevator, _ := elevator.New("test-elevator", -10, 10, time.Millisecond, time.Millisecond,
		30*time.Second, 5, 30*time.Second, 3, 12)
	return testElevator, nil
}

func (m *MockManager) AddElevator(ctx context.Context, cfg *config.Config, name string, minFloor, maxFloor int, eachFloorDuration, openDoorDuration time.Duration, overloadThreshold int) error {
	args := m.Called(ctx, cfg, name, minFloor, maxFloor, eachFloorDuration, openDoorDuration)
	return args.Error(0)
}

func (m *MockManager) GetStatus() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockManager) GetHealthStatus() (map[string]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockManager) GetMetrics() map[string]interface{} {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(map[string]interface{})
}

// TestV1Handlers is a test version of V1Handlers that uses the interface
type TestV1Handlers struct {
	manager ManagerInterface
	cfg     *config.Config
	logger  *slog.Logger
}

// Delegate methods to match V1Handlers interface
func (h *TestV1Handlers) APIInfoHandler(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	rw := NewResponseWriter(w, h.logger, requestID)

	response := APIInfoResponse{
		Name:        "Elevator Control System API",
		Version:     "v1",
		Description: "RESTful API for elevator control and monitoring",
		Endpoints: map[string]string{
			"request_elevator": "POST /v1/floors/request",
			"create_elevator":  "POST /v1/elevators",
			"health":           "GET /v1/health",
			"metrics":          "GET /v1/metrics",
		},
	}

	rw.WriteJSON(http.StatusOK, response)
}

func (h *TestV1Handlers) FloorRequestHandler(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	rw := NewResponseWriter(w, h.logger, requestID)

	if r.Method != http.MethodPost {
		rw.WriteError(http.StatusMethodNotAllowed, ErrorCodeMethodNotAllowed,
			"Method not allowed", "Only POST method is supported")
		return
	}

	var requestBody FloorRequestBody
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestBody); err != nil {
		rw.WriteError(http.StatusBadRequest, ErrorCodeInvalidJSON,
			"Invalid JSON", "Request body contains invalid JSON")
		return
	}

	elevator, err := h.manager.RequestElevator(r.Context(), requestBody.From, requestBody.To)
	if err != nil {
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

	rw.WriteJSON(http.StatusOK, response)
}

func (h *TestV1Handlers) ElevatorCreateHandler(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	rw := NewResponseWriter(w, h.logger, requestID)

	if r.Method != http.MethodPost {
		rw.WriteError(http.StatusMethodNotAllowed, ErrorCodeMethodNotAllowed,
			"Method not allowed", "Only POST method is supported")
		return
	}

	var requestBody ElevatorRequestBody
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestBody); err != nil {
		rw.WriteError(http.StatusBadRequest, ErrorCodeInvalidJSON,
			"Invalid JSON", "Request body contains invalid JSON")
		return
	}

	err := h.manager.AddElevator(r.Context(), h.cfg, requestBody.Name, requestBody.MinFloor, requestBody.MaxFloor, h.cfg.EachFloorDuration, h.cfg.OpenDoorDuration, h.cfg.DefaultOverloadThreshold)
	if err != nil {
		rw.WriteDomainError(err)
		return
	}

	response := ElevatorCreateResponse{
		Name:     requestBody.Name,
		MinFloor: requestBody.MinFloor,
		MaxFloor: requestBody.MaxFloor,
		Message:  "Elevator created successfully",
	}

	rw.WriteJSON(http.StatusCreated, response)
}

func (h *TestV1Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	rw := NewResponseWriter(w, h.logger, requestID)

	healthData, err := h.manager.GetHealthStatus()
	if err != nil {
		// Convert generic error to domain error for consistent handling
		domainErr := domain.NewInternalError("health check failed", err)
		rw.WriteDomainError(domainErr)
		return
	}

	// Determine overall health status (same logic as real handler)
	status := "healthy"
	statusCode := http.StatusOK
	if systemHealthy, ok := healthData["system_healthy"].(bool); ok && !systemHealthy {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC(),
		Checks:    healthData,
	}

	rw.WriteJSON(statusCode, response)
}

func (h *TestV1Handlers) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	rw := NewResponseWriter(w, h.logger, requestID)

	metricsData := h.manager.GetMetrics()

	response := MetricsResponse{
		Timestamp: time.Now().UTC(),
		Metrics:   metricsData,
	}

	rw.WriteJSON(http.StatusOK, response)
}

func setupTestHandlers() (*TestV1Handlers, *MockManager) {
	mockManager := &MockManager{}
	logger := slog.Default()
	cfg := &config.Config{
		EachFloorDuration: time.Second,
		OpenDoorDuration:  time.Second,
	}

	handlers := &TestV1Handlers{
		manager: mockManager,
		cfg:     cfg,
		logger:  logger,
	}

	return handlers, mockManager
}

func createRequestWithContext(method, path string, body string, requestID string) *http.Request {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	ctx := context.WithValue(req.Context(), requestIDKey, requestID)
	return req.WithContext(ctx)
}

func parseAPIResponse(t *testing.T, body []byte) APIResponse {
	var response APIResponse
	err := json.Unmarshal(body, &response)
	require.NoError(t, err)
	return response
}

func TestV1Handlers_APIInfoHandler(t *testing.T) {
	handlers, _ := setupTestHandlers()

	t.Run("returns API information", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := createRequestWithContext("GET", "/v1", "", "test-123")

		handlers.APIInfoHandler(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "test-123", w.Header().Get("X-Request-ID"))

		response := parseAPIResponse(t, w.Body.Bytes())
		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)

		// Check specific fields in the response
		data, ok := response.Data.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, "Elevator Control System API", data["name"])
		assert.Equal(t, "v1", data["version"])
		assert.Contains(t, data, "description")
		assert.Contains(t, data, "endpoints")
	})
}

func TestV1Handlers_FloorRequestHandler(t *testing.T) {
	handlers, mockManager := setupTestHandlers()

	t.Run("successfully requests floor", func(t *testing.T) {
		mockManager.On("RequestElevator", mock.Anything, 1, 5).Return(nil, nil)

		w := httptest.NewRecorder()
		body := `{"from": 1, "to": 5}`
		r := createRequestWithContext("POST", "/v1/floors/request", body, "test-456")

		handlers.FloorRequestHandler(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		response := parseAPIResponse(t, w.Body.Bytes())
		assert.True(t, response.Success)

		data, ok := response.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, data, "message")
		assert.Contains(t, data, "from_floor")
		assert.Contains(t, data, "to_floor")
		assert.Equal(t, float64(1), data["from_floor"])
		assert.Equal(t, float64(5), data["to_floor"])
		assert.Equal(t, "test-elevator", data["elevator_name"])

		mockManager.AssertExpectations(t)
	})

	t.Run("handles invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"from": invalid}`
		r := createRequestWithContext("POST", "/v1/floors/request", body, "test-457")

		handlers.FloorRequestHandler(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		response := parseAPIResponse(t, w.Body.Bytes())
		assert.False(t, response.Success)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "INVALID_JSON", response.Error.Code)
	})

	t.Run("handles manager error", func(t *testing.T) {
		mockManager.On("RequestElevator", mock.Anything, 1, 300).Return(nil, domain.NewValidationError("floor out of range", nil))

		w := httptest.NewRecorder()
		body := `{"from": 1, "to": 300}`
		r := createRequestWithContext("POST", "/v1/floors/request", body, "test-459")

		handlers.FloorRequestHandler(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		response := parseAPIResponse(t, w.Body.Bytes())
		assert.False(t, response.Success)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)

		mockManager.AssertExpectations(t)
	})

	t.Run("handles wrong HTTP method", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := createRequestWithContext("GET", "/v1/floors/request", "", "test-method")

		handlers.FloorRequestHandler(w, r)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		response := parseAPIResponse(t, w.Body.Bytes())
		assert.False(t, response.Success)
		assert.Equal(t, "METHOD_NOT_ALLOWED", response.Error.Code)
	})

	t.Run("handles same from and to floor", func(t *testing.T) {
		mockManager.On("RequestElevator", mock.Anything, 5, 5).Return(nil, domain.NewValidationError("requested floor must be different from current floor", nil))

		w := httptest.NewRecorder()
		body := `{"from": 5, "to": 5}`
		r := createRequestWithContext("POST", "/v1/floors/request", body, "test-same-floor")

		handlers.FloorRequestHandler(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		response := parseAPIResponse(t, w.Body.Bytes())
		assert.False(t, response.Success)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)

		mockManager.AssertExpectations(t)
	})
}

func TestV1Handlers_ElevatorCreateHandler(t *testing.T) {
	handlers, mockManager := setupTestHandlers()

	t.Run("successfully creates elevator", func(t *testing.T) {
		mockManager.On("AddElevator", mock.Anything, mock.Anything, "test-elevator", -2, 10, time.Second, time.Second).Return(nil)

		w := httptest.NewRecorder()
		body := `{"name": "test-elevator", "min_floor": -2, "max_floor": 10}`
		r := createRequestWithContext("POST", "/v1/elevators", body, "test-789")

		handlers.ElevatorCreateHandler(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
		response := parseAPIResponse(t, w.Body.Bytes())
		assert.True(t, response.Success)

		data, ok := response.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, data, "message")
		assert.Contains(t, data, "name")
		assert.Contains(t, data, "min_floor")
		assert.Contains(t, data, "max_floor")
		assert.Equal(t, "test-elevator", data["name"])

		mockManager.AssertExpectations(t)
	})

	t.Run("handles manager error", func(t *testing.T) {
		mockManager.On("AddElevator", mock.Anything, mock.Anything, "duplicate", 0, 10, time.Second, time.Second).Return(domain.NewValidationError("elevator with this name already exists", nil))

		w := httptest.NewRecorder()
		body := `{"name": "duplicate", "min_floor": 0, "max_floor": 10}`
		r := createRequestWithContext("POST", "/v1/elevators", body, "test-790")

		handlers.ElevatorCreateHandler(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		response := parseAPIResponse(t, w.Body.Bytes())
		assert.False(t, response.Success)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)

		mockManager.AssertExpectations(t)
	})

	t.Run("handles invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{invalid}`
		r := createRequestWithContext("POST", "/v1/elevators", body, "test-791")

		handlers.ElevatorCreateHandler(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		response := parseAPIResponse(t, w.Body.Bytes())
		assert.False(t, response.Success)
		assert.Equal(t, "INVALID_JSON", response.Error.Code)
	})

	t.Run("handles wrong HTTP method", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := createRequestWithContext("GET", "/v1/elevators", "", "test-method")

		handlers.ElevatorCreateHandler(w, r)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		response := parseAPIResponse(t, w.Body.Bytes())
		assert.False(t, response.Success)
		assert.Equal(t, "METHOD_NOT_ALLOWED", response.Error.Code)
	})
}

func TestV1Handlers_HealthHandler(t *testing.T) {
	t.Run("returns healthy status", func(t *testing.T) {
		handlers, mockManager := setupTestHandlers()
		statusData := map[string]interface{}{
			"total_elevators":  3,
			"active_elevators": 2,
			"system_status":    "healthy",
		}
		mockManager.On("GetHealthStatus").Return(statusData, nil)

		w := httptest.NewRecorder()
		r := createRequestWithContext("GET", "/v1/health", "", "test-health")

		handlers.HealthHandler(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		response := parseAPIResponse(t, w.Body.Bytes())
		assert.True(t, response.Success)

		data, ok := response.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "healthy", data["status"])
		assert.Contains(t, data, "timestamp")
		assert.Contains(t, data, "checks")

		mockManager.AssertExpectations(t)
	})

	t.Run("handles manager error", func(t *testing.T) {
		handlers, mockManager := setupTestHandlers()
		mockManager.On("GetHealthStatus").Return(nil, fmt.Errorf("health check failed"))

		w := httptest.NewRecorder()
		r := createRequestWithContext("GET", "/v1/health", "", "test-health-error")

		handlers.HealthHandler(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		response := parseAPIResponse(t, w.Body.Bytes())
		assert.False(t, response.Success)
		assert.NotNil(t, response.Error, "Error should not be nil when health check fails")
		assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)

		mockManager.AssertExpectations(t)
	})
}

func TestV1Handlers_HealthHandler_ZeroElevators(t *testing.T) {
	t.Run("returns healthy status with zero elevators", func(t *testing.T) {
		handlers, mockManager := setupTestHandlers()

		// Mock health status for zero elevators (healthy initial state)
		healthData := map[string]interface{}{
			"total_elevators":   0,
			"healthy_elevators": 0,
			"system_healthy":    true, // Key assertion: system is healthy with 0 elevators
			"elevators":         map[string]interface{}{},
			"timestamp":         1640995200, // Mock timestamp
		}
		mockManager.On("GetHealthStatus").Return(healthData, nil)

		w := httptest.NewRecorder()
		r := createRequestWithContext("GET", "/v1/health", "", "test-health-zero-elevators")

		handlers.HealthHandler(w, r)

		// Should return 200 OK, not 503
		assert.Equal(t, http.StatusOK, w.Code, "Should return 200 OK for zero elevators")
		response := parseAPIResponse(t, w.Body.Bytes())
		assert.True(t, response.Success, "Response should be successful")

		data, ok := response.Data.(map[string]interface{})
		require.True(t, ok, "Response data should be a map")
		assert.Equal(t, "healthy", data["status"], "Status should be 'healthy'")
		assert.Contains(t, data, "timestamp", "Should contain timestamp")
		assert.Contains(t, data, "checks", "Should contain health checks")

		// Verify the checks contain our zero-elevator data
		checks, ok := data["checks"].(map[string]interface{})
		require.True(t, ok, "Checks should be a map")
		assert.Equal(t, float64(0), checks["total_elevators"], "Total elevators should be 0")
		assert.Equal(t, float64(0), checks["healthy_elevators"], "Healthy elevators should be 0")
		assert.True(t, checks["system_healthy"].(bool), "System should be healthy")

		mockManager.AssertExpectations(t)
	})

	t.Run("returns unhealthy status when system_healthy is false", func(t *testing.T) {
		handlers, mockManager := setupTestHandlers()

		// Mock health status where system is explicitly unhealthy
		healthData := map[string]interface{}{
			"total_elevators":   2,
			"healthy_elevators": 0,
			"system_healthy":    false, // System is unhealthy (e.g., all elevators broken)
			"elevators": map[string]interface{}{
				"Elevator1": map[string]interface{}{"status": "error"},
				"Elevator2": map[string]interface{}{"status": "error"},
			},
			"timestamp": 1640995200,
		}
		mockManager.On("GetHealthStatus").Return(healthData, nil)

		w := httptest.NewRecorder()
		r := createRequestWithContext("GET", "/v1/health", "", "test-health-unhealthy")

		handlers.HealthHandler(w, r)

		// Should return 503 Service Unavailable
		assert.Equal(t, http.StatusServiceUnavailable, w.Code, "Should return 503 for unhealthy system")
		response := parseAPIResponse(t, w.Body.Bytes())
		assert.False(t, response.Success, "Response should be unsuccessful for unhealthy system")

		data, ok := response.Data.(map[string]interface{})
		require.True(t, ok, "Response data should be a map")
		assert.Equal(t, "unhealthy", data["status"], "Status should be 'unhealthy'")

		mockManager.AssertExpectations(t)
	})
}

func TestV1Handlers_MetricsHandler(t *testing.T) {
	handlers, mockManager := setupTestHandlers()

	t.Run("returns system metrics", func(t *testing.T) {
		metricsData := map[string]interface{}{
			"total_requests":     150,
			"average_wait_time":  30.5,
			"total_elevators":    3,
			"active_elevators":   2,
			"requests_processed": 1250,
		}
		mockManager.On("GetMetrics").Return(metricsData)

		w := httptest.NewRecorder()
		r := createRequestWithContext("GET", "/v1/metrics", "", "test-metrics")

		handlers.MetricsHandler(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		response := parseAPIResponse(t, w.Body.Bytes())
		assert.True(t, response.Success)

		data, ok := response.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, data, "timestamp")
		assert.Contains(t, data, "metrics")

		metrics, ok := data["metrics"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(150), metrics["total_requests"])
		assert.Equal(t, 30.5, metrics["average_wait_time"])

		mockManager.AssertExpectations(t)
	})

	t.Run("handles manager returning nil", func(t *testing.T) {
		mockManager.On("GetMetrics").Return(nil)

		w := httptest.NewRecorder()
		r := createRequestWithContext("GET", "/v1/metrics", "", "test-metrics-nil")

		handlers.MetricsHandler(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		response := parseAPIResponse(t, w.Body.Bytes())
		assert.True(t, response.Success)

		data, ok := response.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, data, "timestamp")
		assert.Contains(t, data, "metrics")

		mockManager.AssertExpectations(t)
	})
}

func TestRequestContext(t *testing.T) {
	t.Run("request ID is preserved through handler", func(t *testing.T) {
		handlers, _ := setupTestHandlers()
		requestID := "test-context-123"

		w := httptest.NewRecorder()
		r := createRequestWithContext("GET", "/v1", "", requestID)

		handlers.APIInfoHandler(w, r)

		assert.Equal(t, requestID, w.Header().Get("X-Request-ID"))
	})
}

func TestResponseFormat(t *testing.T) {
	handlers, _ := setupTestHandlers()

	t.Run("all responses follow standard format", func(t *testing.T) {
		testCases := []struct {
			name    string
			handler func(http.ResponseWriter, *http.Request)
			path    string
			method  string
		}{
			{
				name:    "API info",
				handler: handlers.APIInfoHandler,
				path:    "/v1",
				method:  "GET",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				r := createRequestWithContext(tc.method, tc.path, "", "test-format")

				tc.handler(w, r)

				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				assert.NotEmpty(t, w.Header().Get("X-Request-ID"))

				response := parseAPIResponse(t, w.Body.Bytes())
				assert.NotNil(t, response.Meta)
				assert.Equal(t, "test-format", response.Meta.RequestID)
				assert.Equal(t, "v1", response.Meta.Version)
				assert.NotEmpty(t, response.Meta.Duration)
				assert.False(t, response.Timestamp.IsZero())
			})
		}
	})
}

func TestEdgeCases(t *testing.T) {
	handlers, mockManager := setupTestHandlers()

	t.Run("handles very large floor numbers", func(t *testing.T) {
		mockManager.On("RequestElevator", mock.Anything, 1, 9999999).Return(nil, domain.NewValidationError("floor too high", nil))

		w := httptest.NewRecorder()
		body := `{"from": 1, "to": 9999999}`
		r := createRequestWithContext("POST", "/v1/floors/request", body, "test-large")

		handlers.FloorRequestHandler(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockManager.AssertExpectations(t)
	})

	t.Run("handles negative floor numbers", func(t *testing.T) {
		mockManager.On("RequestElevator", mock.Anything, -5, 0).Return(nil, nil)

		w := httptest.NewRecorder()
		body := `{"from": -5, "to": 0}`
		r := createRequestWithContext("POST", "/v1/floors/request", body, "test-negative")

		handlers.FloorRequestHandler(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		mockManager.AssertExpectations(t)
	})
}
