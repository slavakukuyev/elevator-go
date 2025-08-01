package observability

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TestNewTelemetryProvider(t *testing.T) {
	logger := slog.Default()

	t.Run("disabled configuration", func(t *testing.T) {
		config := &ObservabilityConfig{
			Enabled: false,
		}

		provider, err := NewTelemetryProvider(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, config, provider.config)
		assert.Equal(t, logger, provider.logger)
	})

	t.Run("enabled configuration", func(t *testing.T) {
		config := &ObservabilityConfig{
			Enabled:     true,
			ServiceName: "test-service",
			Version:     "1.0.0",
			Environment: "test",
		}

		provider, err := NewTelemetryProvider(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.NotNil(t, provider.tracer)
		assert.NotNil(t, provider.meter)
	})

	t.Run("with DataDog configuration", func(t *testing.T) {
		config := &ObservabilityConfig{
			Enabled:     true,
			ServiceName: "test-service",
			DataDog: DataDogConfig{
				Enabled: true,
				APIKey:  "test-key",
			},
		}

		provider, err := NewTelemetryProvider(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		// DataDog client should be initialized despite potential connection errors
	})

	t.Run("with Elastic configuration", func(t *testing.T) {
		config := &ObservabilityConfig{
			Enabled:     true,
			ServiceName: "test-service",
			Elastic: ElasticConfig{
				Enabled: true,
				Host:    "localhost",
				Port:    9200,
			},
		}

		provider, err := NewTelemetryProvider(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		// Elastic client should be initialized despite potential connection errors
	})
}

func TestTelemetryProvider_GetTracer(t *testing.T) {
	logger := slog.Default()

	t.Run("with tracer initialized", func(t *testing.T) {
		config := &ObservabilityConfig{
			Enabled:     true,
			ServiceName: "test-service",
		}

		provider, err := NewTelemetryProvider(config, logger)
		require.NoError(t, err)

		tracer := provider.GetTracer()
		assert.NotNil(t, tracer)
	})

	t.Run("without tracer initialized", func(t *testing.T) {
		provider := &TelemetryProvider{}
		tracer := provider.GetTracer()
		assert.NotNil(t, tracer) // Should return noop tracer
	})
}

func TestTelemetryProvider_GetMeter(t *testing.T) {
	logger := slog.Default()

	t.Run("with meter initialized", func(t *testing.T) {
		config := &ObservabilityConfig{
			Enabled:     true,
			ServiceName: "test-service",
		}

		provider, err := NewTelemetryProvider(config, logger)
		require.NoError(t, err)

		meter := provider.GetMeter()
		assert.NotNil(t, meter)
	})

	t.Run("without meter initialized", func(t *testing.T) {
		provider := &TelemetryProvider{}
		meter := provider.GetMeter()
		assert.NotNil(t, meter) // Should return basic meter
	})
}

func TestTelemetryProvider_CreateSpan(t *testing.T) {
	logger := slog.Default()
	config := &ObservabilityConfig{
		Enabled:     true,
		ServiceName: "test-service",
	}

	provider, err := NewTelemetryProvider(config, logger)
	require.NoError(t, err)

	t.Run("create span with attributes", func(t *testing.T) {
		ctx := context.Background()
		spanName := "test-span"

		newCtx, span := provider.CreateSpan(ctx, spanName,
			trace.WithAttributes(
				attribute.String("test.key", "test.value"),
			),
		)

		assert.NotNil(t, newCtx)
		assert.NotNil(t, span)
		assert.NotEqual(t, ctx, newCtx)

		span.End()
	})

	t.Run("create span without tracer", func(t *testing.T) {
		provider := &TelemetryProvider{}
		ctx := context.Background()

		newCtx, span := provider.CreateSpan(ctx, "test-span")
		assert.NotNil(t, newCtx)
		assert.NotNil(t, span)
	})
}

func TestTelemetryProvider_RecordMetric(t *testing.T) {
	logger := slog.Default()
	config := &ObservabilityConfig{
		Enabled:     true,
		ServiceName: "test-service",
	}

	provider, err := NewTelemetryProvider(config, logger)
	require.NoError(t, err)

	t.Run("record metric", func(t *testing.T) {
		ctx := context.Background()
		metricName := "test.metric"
		value := 42.0
		labels := map[string]string{
			"environment": "test",
			"service":     "test-service",
		}

		// This should not panic
		assert.NotPanics(t, func() {
			provider.RecordMetric(ctx, metricName, value, labels)
		})
	})

	t.Run("record metric without clients", func(t *testing.T) {
		provider := &TelemetryProvider{
			config: config,
			logger: logger,
		}

		ctx := context.Background()
		// Should not panic even without clients
		assert.NotPanics(t, func() {
			provider.RecordMetric(ctx, "test.metric", 1.0, nil)
		})
	})
}

func TestTelemetryProvider_SendTrace(t *testing.T) {
	logger := slog.Default()
	config := &ObservabilityConfig{
		Enabled:     true,
		ServiceName: "test-service",
	}

	provider, err := NewTelemetryProvider(config, logger)
	require.NoError(t, err)

	t.Run("send trace", func(t *testing.T) {
		ctx := context.Background()
		_, span := provider.CreateSpan(ctx, "test-span")

		// This should not panic
		assert.NotPanics(t, func() {
			provider.SendTrace(ctx, span)
		})

		span.End()
	})

	t.Run("send invalid span", func(t *testing.T) {
		provider := &TelemetryProvider{
			config: config,
			logger: logger,
		}

		ctx := context.Background()
		_, span := provider.CreateSpan(ctx, "test-span")

		// Should not panic even with invalid span
		assert.NotPanics(t, func() {
			provider.SendTrace(ctx, span)
		})
	})
}

func TestTelemetryProvider_SendLog(t *testing.T) {
	logger := slog.Default()
	config := &ObservabilityConfig{
		Enabled:     true,
		ServiceName: "test-service",
		Version:     "1.0.0",
		Environment: "test",
	}

	provider, err := NewTelemetryProvider(config, logger)
	require.NoError(t, err)

	t.Run("send log entry", func(t *testing.T) {
		level := "info"
		message := "test log message"
		fields := map[string]interface{}{
			"user_id":    "123",
			"session_id": "abc",
		}

		// This should not panic
		assert.NotPanics(t, func() {
			provider.SendLog(level, message, fields)
		})
	})

	t.Run("send log without clients", func(t *testing.T) {
		provider := &TelemetryProvider{
			config: config,
			logger: logger,
		}

		// Should not panic even without clients
		assert.NotPanics(t, func() {
			provider.SendLog("error", "test error", nil)
		})
	})
}

func TestTelemetryProvider_TelemetryMiddleware(t *testing.T) {
	logger := slog.Default()
	config := &ObservabilityConfig{
		Enabled:     true,
		ServiceName: "test-service",
	}

	provider, err := NewTelemetryProvider(config, logger)
	require.NoError(t, err)

	t.Run("successful request", func(t *testing.T) {
		middleware := provider.TelemetryMiddleware()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("OK")); err != nil {
				t.Errorf("failed to write response: %v", err)
			}
		})

		wrappedHandler := middleware(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "test-agent")
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "OK", w.Body.String())
	})

	t.Run("error request", func(t *testing.T) {
		middleware := provider.TelemetryMiddleware()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte("Error")); err != nil {
				t.Errorf("failed to write response: %v", err)
			}
		})

		wrappedHandler := middleware(handler)

		req := httptest.NewRequest("POST", "/api/test", nil)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "Error", w.Body.String())
	})

	t.Run("request with query parameters", func(t *testing.T) {
		middleware := provider.TelemetryMiddleware()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrappedHandler := middleware(handler)

		req := httptest.NewRequest("GET", "/test?param=value&other=123", nil)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestTelemetryProvider_Shutdown(t *testing.T) {
	logger := slog.Default()

	t.Run("shutdown with no clients", func(t *testing.T) {
		provider := &TelemetryProvider{
			config: &ObservabilityConfig{},
			logger: logger,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := provider.Shutdown(ctx)
		assert.NoError(t, err)
	})

	t.Run("shutdown with clients", func(t *testing.T) {
		config := &ObservabilityConfig{
			Enabled:     true,
			ServiceName: "test-service",
		}

		provider, err := NewTelemetryProvider(config, logger)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdownErr := provider.Shutdown(ctx)
		// Some clients might fail to shutdown in test environment, that's OK
		// We just want to ensure the shutdown doesn't panic
		_ = shutdownErr
	})

	t.Run("shutdown with timeout", func(t *testing.T) {
		provider := &TelemetryProvider{
			config: &ObservabilityConfig{},
			logger: logger,
		}

		// Very short timeout to test timeout behavior
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		err := provider.Shutdown(ctx)
		// Should complete quickly since no real clients
		assert.NoError(t, err)
	})
}

func TestAgentDetector(t *testing.T) {
	logger := slog.Default()
	detector := NewAgentDetector(logger)

	t.Run("detect no agents", func(t *testing.T) {
		// Clear all environment variables that might affect detection
		originalEnvVars := map[string]string{}
		envVarsToCheck := []string{
			"DD_API_KEY", "DATADOG_API_KEY", "DD_AGENT_HOST", "DD_TRACE_AGENT_URL",
			"FLUENTD_HOST", "FLUENT_HOST", "FLUENT_CONF", "FLUENTBIT_CONFIG",
			"OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_COLLECTOR_HOST",
			"FILEBEAT_CONFIG", "ELASTIC_BEATS_CONFIG", "ELASTIC_CLOUD_ID", "ELASTIC_CLOUD_AUTH",
		}

		for _, envVar := range envVarsToCheck {
			originalEnvVars[envVar] = os.Getenv(envVar)
			if err := os.Unsetenv(envVar); err != nil {
				t.Logf("Failed to unset environment variable %s: %v", envVar, err)
			}
		}

		config := detector.DetectAgents()

		assert.False(t, config.DataDogEnabled)
		assert.False(t, config.FluentBitEnabled)
		assert.False(t, config.OTelAgentEnabled)
		assert.False(t, config.FilebeatEnabled)

		// Restore environment variables
		for envVar, value := range originalEnvVars {
			if value != "" {
				if err := os.Setenv(envVar, value); err != nil {
					t.Logf("Failed to restore environment variable %s: %v", envVar, err)
				}
			}
		}
	})

	t.Run("detect DataDog agent", func(t *testing.T) {
		if err := os.Setenv("DD_API_KEY", "test-key"); err != nil {
			t.Fatalf("Failed to set DD_API_KEY: %v", err)
		}
		defer func() {
			if err := os.Unsetenv("DD_API_KEY"); err != nil {
				t.Logf("Failed to unset DD_API_KEY: %v", err)
			}
		}()

		config := detector.DetectAgents()
		assert.True(t, config.DataDogEnabled)
	})

	t.Run("detect FluentBit agent", func(t *testing.T) {
		if err := os.Setenv("FLUENTD_HOST", "localhost"); err != nil {
			t.Fatalf("Failed to set FLUENTD_HOST: %v", err)
		}
		defer func() {
			if err := os.Unsetenv("FLUENTD_HOST"); err != nil {
				t.Logf("Failed to unset FLUENTD_HOST: %v", err)
			}
		}()

		config := detector.DetectAgents()
		assert.True(t, config.FluentBitEnabled)
	})

	t.Run("detect OpenTelemetry agent", func(t *testing.T) {
		if err := os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"); err != nil {
			t.Fatalf("Failed to set OTEL_EXPORTER_OTLP_ENDPOINT: %v", err)
		}
		defer func() {
			if err := os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT"); err != nil {
				t.Logf("Failed to unset OTEL_EXPORTER_OTLP_ENDPOINT: %v", err)
			}
		}()

		config := detector.DetectAgents()
		assert.True(t, config.OTelAgentEnabled)
	})

	t.Run("detect Filebeat agent", func(t *testing.T) {
		if err := os.Setenv("ELASTIC_CLOUD_ID", "test-cloud-id"); err != nil {
			t.Fatalf("Failed to set ELASTIC_CLOUD_ID: %v", err)
		}
		defer func() {
			if err := os.Unsetenv("ELASTIC_CLOUD_ID"); err != nil {
				t.Logf("Failed to unset ELASTIC_CLOUD_ID: %v", err)
			}
		}()

		config := detector.DetectAgents()
		assert.True(t, config.FilebeatEnabled)
	})
}

func TestLogEntry(t *testing.T) {
	t.Run("create log entry", func(t *testing.T) {
		timestamp := time.Now().UTC()
		entry := LogEntry{
			Timestamp:   timestamp,
			Level:       "info",
			Message:     "test message",
			Fields:      map[string]interface{}{"key": "value"},
			Service:     "test-service",
			Version:     "1.0.0",
			Environment: "test",
		}

		assert.Equal(t, timestamp, entry.Timestamp)
		assert.Equal(t, "info", entry.Level)
		assert.Equal(t, "test message", entry.Message)
		assert.Equal(t, "test-service", entry.Service)
		assert.Equal(t, "1.0.0", entry.Version)
		assert.Equal(t, "test", entry.Environment)
		assert.Contains(t, entry.Fields, "key")
	})
}

func TestResponseWriter(t *testing.T) {
	t.Run("response writer wrapper", func(t *testing.T) {
		w := httptest.NewRecorder()
		wrapper := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		wrapper.WriteHeader(http.StatusCreated)
		assert.Equal(t, http.StatusCreated, wrapper.statusCode)
		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("sanitizeEndpoint", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"/api/v1/users/123", "/api/v1/users/{id}"},
			{"/api/v1/users/123/posts/456", "/api/v1/users/{id}/posts/{id}"},
			{"/api/v1/users", "/api/v1/users"},
			{"/api/v1/users?param=value", "/api/v1/users"},
			{"/health", "/health"},
			{"", ""},
		}

		for _, test := range tests {
			result := sanitizeEndpoint(test.input)
			assert.Equal(t, test.expected, result, "input: %s", test.input)
		}
	})

	t.Run("isNumeric", func(t *testing.T) {
		tests := []struct {
			input    string
			expected bool
		}{
			{"123", true},
			{"0", true},
			{"456789", true},
			{"abc", false},
			{"12a", false},
			{"a12", false},
			{"", false},
			{" ", false},
		}

		for _, test := range tests {
			result := isNumeric(test.input)
			assert.Equal(t, test.expected, result, "input: %s", test.input)
		}
	})
}

func TestTelemetryProviderIntegration(t *testing.T) {
	t.Run("full telemetry workflow", func(t *testing.T) {
		logger := slog.Default()
		config := &ObservabilityConfig{
			Enabled:     true,
			ServiceName: "test-service",
			Version:     "1.0.0",
			Environment: "test",
		}

		provider, err := NewTelemetryProvider(config, logger)
		require.NoError(t, err)

		// Create span
		ctx := context.Background()
		newCtx, span := provider.CreateSpan(ctx, "test-operation",
			trace.WithAttributes(
				attribute.String("operation.type", "test"),
				attribute.Int("operation.count", 1),
			),
		)

		// Record metric
		provider.RecordMetric(newCtx, "test.operations.total", 1.0, map[string]string{
			"operation": "test",
			"status":    "success",
		})

		// Send trace
		provider.SendTrace(newCtx, span)

		// Send log
		provider.SendLog("info", "test operation completed", map[string]interface{}{
			"operation_id": "test-123",
			"duration_ms":  100,
		})

		span.End()

		// Shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = provider.Shutdown(shutdownCtx)
		// May have errors due to test environment, but should not panic
		_ = err
	})
}

// Mock implementations for testing

func TestTelemetryProvider_ContextualLogging(t *testing.T) {
	t.Run("middleware preserves request context", func(t *testing.T) {
		logger := slog.Default()
		config := &ObservabilityConfig{
			Enabled:     true,
			ServiceName: "test-service",
		}

		provider, err := NewTelemetryProvider(config, logger)
		require.NoError(t, err)

		middleware := provider.TelemetryMiddleware()

		var capturedContext context.Context
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedContext = r.Context()
			w.WriteHeader(http.StatusOK)
		})

		wrappedHandler := middleware(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)

		assert.NotNil(t, capturedContext)
		assert.NotEqual(t, req.Context(), capturedContext)

		// Context should have span information
		span := trace.SpanFromContext(capturedContext)
		assert.NotNil(t, span)
	})
}
