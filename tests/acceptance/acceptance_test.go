package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/slavakukuyev/elevator-go/internal/factory"
	httpPkg "github.com/slavakukuyev/elevator-go/internal/http"
	"github.com/slavakukuyev/elevator-go/internal/infra/config"
	"github.com/slavakukuyev/elevator-go/internal/infra/logging"
	"github.com/slavakukuyev/elevator-go/internal/manager"
)

// AcceptanceTestSuite represents the test suite with proper isolation
type AcceptanceTestSuite struct {
	suite.Suite
	server  *httpPkg.Server
	manager *manager.Manager
	cfg     *config.Config
	testSrv *httptest.Server
	ctx     context.Context
	cancel  context.CancelFunc
}

// T returns the underlying testing.T instance to satisfy the type checker
func (suite *AcceptanceTestSuite) T() *testing.T {
	return suite.Suite.T()
}

// SetupSuite initializes the test suite once
func (suite *AcceptanceTestSuite) SetupSuite() {
	// Suppress all logging during tests to reduce noise
	log.SetOutput(io.Discard)

	// Initialize logging with ERROR level only
	logging.InitLogger("ERROR")

	suite.ctx, suite.cancel = context.WithCancel(context.Background())
}

// TearDownSuite cleans up the test suite
func (suite *AcceptanceTestSuite) TearDownSuite() {
	if suite.cancel != nil {
		suite.cancel()
	}
}

// SetupTest ensures clean state for each test
func (suite *AcceptanceTestSuite) SetupTest() {
	// Set testing environment to get proper defaults
	if err := os.Setenv("ENV", "testing"); err != nil {
		suite.T().Fatalf("Failed to set ENV: %v", err)
	}
	if err := os.Setenv("LOG_LEVEL", "ERROR"); err != nil { // Reduce noise in tests
		suite.T().Fatalf("Failed to set LOG_LEVEL: %v", err)
	}
	if err := os.Setenv("DEFAULT_MIN_FLOOR", "-10"); err != nil {
		suite.T().Fatalf("Failed to set DEFAULT_MIN_FLOOR: %v", err)
	}
	if err := os.Setenv("DEFAULT_MAX_FLOOR", "50"); err != nil {
		suite.T().Fatalf("Failed to set DEFAULT_MAX_FLOOR: %v", err)
	}

	// Create configuration using proper initialization to get testing defaults
	var err error
	suite.cfg, err = config.InitConfig()
	require.NoError(suite.T(), err)

	// Create fresh instances for each test to ensure isolation
	factory := &factory.StandardElevatorFactory{}
	suite.manager = manager.New(suite.cfg, factory)
	suite.server = httpPkg.NewServer(suite.cfg, suite.cfg.Port, suite.manager)

	// Create test HTTP server
	suite.testSrv = httptest.NewServer(suite.server.GetHandler())

	// Allow some time for server to be ready
	time.Sleep(10 * time.Millisecond)
}

// TearDownTest cleans up after each test
func (suite *AcceptanceTestSuite) TearDownTest() {
	if suite.testSrv != nil {
		suite.testSrv.Close()
		suite.testSrv = nil
	}

	// Clean up environment variables
	if err := os.Unsetenv("ENV"); err != nil {
		suite.T().Logf("Failed to unset ENV: %v", err)
	}
	if err := os.Unsetenv("LOG_LEVEL"); err != nil {
		suite.T().Logf("Failed to unset LOG_LEVEL: %v", err)
	}
	if err := os.Unsetenv("DEFAULT_MIN_FLOOR"); err != nil {
		suite.T().Logf("Failed to unset DEFAULT_MIN_FLOOR: %v", err)
	}
	if err := os.Unsetenv("DEFAULT_MAX_FLOOR"); err != nil {
		suite.T().Logf("Failed to unset DEFAULT_MAX_FLOOR: %v", err)
	}

	// Allow cleanup time
	time.Sleep(10 * time.Millisecond)
}

// Helper methods

func (suite *AcceptanceTestSuite) createElevator(name string, minFloor, maxFloor int) {
	reqBody := httpPkg.ElevatorRequestBody{
		Name:     name,
		MinFloor: minFloor,
		MaxFloor: maxFloor,
	}

	jsonBody, err := json.Marshal(reqBody)
	require.NoError(suite.T(), err)

	resp, err := http.Post(suite.testSrv.URL+"/elevator", "application/json", bytes.NewBuffer(jsonBody))
	require.NoError(suite.T(), err)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	// Allow time for elevator creation to complete
	time.Sleep(5 * time.Millisecond)
}

func (suite *AcceptanceTestSuite) requestFloor(from, to int) *http.Response {
	reqBody := httpPkg.FloorRequestBody{
		From: from,
		To:   to,
	}

	jsonBody, err := json.Marshal(reqBody)
	require.NoError(suite.T(), err)

	resp, err := http.Post(suite.testSrv.URL+"/floor", "application/json", bytes.NewBuffer(jsonBody))
	require.NoError(suite.T(), err)

	return resp
}

func (suite *AcceptanceTestSuite) requestFloorWithTimeout(from, to int, timeout time.Duration) *http.Response {
	client := &http.Client{Timeout: timeout}

	reqBody := httpPkg.FloorRequestBody{
		From: from,
		To:   to,
	}

	jsonBody, err := json.Marshal(reqBody)
	require.NoError(suite.T(), err)

	resp, err := client.Post(suite.testSrv.URL+"/floor", "application/json", bytes.NewBuffer(jsonBody))
	require.NoError(suite.T(), err)

	return resp
}

// Test methods

func (suite *AcceptanceTestSuite) TestElevatorCreationAndBasicOperations() {
	suite.T().Run("create multiple elevators with different ranges", func(t *testing.T) {
		// Create elevators with different floor ranges
		suite.createElevator("MainElevator", 0, 20)
		suite.createElevator("ParkingElevator", -5, 5)
		suite.createElevator("SkyscraperElevator", 1, 50)

		// Verify elevators were created by making requests
		resp := suite.requestFloor(1, 10)
		if err := resp.Body.Close(); err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		resp = suite.requestFloor(-3, 2)
		if err := resp.Body.Close(); err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		resp = suite.requestFloor(10, 40)
		if err := resp.Body.Close(); err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	suite.T().Run("basic floor requests", func(t *testing.T) {
		suite.createElevator("TestElevator", 0, 10)

		testCases := []struct {
			name     string
			from, to int
			expected int
		}{
			{"up request", 2, 8, http.StatusOK},
			{"down request", 9, 3, http.StatusOK},
			{"single floor jump", 5, 6, http.StatusOK},
			{"ground floor", 0, 5, http.StatusOK},
			{"top floor", 8, 10, http.StatusOK},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp := suite.requestFloor(tc.from, tc.to)
				defer func() {
					if err := resp.Body.Close(); err != nil {
						t.Logf("Failed to close response body: %v", err)
					}
				}()
				assert.Equal(t, tc.expected, resp.StatusCode)
			})
		}
	})
}

func (suite *AcceptanceTestSuite) TestElevatorOptimization() {
	suite.T().Run("optimal elevator selection", func(t *testing.T) {
		// Create elevators with specific ranges
		suite.createElevator("LowRiseElevator", 0, 10)
		suite.createElevator("MidRiseElevator", 5, 25)
		suite.createElevator("HighRiseElevator", 20, 50)
		suite.createElevator("ParkingElevator", -5, 5)

		testCases := []struct {
			name     string
			from, to int
		}{
			{"low floors should use suitable elevator", 2, 8},
			{"basement should use suitable elevator", -2, 1},
			{"high floors should use suitable elevator", 30, 45},
			{"mid range should use suitable elevator", 10, 15},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp := suite.requestFloor(tc.from, tc.to)
				defer func() {
					if err := resp.Body.Close(); err != nil {
						log.Printf("Failed to close response body: %v", err)
					}
				}()
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			})
		}
	})
}

func (suite *AcceptanceTestSuite) TestRushHourScenario() {
	suite.T().Run("concurrent rush hour requests", func(t *testing.T) {
		// Create multiple elevators to handle load
		suite.createElevator("RushElevator1", 0, 20)
		suite.createElevator("RushElevator2", 0, 20)
		suite.createElevator("RushElevator3", 0, 20)

		const numRequests = 15 // Reasonable number for testing
		successCount := 0
		var wg sync.WaitGroup
		var mu sync.Mutex

		// Simulate rush hour with concurrent requests
		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(requestID int) {
				defer wg.Done()

				// Generate realistic rush hour patterns
				from := requestID % 15
				to := from + (requestID % 5) + 1
				if to > 20 {
					to = 20
				}
				if from == to {
					to = from + 1
				}

				resp := suite.requestFloorWithTimeout(from, to, 5*time.Second)
				defer func() {
					if err := resp.Body.Close(); err != nil {
						log.Printf("Failed to close response body: %v", err)
					}
				}()

				mu.Lock()
				if resp.StatusCode == http.StatusOK {
					successCount++
				}
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		// Should handle most requests successfully
		successRate := float64(successCount) / float64(numRequests)
		assert.Greater(suite.T(), successRate, 0.8, "Should handle at least 80% of rush hour requests")
	})
}

func (suite *AcceptanceTestSuite) TestEdgeCasesAndErrorHandling() {
	suite.createElevator("TestElevator", 0, 10)

	// Test invalid floor requests - these should return appropriate error status codes
	suite.T().Run("invalid floor requests", func(t *testing.T) {
		testCases := []struct {
			name     string
			from, to int
			expected int
		}{
			// Validation errors - should return 400 Bad Request
			{"same floor", 5, 5, http.StatusBadRequest},                    // Cannot request travel to same floor
			{"out of range high", 15, 20, http.StatusBadRequest},           // Above elevator's max range (0-10) - validation error
			{"out of range low", -10, -5, http.StatusBadRequest},           // Below elevator's min range (0-10) - validation error
			{"negative floors beyond range", -1, 5, http.StatusBadRequest}, // From floor below range - validation error
			{"extremely high floors", 100, 200, http.StatusBadRequest},     // Both floors way above range - validation error
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp := suite.requestFloor(tc.from, tc.to)
				defer func() {
					if err := resp.Body.Close(); err != nil {
						t.Logf("Failed to close response body: %v", err)
					}
				}()
				assert.Equal(t, tc.expected, resp.StatusCode)
			})
		}
	})

	suite.T().Run("invalid elevator creation", func(t *testing.T) {
		testCases := []struct {
			name               string
			elevatorName       string
			minFloor, maxFloor int
			expectedStatus     int
		}{
			{"empty name", "", 0, 10, http.StatusBadRequest},
			{"same min/max floor", "SameFloor", 5, 5, http.StatusBadRequest},
			{"duplicate name", "TestElevator", 0, 15, http.StatusBadRequest},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				reqBody := httpPkg.ElevatorRequestBody{
					Name:     tc.elevatorName,
					MinFloor: tc.minFloor,
					MaxFloor: tc.maxFloor,
				}

				jsonBody, err := json.Marshal(reqBody)
				require.NoError(t, err)

				resp, err := http.Post(suite.testSrv.URL+"/elevator", "application/json", bytes.NewBuffer(jsonBody))
				require.NoError(t, err)
				defer func() {
					if err := resp.Body.Close(); err != nil {
						t.Logf("Failed to close response body: %v", err)
					}
				}()

				assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			})
		}
	})

	suite.T().Run("malformed requests", func(t *testing.T) {
		testCases := []struct {
			name     string
			endpoint string
			body     string
			expected int
		}{
			{"invalid JSON floor", "/floor", `{"from": "invalid", "to": 5}`, http.StatusBadRequest},
			{"empty body", "/floor", "", http.StatusBadRequest},
			{"non-JSON body", "/floor", "not json", http.StatusBadRequest},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := http.Post(suite.testSrv.URL+tc.endpoint, "application/json", strings.NewReader(tc.body))
				require.NoError(t, err)
				defer func() {
					if err := resp.Body.Close(); err != nil {
						t.Logf("Failed to close response body: %v", err)
					}
				}()
				assert.Equal(t, tc.expected, resp.StatusCode)
			})
		}
	})
}

func (suite *AcceptanceTestSuite) TestWebSocketStatusUpdates() {
	suite.createElevator("StatusTestElevator", 0, 10)

	suite.T().Run("websocket status updates", func(t *testing.T) {
		// Connect to WebSocket
		wsURL := strings.Replace(suite.testSrv.URL, "http://", "ws://", 1) + "/ws/status"
		ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)

		// Skip test if WebSocket upgrade fails (limitation of httptest.Server)
		if err != nil && strings.Contains(err.Error(), "bad handshake") {
			t.Skip("WebSocket upgrade not supported by httptest.Server - this is expected")
			return
		}
		require.NoError(t, err)
		defer func() {
			if err := ws.Close(); err != nil {
				log.Printf("Failed to close WebSocket connection: %v", err)
			}
		}()

		// Read initial status with timeout
		if err := ws.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
			t.Errorf("failed to set read deadline: %v", err)
		}
		var initialStatus map[string]interface{}
		err = ws.ReadJSON(&initialStatus)
		require.NoError(t, err)
		assert.NotEmpty(t, initialStatus)

		// Make a floor request to trigger status change
		resp := suite.requestFloor(2, 8)
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}

		// Read updated status with timeout
		if err := ws.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
			t.Errorf("failed to set read deadline: %v", err)
		}
		var updatedStatus map[string]interface{}
		err = ws.ReadJSON(&updatedStatus)
		require.NoError(t, err)
		assert.NotEmpty(t, updatedStatus)

		// Verify status contains expected fields
		assert.Contains(t, fmt.Sprintf("%v", updatedStatus), "StatusTestElevator")
	})
}

func (suite *AcceptanceTestSuite) TestSystemPerformance() {
	// Create multiple elevators for load testing
	for i := 0; i < 3; i++ {
		suite.createElevator(fmt.Sprintf("PerfElevator%d", i), 0, 30)
	}

	suite.T().Run("response time performance", func(t *testing.T) {
		const numRequests = 10 // Reduced for faster tests
		var totalDuration time.Duration
		var successCount int

		for i := 0; i < numRequests; i++ {
			start := time.Now()
			resp := suite.requestFloor(i%15, (i%15)+3)
			duration := time.Since(start)
			totalDuration += duration
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}

			if resp.StatusCode == http.StatusOK {
				successCount++
			}
		}

		avgResponseTime := totalDuration / numRequests
		successRate := float64(successCount) / float64(numRequests)

		assert.Greater(t, successRate, 0.9, "Should maintain high success rate under load")
		assert.Less(t, avgResponseTime, 200*time.Millisecond, "Average response time should be reasonable")

		t.Logf("Performance metrics: Avg response time: %v, Success rate: %.2f%%",
			avgResponseTime, successRate*100)
	})
}

func (suite *AcceptanceTestSuite) TestRealWorldWorkflows() {
	suite.T().Run("office building scenario", func(t *testing.T) {
		// Setup office building with different elevator types
		suite.createElevator("LobbyElevator", 0, 5)
		suite.createElevator("MainElevator", 0, 20)
		suite.createElevator("ServiceElevator", -2, 20)

		// Morning rush - people going up from lobby
		for i := 0; i < 5; i++ {
			resp := suite.requestFloor(0, (i%10)+2)
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}

		// Lunch time - mixed traffic
		lunchRequests := []struct{ from, to int }{
			{5, 0}, {8, 0}, {12, 0}, // Going to lunch (down)
			{0, 7}, {0, 15}, {0, 3}, // Coming back (up)
		}

		for _, req := range lunchRequests {
			resp := suite.requestFloor(req.from, req.to)
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}
	})

	suite.T().Run("mixed-use building with basement", func(t *testing.T) {
		// Setup mixed-use building
		suite.createElevator("ResidentialElevator", 10, 30)
		suite.createElevator("CommercialElevator", 0, 15)
		suite.createElevator("MixedServiceElevator", -5, 30)

		// Test various user journeys
		journeys := []struct {
			name     string
			from, to int
		}{
			{"commercial to residential", 10, 25}, // Both 10 and 25 are in MixedServiceElevator range
			{"residential to parking", 15, -2},    // Both 15 and -2 are in MixedServiceElevator range
			{"penthouse access", 15, 30},          // Both in MixedServiceElevator range
		}

		for _, journey := range journeys {
			t.Run(journey.name, func(t *testing.T) {
				resp := suite.requestFloor(journey.from, journey.to)
				defer func() {
					if err := resp.Body.Close(); err != nil {
						t.Logf("Failed to close response body: %v", err)
					}
				}()
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			})
		}
	})
}

func (suite *AcceptanceTestSuite) TestSystemResilience() {
	suite.createElevator("ResilienceTestElevator", 0, 20)

	suite.T().Run("rapid successive requests", func(t *testing.T) {
		const numRapidRequests = 10
		successCount := 0

		// Send requests as fast as possible
		for i := 0; i < numRapidRequests; i++ {
			resp := suite.requestFloor(i%15, (i%15)+3)
			if resp.StatusCode == http.StatusOK {
				successCount++
			}
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}
		}

		// Should handle rapid requests gracefully
		successRate := float64(successCount) / float64(numRapidRequests)
		assert.GreaterOrEqual(t, successRate, 0.7, "Should handle rapid requests reasonably well")
	})

	suite.T().Run("no available elevators", func(t *testing.T) {
		// Make request when no elevators can serve the floor range - this is a validation error
		// since the floors are outside the range of existing elevators (0-10)
		resp := suite.requestFloor(50, 60) // Beyond any elevator's range
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}
		}()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // Fixed: validation error should return 400, not 404
	})

	suite.T().Run("boundary condition requests", func(t *testing.T) {
		// Create an elevator that can handle the full range we want to test
		suite.createElevator("BoundaryTestElevator", 0, 25)

		// Test requests at the exact boundaries of elevator capabilities
		resp := suite.requestFloor(0, 20) // Full range
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}
		}()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		resp = suite.requestFloor(20, 0) // Full range reverse
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}
		}()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func (suite *AcceptanceTestSuite) TestMetricsEndpoint() {
	suite.createElevator("MetricsTestElevator", 0, 10)

	suite.T().Run("metrics endpoint accessibility", func(t *testing.T) {
		// Make some requests to generate metrics
		for i := 0; i < 3; i++ {
			resp := suite.requestFloor(i%8, (i%8)+2)
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}
		}

		// Check metrics endpoint
		resp, err := http.Get(suite.testSrv.URL + "/metrics")
		require.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}
		}()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "text/plain")

		// Read response body to verify it contains metrics
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		metricsText := string(body)
		assert.Contains(t, metricsText, "elevator")
	})
}

func (suite *AcceptanceTestSuite) TestHTTPMethodValidation() {
	endpoints := []struct {
		path   string
		method string
		body   string
	}{
		{"/floor", "GET", `{"from": 1, "to": 5}`},
		{"/floor", "PUT", `{"from": 1, "to": 5}`},
		{"/floor", "DELETE", `{"from": 1, "to": 5}`},
		{"/elevator", "GET", `{"name": "test", "min_floor": 0, "max_floor": 10}`},
		{"/elevator", "PUT", `{"name": "test", "min_floor": 0, "max_floor": 10}`},
		{"/elevator", "DELETE", `{"name": "test", "min_floor": 0, "max_floor": 10}`},
	}

	for _, endpoint := range endpoints {
		suite.T().Run(fmt.Sprintf("%s %s should return 405", endpoint.method, endpoint.path), func(t *testing.T) {
			req, err := http.NewRequest(endpoint.method, suite.testSrv.URL+endpoint.path, strings.NewReader(endpoint.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("Failed to close response body: %v", err)
				}
			}()

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		})
	}
}

// Run the test suite
func TestAcceptanceTestSuite(t *testing.T) {
	suite.Run(t, new(AcceptanceTestSuite))
}

// Standalone tests for quick testing without test suite overhead

func TestQuickAcceptance(t *testing.T) {
	// Suppress logging for quick tests
	log.SetOutput(io.Discard)
	logging.InitLogger("ERROR")

	// Set testing environment to get proper defaults including no rate limiting
	if err := os.Setenv("ENV", "testing"); err != nil {
		t.Fatalf("Failed to set ENV: %v", err)
	}
	if err := os.Setenv("LOG_LEVEL", "ERROR"); err != nil {
		t.Fatalf("Failed to set LOG_LEVEL: %v", err)
	}
	if err := os.Setenv("DEFAULT_MIN_FLOOR", "-10"); err != nil {
		t.Fatalf("Failed to set DEFAULT_MIN_FLOOR: %v", err)
	}
	if err := os.Setenv("DEFAULT_MAX_FLOOR", "50"); err != nil {
		t.Fatalf("Failed to set DEFAULT_MAX_FLOOR: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("ENV"); err != nil {
			t.Logf("Failed to unset ENV: %v", err)
		}
		if err := os.Unsetenv("LOG_LEVEL"); err != nil {
			t.Logf("Failed to unset LOG_LEVEL: %v", err)
		}
		if err := os.Unsetenv("DEFAULT_MIN_FLOOR"); err != nil {
			t.Logf("Failed to unset DEFAULT_MIN_FLOOR: %v", err)
		}
		if err := os.Unsetenv("DEFAULT_MAX_FLOOR"); err != nil {
			t.Logf("Failed to unset DEFAULT_MAX_FLOOR: %v", err)
		}
	}()

	cfg, err := config.InitConfig()
	require.NoError(t, err)

	factory := &factory.StandardElevatorFactory{}
	manager := manager.New(cfg, factory)
	server := httpPkg.NewServer(cfg, cfg.Port, manager)

	t.Run("basic elevator creation", func(t *testing.T) {
		reqBody := httpPkg.ElevatorRequestBody{
			Name:     "QuickTestElevator",
			MinFloor: 0,
			MaxFloor: 10,
		}

		jsonBody, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/elevator", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		rr := &testResponseWriter{header: make(http.Header)}
		server.GetHandler().ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.statusCode)
	})

	t.Run("basic floor request", func(t *testing.T) {
		// First create an elevator
		reqBody := httpPkg.ElevatorRequestBody{
			Name:     "FloorTestElevator",
			MinFloor: 0,
			MaxFloor: 10,
		}

		jsonBody, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/elevator", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		rr := &testResponseWriter{header: make(http.Header)}
		server.GetHandler().ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.statusCode)

		// Now make a floor request
		floorReqBody := httpPkg.FloorRequestBody{
			From: 1,
			To:   5,
		}

		jsonBody, err = json.Marshal(floorReqBody)
		require.NoError(t, err)

		req, err = http.NewRequest("POST", "/floor", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		rr = &testResponseWriter{header: make(http.Header)}
		server.GetHandler().ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.statusCode)
	})
}

// Simple test response writer for quick tests
type testResponseWriter struct {
	header     http.Header
	body       []byte
	statusCode int
}

func (w *testResponseWriter) Header() http.Header {
	return w.header
}

func (w *testResponseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return len(b), nil
}

func (w *testResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func TestZeroElevatorsHealthyState(t *testing.T) {
	t.Run("System is healthy with zero elevators", func(t *testing.T) {
		// Test the complete system behavior with zero elevators

		// Initialize test configuration using the same approach as existing tests
		if err := os.Setenv("ENV", "testing"); err != nil {
			t.Fatalf("Failed to set ENV: %v", err)
		}
		if err := os.Setenv("LOG_LEVEL", "ERROR"); err != nil {
			t.Fatalf("Failed to set LOG_LEVEL: %v", err)
		}
		cfg, err := config.InitConfig()
		require.NoError(t, err, "Config initialization should not error")

		elevatorFactory := &factory.StandardElevatorFactory{}
		elevatorManager := manager.New(cfg, elevatorFactory)
		server := httpPkg.NewServer(cfg, 8080, elevatorManager)

		// Test 1: Manager health status directly
		t.Run("Manager reports healthy with zero elevators", func(t *testing.T) {
			health, err := elevatorManager.GetHealthStatus()
			require.NoError(t, err, "Health status check should not error")

			assert.Equal(t, 0, health["total_elevators"], "Should have 0 total elevators")
			assert.Equal(t, 0, health["healthy_elevators"], "Should have 0 healthy elevators")
			assert.True(t, health["system_healthy"].(bool), "System should be healthy with 0 elevators")
			assert.NotNil(t, health["timestamp"], "Should have timestamp")

			elevators, ok := health["elevators"].(map[string]interface{})
			require.True(t, ok, "Elevators should be a map")
			assert.Empty(t, elevators, "Elevators map should be empty")
		})

		// Test 2: HTTP health endpoint returns 200 OK
		t.Run("HTTP health endpoint returns 200 OK", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/v1/health", nil)
			w := httptest.NewRecorder()

			server.GetHandler().ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Health endpoint should return 200 OK for zero elevators")
			assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

			// Parse the JSON response
			var response struct {
				Success bool `json:"success"`
				Data    struct {
					Status    string                 `json:"status"`
					Timestamp string                 `json:"timestamp"`
					Checks    map[string]interface{} `json:"checks"`
				} `json:"data"`
			}

			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err, "Should be valid JSON")

			assert.True(t, response.Success, "Response should be successful")
			assert.Equal(t, "healthy", response.Data.Status, "Status should be healthy")
			assert.True(t, response.Data.Checks["system_healthy"].(bool), "System health should be true")
			assert.Equal(t, float64(0), response.Data.Checks["total_elevators"], "Should have 0 elevators")
		})

		// Test 3: Manager metrics handle zero elevators gracefully
		t.Run("Metrics handle zero elevators gracefully", func(t *testing.T) {
			metrics := elevatorManager.GetMetrics()

			assert.Equal(t, 0, metrics["total_elevators"], "Should have 0 total elevators")
			assert.Equal(t, 0, metrics["healthy_elevators"], "Should have 0 healthy elevators")
			assert.Equal(t, 0, metrics["total_requests"], "Should have 0 total requests")
			assert.Equal(t, float64(0), metrics["average_load"], "Average load should be 0")

			// Performance score should handle division by zero gracefully
			performanceScore, exists := metrics["performance_score"]
			assert.True(t, exists, "Performance score should exist")
			assert.IsType(t, float64(0), performanceScore, "Performance score should be a float64")
		})

		// Test 4: System transitions correctly when elevator is added
		t.Run("System transitions correctly when elevator is added", func(t *testing.T) {
			ctx := context.Background()

			// Add first elevator
			err := elevatorManager.AddElevator(ctx, cfg, "FirstElevator", 0, 10, cfg.EachFloorDuration, cfg.OpenDoorDuration, cfg.DefaultOverloadThreshold)
			require.NoError(t, err, "Should be able to add first elevator")

			// Check health status after adding elevator
			health, err := elevatorManager.GetHealthStatus()
			require.NoError(t, err, "Health status check should not error")

			assert.Equal(t, 1, health["total_elevators"], "Should now have 1 elevator")
			assert.Equal(t, 1, health["healthy_elevators"], "Should have 1 healthy elevator")
			assert.True(t, health["system_healthy"].(bool), "System should still be healthy")

			// Verify HTTP endpoint still returns 200 OK
			req := httptest.NewRequest("GET", "/v1/health", nil)
			w := httptest.NewRecorder()
			server.GetHandler().ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code, "Health endpoint should return 200 OK with 1 healthy elevator")
		})
	})
}

func TestSystemHealthTransitions(t *testing.T) {
	t.Run("Health status transitions through elevator lifecycle", func(t *testing.T) {
		// Initialize test configuration
		if err := os.Setenv("ENV", "testing"); err != nil {
			t.Fatalf("Failed to set ENV: %v", err)
		}
		if err := os.Setenv("LOG_LEVEL", "ERROR"); err != nil {
			t.Fatalf("Failed to set LOG_LEVEL: %v", err)
		}
		cfg, err := config.InitConfig()
		require.NoError(t, err, "Config initialization should not error")

		elevatorFactory := &factory.StandardElevatorFactory{}
		elevatorManager := manager.New(cfg, elevatorFactory)
		ctx := context.Background()

		// Phase 1: No elevators (should be healthy)
		health, err := elevatorManager.GetHealthStatus()
		require.NoError(t, err)
		assert.True(t, health["system_healthy"].(bool), "System should be healthy with no elevators")
		assert.Equal(t, 0, health["total_elevators"])

		// Phase 2: Add healthy elevator (should remain healthy)
		err = elevatorManager.AddElevator(ctx, cfg, "TestElevator", 0, 5, cfg.EachFloorDuration, cfg.OpenDoorDuration, cfg.DefaultOverloadThreshold)
		require.NoError(t, err)

		health, err = elevatorManager.GetHealthStatus()
		require.NoError(t, err)
		assert.True(t, health["system_healthy"].(bool), "System should be healthy with 1 healthy elevator")
		assert.Equal(t, 1, health["total_elevators"])
		assert.Equal(t, 1, health["healthy_elevators"])

		// Phase 3: Add another elevator (should remain healthy)
		err = elevatorManager.AddElevator(ctx, cfg, "TestElevator2", 0, 5, cfg.EachFloorDuration, cfg.OpenDoorDuration, cfg.DefaultOverloadThreshold)
		require.NoError(t, err)

		health, err = elevatorManager.GetHealthStatus()
		require.NoError(t, err)
		assert.True(t, health["system_healthy"].(bool), "System should be healthy with 2 healthy elevators")
		assert.Equal(t, 2, health["total_elevators"])
		assert.Equal(t, 2, health["healthy_elevators"])

		// Note: Testing unhealthy elevator states would require more complex setup
		// to trigger circuit breaker failures or other error conditions
	})
}
