package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/slavakukuyev/elevator-go/internal/factory"
	"github.com/slavakukuyev/elevator-go/internal/infra/config"
	"github.com/slavakukuyev/elevator-go/internal/manager"
)

func buildServerTestConfig() *config.Config {
	return &config.Config{
		LogLevel:                       "INFO",
		Port:                           8080,
		MinFloor:                       -5,
		MaxFloor:                       20,
		EachFloorDuration:              time.Millisecond * 50, // Fast for testing
		OpenDoorDuration:               time.Millisecond * 50, // Fast for testing
		OperationTimeout:               time.Second * 5,       // Adequate time for operations
		CreateElevatorTimeout:          time.Second * 2,       // Adequate time for creation
		RequestTimeout:                 time.Second * 2,       // Adequate time for requests
		StatusUpdateTimeout:            time.Second * 1,       // Fast status updates
		HealthCheckTimeout:             time.Second * 1,       // Fast health checks
		CircuitBreakerEnabled:          true,
		CircuitBreakerMaxFailures:      5,
		CircuitBreakerResetTimeout:     time.Second * 30,
		CircuitBreakerFailureThreshold: 0.6,
	}
}

func setupTestServer() (*Server, *manager.Manager) {
	cfg := buildServerTestConfig()
	factory := &factory.StandardElevatorFactory{}
	manager := manager.New(cfg, factory)
	server := NewServer(cfg, 8080, manager)
	return server, manager
}

func TestElevatorHandler_Comprehensive(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    interface{}
		expectedStatus int
		setupElevators []string // Names of elevators to create first
		expectError    bool
		errorContains  string
	}{
		{
			name:   "valid elevator creation",
			method: "POST",
			requestBody: ElevatorRequestBody{
				Name:     "TestElevator1",
				MinFloor: 0,
				MaxFloor: 10,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "basement elevator creation",
			method: "POST",
			requestBody: ElevatorRequestBody{
				Name:     "BasementElevator",
				MinFloor: -5,
				MaxFloor: 0,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "high-rise elevator creation",
			method: "POST",
			requestBody: ElevatorRequestBody{
				Name:     "HighRiseElevator",
				MinFloor: 0,
				MaxFloor: 100,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "duplicate elevator name should fail",
			method: "POST",
			requestBody: ElevatorRequestBody{
				Name:     "DuplicateElevator",
				MinFloor: 0,
				MaxFloor: 10,
			},
			setupElevators: []string{"DuplicateElevator"},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "invalid HTTP method",
			method:         "GET",
			requestBody:    ElevatorRequestBody{Name: "TestElevator", MinFloor: 0, MaxFloor: 10},
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    true,
		},
		{
			name:           "invalid JSON body",
			method:         "POST",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:   "empty elevator name",
			method: "POST",
			requestBody: ElevatorRequestBody{
				Name:     "",
				MinFloor: 0,
				MaxFloor: 10,
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:   "same min and max floor",
			method: "POST",
			requestBody: ElevatorRequestBody{
				Name:     "SameFloorElevator",
				MinFloor: 5,
				MaxFloor: 5,
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, elevatorManager := setupTestServer()
			ctx := context.Background()

			// Setup existing elevators if specified
			for _, elevatorName := range tt.setupElevators {
				err := elevatorManager.AddElevator(ctx, buildServerTestConfig(), elevatorName, 0, 10, time.Millisecond*50, time.Millisecond*50, 12)
				require.NoError(t, err)
			}

			// Prepare request body
			var requestBodyBytes []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				requestBodyBytes = []byte(str)
			} else {
				requestBodyBytes, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			// Create request
			req, err := http.NewRequest(tt.method, "/elevator", bytes.NewBuffer(requestBodyBytes))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(server.elevatorHandler)
			handler.ServeHTTP(rr, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectError && tt.expectedStatus == http.StatusOK {
				responseBody := rr.Body.String()
				assert.Contains(t, responseBody, "created successfully")
			}
		})
	}
}

// TestFloorHandler_Comprehensive tests all aspects of the floor request handler
// including valid requests, validation errors, HTTP method validation, and edge cases
func TestFloorHandler_Comprehensive(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    interface{}
		expectedStatus int
		expectError    bool
		elevatorSetup  bool // Whether to setup elevators
	}{
		// Valid request test cases - these should succeed with 200 OK
		{
			name:   "valid up request",
			method: "POST",
			requestBody: FloorRequestBody{
				From: 2,
				To:   8,
			},
			expectedStatus: http.StatusOK,
			elevatorSetup:  true,
		},
		{
			name:   "valid down request",
			method: "POST",
			requestBody: FloorRequestBody{
				From: 15,
				To:   5,
			},
			expectedStatus: http.StatusOK,
			elevatorSetup:  true,
		},
		{
			name:   "basement request", // Test negative floor handling
			method: "POST",
			requestBody: FloorRequestBody{
				From: -3,
				To:   0,
			},
			expectedStatus: http.StatusOK,
			elevatorSetup:  true,
		},
		{
			name:   "boundary floor request", // Test elevator's min/max floor boundaries
			method: "POST",
			requestBody: FloorRequestBody{
				From: -5,
				To:   20,
			},
			expectedStatus: http.StatusOK,
			elevatorSetup:  true,
		},

		// Validation error test cases - these should return 400 Bad Request
		{
			name:   "same floor request should fail", // Cannot request same from/to floor
			method: "POST",
			requestBody: FloorRequestBody{
				From: 5,
				To:   5,
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			elevatorSetup:  true,
		},
		{
			name:   "floors out of range", // Floors outside elevator's range should return 400
			method: "POST",
			requestBody: FloorRequestBody{
				From: 25, // Above max floor of 20
				To:   30, // Above max floor of 20
			},
			expectedStatus: http.StatusBadRequest, // Fixed: was expecting 404, should be 400 for validation error
			expectError:    true,
			elevatorSetup:  true,
		},
		{
			name:   "negative floor validation", // Floor outside global valid range
			method: "POST",
			requestBody: FloorRequestBody{
				From: -150, // Outside valid range (-100 to 200)
				To:   0,
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			elevatorSetup:  true,
		},

		// HTTP method validation test cases - these should return 405 Method Not Allowed
		{
			name:           "invalid HTTP method", // Only POST is allowed for floor requests
			method:         "GET",
			requestBody:    FloorRequestBody{From: 2, To: 8},
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    true,
			elevatorSetup:  true,
		},

		// Request format error test cases - these should return 400 Bad Request
		{
			name:           "invalid JSON body", // Malformed JSON should return 400
			method:         "POST",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			elevatorSetup:  true,
		},

		// System state error test cases - these should return 500 Internal Server Error
		{
			name:   "no elevators available", // System error when no elevators exist
			method: "POST",
			requestBody: FloorRequestBody{
				From: 2,
				To:   8,
			},
			expectedStatus: http.StatusInternalServerError, // Fixed: was expecting 404, should be 500 for system error
			expectError:    true,
			elevatorSetup:  false, // Don't setup any elevators for this test
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, elevatorManager := setupTestServer()

			if tt.elevatorSetup {
				// Setup test elevators
				ctx := context.Background()
				err := elevatorManager.AddElevator(ctx, buildServerTestConfig(), "TestElevator", -5, 20, time.Millisecond*50, time.Millisecond*50, 12)
				require.NoError(t, err)
			}

			// Prepare request body
			var requestBodyBytes []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				requestBodyBytes = []byte(str)
			} else {
				requestBodyBytes, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			// Create request
			req, err := http.NewRequest(tt.method, "/floor", bytes.NewBuffer(requestBodyBytes))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(server.floorHandler)
			handler.ServeHTTP(rr, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectError && tt.expectedStatus == http.StatusOK {
				responseBody := rr.Body.String()
				assert.Contains(t, responseBody, "received request")
			}
		})
	}
}

func TestServer_NewServer(t *testing.T) {
	cfg := buildServerTestConfig()
	factory := &factory.StandardElevatorFactory{}
	manager := manager.New(cfg, factory)

	server := NewServer(cfg, 8080, manager)

	assert.NotNil(t, server)
	assert.Equal(t, manager, server.manager)
	assert.Equal(t, cfg, server.cfg)
	assert.NotNil(t, server.httpServer)
	assert.NotNil(t, server.logger)
}

func TestServer_ConcurrentRequests(t *testing.T) {
	server, elevatorManager := setupTestServer()
	ctx := context.Background()

	// Setup multiple elevators
	for i := 0; i < 3; i++ {
		elevatorName := fmt.Sprintf("ConcurrentTestElevator%d", i)
		err := elevatorManager.AddElevator(ctx, buildServerTestConfig(), elevatorName, 0, 20, time.Millisecond*50, time.Millisecond*50, 12)
		require.NoError(t, err)
	}

	const numRequests = 20
	done := make(chan bool, numRequests)
	successCount := 0

	// Launch concurrent floor requests
	for i := 0; i < numRequests; i++ {
		go func(requestID int) {
			from := requestID % 15
			to := from + 3
			if to > 20 {
				to = 20
			}

			floorRequest := FloorRequestBody{From: from, To: to}
			requestBody, _ := json.Marshal(floorRequest)

			req, _ := http.NewRequest("POST", "/floor", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(server.floorHandler)
			handler.ServeHTTP(rr, req)

			done <- rr.Code == http.StatusOK
		}(i)
	}

	// Wait for all requests and count successes
	for i := 0; i < numRequests; i++ {
		if <-done {
			successCount++
		}
	}

	// Should have many successful requests
	assert.Greater(t, successCount, numRequests/2, "Should handle concurrent requests successfully")
}

func TestServer_ErrorHandling(t *testing.T) {
	server, _ := setupTestServer()

	t.Run("malformed JSON in elevator request", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/elevator", bytes.NewBuffer([]byte("{invalid json")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(server.elevatorHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("malformed JSON in floor request", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/floor", bytes.NewBuffer([]byte("{invalid json")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(server.floorHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("missing content type", func(t *testing.T) {
		validRequest := ElevatorRequestBody{Name: "TestElevator", MinFloor: 0, MaxFloor: 10}
		requestBody, _ := json.Marshal(validRequest)

		req, err := http.NewRequest("POST", "/elevator", bytes.NewBuffer(requestBody))
		require.NoError(t, err)
		// Not setting Content-Type header

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(server.elevatorHandler)
		handler.ServeHTTP(rr, req)

		// Should still work as Go's JSON decoder is lenient
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestServer_EdgeCases(t *testing.T) {
	t.Run("maximum floor range elevator", func(t *testing.T) {
		server, _ := setupTestServer()

		elevatorRequest := ElevatorRequestBody{
			Name:     "MaxRangeElevator",
			MinFloor: -100, // Within valid range
			MaxFloor: 200,  // Within valid range
		}
		requestBody, err := json.Marshal(elevatorRequest)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/elevator", bytes.NewBuffer(requestBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(server.elevatorHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("single floor elevator", func(t *testing.T) {
		server, _ := setupTestServer()

		elevatorRequest := ElevatorRequestBody{
			Name:     "SingleFloorElevator",
			MinFloor: 10,
			MaxFloor: 11,
		}
		requestBody, err := json.Marshal(elevatorRequest)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/elevator", bytes.NewBuffer(requestBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(server.elevatorHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("floor request with existing elevator", func(t *testing.T) {
		server, elevatorManager := setupTestServer()
		ctx := context.Background()

		// Add an elevator first
		err := elevatorManager.AddElevator(ctx, buildServerTestConfig(), "ExistingElevator", 0, 10, time.Millisecond*50, time.Millisecond*50, 12)
		require.NoError(t, err)

		// Make multiple requests to the same elevator
		for i := 0; i < 5; i++ {
			floorRequest := FloorRequestBody{From: 1, To: 5}
			requestBody, err := json.Marshal(floorRequest)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", "/floor", bytes.NewBuffer(requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(server.floorHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
		}
	})
}
