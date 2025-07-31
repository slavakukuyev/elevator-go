package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/slavakukuyev/elevator-go/internal/factory"
	httpPkg "github.com/slavakukuyev/elevator-go/internal/http"
	"github.com/slavakukuyev/elevator-go/internal/infra/config"
	"github.com/slavakukuyev/elevator-go/internal/infra/health"
	"github.com/slavakukuyev/elevator-go/internal/infra/logging"
	"github.com/slavakukuyev/elevator-go/internal/manager"
	"github.com/slavakukuyev/elevator-go/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonitoringAndObservability(t *testing.T) {
	// Initialize configuration for testing
	cfg := &config.Config{
		EachFloorDuration:     500 * time.Millisecond,
		OpenDoorDuration:      2 * time.Second,
		RequestTimeout:        5 * time.Second,
		HealthCheckTimeout:    2 * time.Second,
		StatusUpdateTimeout:   3 * time.Second,
		OperationTimeout:      30 * time.Second,
		CreateElevatorTimeout: 10 * time.Second,
		MaxElevators:          100,
		DefaultElevatorCount:  0,
		NamePrefix:            "Test-Elevator",
		SwitchOnChannelBuffer: 10,
		MetricsEnabled:        true,
		HealthEnabled:         true,
		StructuredLogging:     true,
		LogRequestDetails:     true,
		CorrelationIDHeader:   "X-Request-ID",
		RateLimitRPM:          10000, // High rate limit for testing
		RateLimitWindow:       1 * time.Minute,
		RateLimitCleanup:      5 * time.Minute,
	}

	// Initialize logging
	logging.InitLogger("INFO")

	// Create manager and server
	elevatorFactory := &factory.StandardElevatorFactory{}
	elevatorManager := manager.New(cfg, elevatorFactory)
	server := httpPkg.NewServer(cfg, 8080, elevatorManager)

	t.Run("Health Check System", func(t *testing.T) {
		testHealthCheckSystem(t, server)
	})

	t.Run("Metrics Collection", func(t *testing.T) {
		testMetricsCollection(t, server, elevatorManager, cfg)
	})

	t.Run("Performance Monitoring", func(t *testing.T) {
		testPerformanceMonitoring(t, server, elevatorManager, cfg)
	})

	t.Run("Correlation ID Tracking", func(t *testing.T) {
		testCorrelationIDTracking(t, server)
	})

	t.Run("Error Rate Monitoring", func(t *testing.T) {
		testErrorRateMonitoring(t, server)
	})
}

func testHealthCheckSystem(t *testing.T, server *httpPkg.Server) {
	t.Run("Liveness Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/health/live", nil)
		w := httptest.NewRecorder()

		server.GetHandler().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

		body := w.Body.String()
		assert.Contains(t, body, "liveness")
		assert.Contains(t, body, "Application is alive")
	})

	t.Run("Readiness Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/health/ready", nil)
		w := httptest.NewRecorder()

		server.GetHandler().ServeHTTP(w, req)

		// Readiness might fail if no elevators are configured, which is expected
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusServiceUnavailable)
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

		body := w.Body.String()
		assert.Contains(t, body, "readiness")
	})

	t.Run("Detailed Health Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/health/detailed", nil)
		w := httptest.NewRecorder()

		server.GetHandler().ServeHTTP(w, req)

		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusServiceUnavailable)
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

		body := w.Body.String()
		assert.Contains(t, body, "status")
		assert.Contains(t, body, "checks")
		assert.Contains(t, body, "summary")
		assert.Contains(t, body, "system_resources")
		assert.Contains(t, body, "liveness")
		assert.Contains(t, body, "manager")
	})
}

func testMetricsCollection(t *testing.T, server *httpPkg.Server, elevatorManager *manager.Manager, cfg *config.Config) {
	ctx := context.Background()

	// Add an elevator for testing
	err := elevatorManager.AddElevator(ctx, cfg, "TestElevator-1", 0, 10, cfg.EachFloorDuration, cfg.OpenDoorDuration, cfg.DefaultOverloadThreshold)
	require.NoError(t, err)

	t.Run("Request Metrics Collection", func(t *testing.T) {
		// Test metrics recording
		metrics.RecordRequestDuration("TestElevator-1", "success", 1.5)
		metrics.IncRequestsTotal("TestElevator-1", "up", "success")
		metrics.SetElevatorEfficiency("TestElevator-1", 0.95)
		metrics.RecordWaitTime("TestElevator-1", 10.0)
		metrics.RecordTravelTime("TestElevator-1", "5", 15.0)

		// Verify metrics can be gathered
		metricFamilies, err := prometheus.DefaultGatherer.Gather()
		require.NoError(t, err)

		// Check that our metrics are present
		foundMetrics := make(map[string]bool)
		for _, mf := range metricFamilies {
			name := mf.GetName()
			if strings.HasPrefix(name, "elevator_") {
				foundMetrics[name] = true
			}
		}

		expectedMetrics := []string{
			"elevator_request_duration_seconds",
			"elevator_requests_total",
			"elevator_efficiency_ratio",
			"elevator_wait_time_seconds",
			"elevator_travel_time_seconds",
		}

		for _, expectedMetric := range expectedMetrics {
			assert.True(t, foundMetrics[expectedMetric], "Expected metric %s not found", expectedMetric)
		}
	})

	t.Run("System Health Metrics", func(t *testing.T) {
		metrics.SetSystemHealth("elevators", true)
		metrics.SetSystemHealth("manager", true)
		metrics.SetCurrentFloor("TestElevator-1", 5.0)
		metrics.SetPendingRequests("TestElevator-1", "up", 2.0)
		metrics.SetCircuitBreakerState("TestElevator-1", 0.0) // closed

		// Get system metrics through manager
		systemMetrics := elevatorManager.GetMetrics()
		assert.Contains(t, systemMetrics, "total_elevators")
		assert.Contains(t, systemMetrics, "healthy_elevators")
		assert.Contains(t, systemMetrics, "performance_score")
		assert.Equal(t, 1, systemMetrics["total_elevators"])
	})
}

func testPerformanceMonitoring(t *testing.T, server *httpPkg.Server, elevatorManager *manager.Manager, cfg *config.Config) {
	ctx := context.Background()

	t.Run("HTTP Request Performance", func(t *testing.T) {
		// Add an elevator
		err := elevatorManager.AddElevator(ctx, cfg, "PerfTest-1", 0, 10, cfg.EachFloorDuration, cfg.OpenDoorDuration, cfg.DefaultOverloadThreshold)
		require.NoError(t, err)

		// Make a request and check response time tracking
		reqBody := `{"from": 0, "to": 5}`
		req := httptest.NewRequest("POST", "/v1/floors/request", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		start := time.Now()
		server.GetHandler().ServeHTTP(w, req)
		duration := time.Since(start)

		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest)
		assert.True(t, duration < 5*time.Second, "Request took too long: %v", duration)

		// Check that performance metrics were recorded
		metricFamilies, err := prometheus.DefaultGatherer.Gather()
		require.NoError(t, err)

		foundHTTPMetrics := false
		for _, mf := range metricFamilies {
			if strings.Contains(mf.GetName(), "http_request") {
				foundHTTPMetrics = true
				break
			}
		}
		assert.True(t, foundHTTPMetrics, "HTTP performance metrics not found")
	})

	t.Run("Memory Usage Tracking", func(t *testing.T) {
		metrics.SetMemoryUsage("alloc", 1024*1024) // 1MB
		metrics.SetMemoryUsage("sys", 2048*1024)   // 2MB

		// Verify memory metrics
		metricFamilies, err := prometheus.DefaultGatherer.Gather()
		require.NoError(t, err)

		foundMemoryMetrics := false
		for _, mf := range metricFamilies {
			if strings.Contains(mf.GetName(), "memory_usage") {
				foundMemoryMetrics = true
				break
			}
		}
		assert.True(t, foundMemoryMetrics, "Memory usage metrics not found")
	})
}

func testCorrelationIDTracking(t *testing.T, server *httpPkg.Server) {
	t.Run("Request ID Generation", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/health", nil)
		w := httptest.NewRecorder()

		server.GetHandler().ServeHTTP(w, req)

		// Check that request ID was added to response
		requestID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID, "Request ID should be generated and returned")
		assert.True(t, len(requestID) > 8, "Request ID should be sufficiently long")
	})

	t.Run("Request ID Preservation", func(t *testing.T) {
		existingRequestID := "test-request-123"
		req := httptest.NewRequest("GET", "/v1/health", nil)
		req.Header.Set("X-Request-ID", existingRequestID)
		w := httptest.NewRecorder()

		server.GetHandler().ServeHTTP(w, req)

		// Check that existing request ID was preserved
		returnedRequestID := w.Header().Get("X-Request-ID")
		assert.Equal(t, existingRequestID, returnedRequestID, "Existing request ID should be preserved")
	})
}

func testErrorRateMonitoring(t *testing.T, server *httpPkg.Server) {
	t.Run("404 Error Tracking", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/nonexistent", nil)
		w := httptest.NewRecorder()

		server.GetHandler().ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		// Error metrics should be automatically recorded by middleware
		metricFamilies, err := prometheus.DefaultGatherer.Gather()
		require.NoError(t, err)

		foundErrorMetrics := false
		for _, mf := range metricFamilies {
			if strings.Contains(mf.GetName(), "errors_total") || strings.Contains(mf.GetName(), "http_requests_total") {
				foundErrorMetrics = true
				break
			}
		}
		assert.True(t, foundErrorMetrics, "Error tracking metrics not found")
	})

	t.Run("Method Not Allowed Error", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/v1/health", nil)
		w := httptest.NewRecorder()

		server.GetHandler().ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

		// Check request ID is still present in error responses
		requestID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID, "Request ID should be present even in error responses")
	})
}

func TestHealthServiceStandalone(t *testing.T) {
	t.Run("Health Service Components", func(t *testing.T) {
		healthService := health.NewHealthService(10 * time.Second)

		// Add checkers
		resourceChecker := health.NewSystemResourceChecker(90.0, 1500)
		livenessChecker := health.NewLivenessChecker()

		healthService.Register(resourceChecker)
		healthService.Register(livenessChecker)

		ctx := context.Background()

		// Test individual check
		result, err := healthService.Check(ctx, "system_resources")
		require.NoError(t, err)
		assert.Equal(t, "system_resources", result.Name)
		assert.True(t, result.Status == health.StatusHealthy || result.Status == health.StatusDegraded)

		// Test overall health
		overallStatus, results := healthService.GetOverallStatus(ctx)
		assert.True(t, overallStatus == health.StatusHealthy || overallStatus == health.StatusDegraded)
		assert.Len(t, results, 2)
	})
}

func TestMetricsCollection(t *testing.T) {
	t.Run("Prometheus Metrics", func(t *testing.T) {
		// Test various metric types
		metrics.RecordRequestDuration("test-elevator", "success", 2.5)
		metrics.IncRequestsTotal("test-elevator", "up", "success")
		metrics.SetElevatorEfficiency("test-elevator", 0.85)
		metrics.RecordWaitTime("test-elevator", 30.0)
		metrics.SetSystemHealth("test-component", true)
		metrics.IncError("validation_error", "test-component")

		// Gather metrics
		metricFamilies, err := prometheus.DefaultGatherer.Gather()
		require.NoError(t, err)
		assert.True(t, len(metricFamilies) > 0, "Should have metrics registered")

		// Check metric families
		metricNames := make([]string, len(metricFamilies))
		for i, mf := range metricFamilies {
			metricNames[i] = mf.GetName()
		}

		expectedPrefixes := []string{"elevator_", "go_", "promhttp_"}
		foundExpected := false
		for _, name := range metricNames {
			for _, prefix := range expectedPrefixes {
				if strings.HasPrefix(name, prefix) {
					foundExpected = true
					break
				}
			}
			if foundExpected {
				break
			}
		}
		assert.True(t, foundExpected, "Should find metrics with expected prefixes")
	})
}
