package observability

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TestDataDogClient(t *testing.T) {
	logger := slog.Default()

	t.Run("create new DataDog client", func(t *testing.T) {
		config := &DataDogConfig{
			Enabled:   true,
			APIKey:    "test-api-key",
			Site:      "datadoghq.com",
			Host:      "localhost",
			Port:      8125,
			Namespace: "test",
		}

		client, err := NewDataDogClient(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, config, client.config)
		assert.Equal(t, logger, client.logger)
		assert.NotNil(t, client.httpClient)
		assert.Equal(t, 10*time.Second, client.httpClient.Timeout)
	})

	t.Run("record metric", func(t *testing.T) {
		config := &DataDogConfig{
			Enabled:   true,
			APIKey:    "test-api-key",
			Namespace: "test",
		}

		client, err := NewDataDogClient(config, logger)
		require.NoError(t, err)

		metricName := "test.metric"
		value := 42.0
		labels := map[string]string{
			"environment": "test",
			"service":     "elevator",
		}

		// Should not panic
		assert.NotPanics(t, func() {
			client.RecordMetric(metricName, value, labels)
		})
	})

	t.Run("send trace", func(t *testing.T) {
		config := &DataDogConfig{
			Enabled:    true,
			APMEnabled: true,
			APMHost:    "localhost",
			APMPort:    8126,
		}

		client, err := NewDataDogClient(config, logger)
		require.NoError(t, err)

		// Create a test span
		tracer := otel.Tracer("test")
		ctx, span := tracer.Start(context.Background(), "test-span",
			trace.WithAttributes(attribute.String("test.key", "test.value")))

		// Should not panic
		assert.NotPanics(t, func() {
			client.SendTrace(span)
		})

		span.End()
		_ = ctx
	})

	t.Run("send log", func(t *testing.T) {
		config := &DataDogConfig{
			Enabled:     true,
			LogEnabled:  true,
			LogEndpoint: "https://http-intake.logs.datadoghq.com/v1/input/test-key",
		}

		client, err := NewDataDogClient(config, logger)
		require.NoError(t, err)

		logEntry := LogEntry{
			Timestamp:   time.Now().UTC(),
			Level:       "info",
			Message:     "test log message",
			Service:     "test-service",
			Version:     "1.0.0",
			Environment: "test",
			Fields: map[string]interface{}{
				"user_id": "123",
				"action":  "login",
			},
		}

		// Should not panic
		assert.NotPanics(t, func() {
			client.SendLog(logEntry)
		})
	})

	t.Run("close client", func(t *testing.T) {
		config := &DataDogConfig{Enabled: true}
		client, err := NewDataDogClient(config, logger)
		require.NoError(t, err)

		err = client.Close()
		assert.NoError(t, err)
	})
}

func TestElasticClient(t *testing.T) {
	logger := slog.Default()

	t.Run("create new Elastic client", func(t *testing.T) {
		config := &ElasticConfig{
			Enabled:        true,
			Host:           "localhost",
			Port:           9200,
			Username:       "elastic",
			Password:       "password",
			Index:          "test-logs",
			IndexRotation:  "daily",
			Timeout:        30 * time.Second,
			LogsEnabled:    true,
			MetricsEnabled: true,
			TracesEnabled:  true,
		}

		client, err := NewElasticClient(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, config, client.config)
		assert.Equal(t, logger, client.logger)
		assert.NotNil(t, client.httpClient)
		assert.Equal(t, config.Timeout, client.httpClient.Timeout)
	})

	t.Run("record metric", func(t *testing.T) {
		config := &ElasticConfig{
			Enabled:        true,
			MetricsEnabled: true,
			Index:          "test-metrics",
		}

		client, err := NewElasticClient(config, logger)
		require.NoError(t, err)

		metricName := "elasticsearch.metric"
		value := 100.5
		labels := map[string]string{
			"node":    "node-1",
			"cluster": "test-cluster",
		}

		// Should not panic
		assert.NotPanics(t, func() {
			client.RecordMetric(metricName, value, labels)
		})
	})

	t.Run("send trace", func(t *testing.T) {
		config := &ElasticConfig{
			Enabled:       true,
			TracesEnabled: true,
		}

		client, err := NewElasticClient(config, logger)
		require.NoError(t, err)

		// Create a test span
		tracer := otel.Tracer("test")
		ctx, span := tracer.Start(context.Background(), "elastic-test-span")

		// Should not panic
		assert.NotPanics(t, func() {
			client.SendTrace(span)
		})

		span.End()
		_ = ctx
	})

	t.Run("send log when enabled", func(t *testing.T) {
		config := &ElasticConfig{
			Enabled:       true,
			LogsEnabled:   true,
			Index:         "test-logs",
			IndexRotation: "daily",
		}

		client, err := NewElasticClient(config, logger)
		require.NoError(t, err)

		logEntry := LogEntry{
			Timestamp:   time.Now().UTC(),
			Level:       "error",
			Message:     "test error message",
			Service:     "test-service",
			Environment: "production",
			Fields: map[string]interface{}{
				"error_code": 500,
				"endpoint":   "/api/test",
			},
		}

		// Should not panic
		assert.NotPanics(t, func() {
			client.SendLog(logEntry)
		})
	})

	t.Run("send log when disabled", func(t *testing.T) {
		config := &ElasticConfig{
			Enabled:     true,
			LogsEnabled: false, // Logs disabled
		}

		client, err := NewElasticClient(config, logger)
		require.NoError(t, err)

		logEntry := LogEntry{
			Timestamp: time.Now().UTC(),
			Level:     "info",
			Message:   "this should not be sent",
		}

		// Should not panic even when logs are disabled
		assert.NotPanics(t, func() {
			client.SendLog(logEntry)
		})
	})

	t.Run("get index name - daily rotation", func(t *testing.T) {
		config := &ElasticConfig{
			Index:         "test-logs",
			IndexRotation: "daily",
		}

		client, err := NewElasticClient(config, logger)
		require.NoError(t, err)

		timestamp := time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC)
		indexName := client.getIndexName(timestamp)
		assert.Equal(t, "test-logs-2023.12.25", indexName)
	})

	t.Run("get index name - weekly rotation", func(t *testing.T) {
		config := &ElasticConfig{
			Index:         "app-logs",
			IndexRotation: "weekly",
		}

		client, err := NewElasticClient(config, logger)
		require.NoError(t, err)

		timestamp := time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC) // Week 52 of 2023
		indexName := client.getIndexName(timestamp)
		assert.Equal(t, "app-logs-2023.52", indexName)
	})

	t.Run("get index name - monthly rotation", func(t *testing.T) {
		config := &ElasticConfig{
			Index:         "monthly-logs",
			IndexRotation: "monthly",
		}

		client, err := NewElasticClient(config, logger)
		require.NoError(t, err)

		timestamp := time.Date(2023, 7, 15, 14, 45, 0, 0, time.UTC)
		indexName := client.getIndexName(timestamp)
		assert.Equal(t, "monthly-logs-2023.07", indexName)
	})

	t.Run("get index name - invalid rotation (default)", func(t *testing.T) {
		config := &ElasticConfig{
			Index:         "static-logs",
			IndexRotation: "invalid",
		}

		client, err := NewElasticClient(config, logger)
		require.NoError(t, err)

		timestamp := time.Now()
		indexName := client.getIndexName(timestamp)
		assert.Equal(t, "static-logs", indexName)
	})

	t.Run("close client", func(t *testing.T) {
		config := &ElasticConfig{Enabled: true}
		client, err := NewElasticClient(config, logger)
		require.NoError(t, err)

		err = client.Close()
		assert.NoError(t, err)
	})
}

func TestPrometheusClient(t *testing.T) {
	logger := slog.Default()

	t.Run("create new Prometheus client with push enabled", func(t *testing.T) {
		config := &PrometheusConfig{
			Enabled:      true,
			PushEnabled:  true,
			PushGateway:  "http://localhost:9091",
			PushInterval: 15 * time.Second,
			PushJob:      "test-job",
			PushTimeout:  10 * time.Second,
		}

		client, err := NewPrometheusClient(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, config, client.config)
		assert.Equal(t, logger, client.logger)
		assert.NotNil(t, client.httpClient)
		assert.Equal(t, config.PushTimeout, client.httpClient.Timeout)
	})

	t.Run("create new Prometheus client with push disabled", func(t *testing.T) {
		config := &PrometheusConfig{
			Enabled:     true,
			PushEnabled: false, // Push disabled
		}

		client, err := NewPrometheusClient(config, logger)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "prometheus push gateway not configured")
	})

	t.Run("create new Prometheus client without gateway URL", func(t *testing.T) {
		config := &PrometheusConfig{
			Enabled:     true,
			PushEnabled: true,
			PushGateway: "", // No gateway URL
		}

		client, err := NewPrometheusClient(config, logger)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "prometheus push gateway not configured")
	})

	t.Run("record metric", func(t *testing.T) {
		config := &PrometheusConfig{
			Enabled:     true,
			PushEnabled: true,
			PushGateway: "http://localhost:9091",
			PushJob:     "elevator-metrics",
		}

		client, err := NewPrometheusClient(config, logger)
		require.NoError(t, err)

		metricName := "prometheus_test_metric"
		value := 123.45
		labels := map[string]string{
			"job":      "test",
			"instance": "localhost:8080",
		}

		// Should not panic
		assert.NotPanics(t, func() {
			client.RecordMetric(metricName, value, labels)
		})
	})

	t.Run("close client", func(t *testing.T) {
		config := &PrometheusConfig{
			Enabled:     true,
			PushEnabled: true,
			PushGateway: "http://localhost:9091",
		}

		client, err := NewPrometheusClient(config, logger)
		require.NoError(t, err)

		err = client.Close()
		assert.NoError(t, err)
	})
}

func TestOTLPClient(t *testing.T) {
	logger := slog.Default()

	t.Run("create new OTLP client", func(t *testing.T) {
		config := &OTLPConfig{
			Enabled:     true,
			Endpoint:    "http://localhost:4317",
			Insecure:    true,
			Timeout:     10 * time.Second,
			Compression: "gzip",
		}

		client, err := NewOTLPClient(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, config, client.config)
		assert.Equal(t, logger, client.logger)
		assert.NotNil(t, client.httpClient)
		assert.Equal(t, config.Timeout, client.httpClient.Timeout)
	})

	t.Run("create OTLP client with HTTP endpoint", func(t *testing.T) {
		config := &OTLPConfig{
			Enabled:      true,
			HTTPEndpoint: "http://localhost:4318",
			Insecure:     true,
			Headers:      "x-api-key=secret",
		}

		client, err := NewOTLPClient(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("record metric", func(t *testing.T) {
		config := &OTLPConfig{
			Enabled:  true,
			Endpoint: "http://localhost:4317",
		}

		client, err := NewOTLPClient(config, logger)
		require.NoError(t, err)

		metricName := "otlp.test.metric"
		value := 99.99
		labels := map[string]string{
			"service.name":    "test-service",
			"service.version": "1.0.0",
		}

		// Should not panic
		assert.NotPanics(t, func() {
			client.RecordMetric(metricName, value, labels)
		})
	})

	t.Run("send trace", func(t *testing.T) {
		config := &OTLPConfig{
			Enabled:  true,
			Endpoint: "http://localhost:4317",
		}

		client, err := NewOTLPClient(config, logger)
		require.NoError(t, err)

		// Create a test span
		tracer := otel.Tracer("test")
		ctx, span := tracer.Start(context.Background(), "otlp-test-span",
			trace.WithAttributes(
				attribute.String("http.method", "GET"),
				attribute.String("http.url", "/api/test"),
			))

		// Should not panic
		assert.NotPanics(t, func() {
			client.SendTrace(span)
		})

		span.End()
		_ = ctx
	})

	t.Run("close client", func(t *testing.T) {
		config := &OTLPConfig{Enabled: true}
		client, err := NewOTLPClient(config, logger)
		require.NoError(t, err)

		err = client.Close()
		assert.NoError(t, err)
	})
}

func TestStructuredLogger(t *testing.T) {
	logger := slog.Default()

	t.Run("create new structured logger", func(t *testing.T) {
		config := &LoggingConfig{
			Enabled: true,
			Level:   "info",
			Format:  "json",
		}

		provider := &TelemetryProvider{
			config: &ObservabilityConfig{
				Enabled:     true,
				ServiceName: "test-service",
			},
			logger: logger,
		}

		structuredLogger := NewStructuredLogger(config, provider)
		assert.NotNil(t, structuredLogger)
		assert.Equal(t, provider, structuredLogger.telemetryProvider)
		assert.Equal(t, logger, structuredLogger.logger)
	})

	t.Run("log message", func(t *testing.T) {
		config := &LoggingConfig{
			Enabled: true,
			Level:   "debug",
		}

		provider := &TelemetryProvider{
			config: &ObservabilityConfig{
				ServiceName: "test-service",
			},
			logger: logger,
		}

		structuredLogger := NewStructuredLogger(config, provider)

		fields := map[string]interface{}{
			"user_id":     "12345",
			"session_id":  "abcdef",
			"duration_ms": 150,
		}

		// Should not panic
		assert.NotPanics(t, func() {
			structuredLogger.Log("info", "User action completed", fields)
		})
	})

	t.Run("log message without telemetry provider", func(t *testing.T) {
		config := &LoggingConfig{
			Enabled: true,
		}

		structuredLogger := NewStructuredLogger(config, nil)

		// Should not panic even without telemetry provider
		assert.NotPanics(t, func() {
			structuredLogger.Log("error", "Error occurred", map[string]interface{}{
				"error_code": 500,
			})
		})
	})
}

func TestClientsHelperFunctions(t *testing.T) {
	t.Run("parseKeyValuePairs", func(t *testing.T) {
		tests := []struct {
			input    string
			expected map[string]string
		}{
			{
				input:    "key1=value1,key2=value2",
				expected: map[string]string{"key1": "value1", "key2": "value2"},
			},
			{
				input:    "env=production,region=us-east-1,version=1.0.0",
				expected: map[string]string{"env": "production", "region": "us-east-1", "version": "1.0.0"},
			},
			{
				input:    "single=value",
				expected: map[string]string{"single": "value"},
			},
			{
				input:    "",
				expected: map[string]string{},
			},
			{
				input:    "invalid_format",
				expected: map[string]string{},
			},
			{
				input:    "key1=,key2=value2",
				expected: map[string]string{"key2": "value2"}, // Empty values are filtered out
			},
			{
				input:    "key1=value1,key2=",
				expected: map[string]string{"key1": "value1"}, // Empty values are filtered out
			},
		}

		for _, test := range tests {
			result := parseKeyValuePairs(test.input)
			assert.Equal(t, test.expected, result, "input: %s", test.input)
		}
	})

	t.Run("splitAndTrim", func(t *testing.T) {
		tests := []struct {
			input     string
			separator string
			expected  []string
		}{
			{
				input:     "a,b,c",
				separator: ",",
				expected:  []string{"a", "b", "c"},
			},
			{
				input:     "a, b , c ",
				separator: ",",
				expected:  []string{"a", "b", "c"},
			},
			{
				input:     "one|two|three",
				separator: "|",
				expected:  []string{"one", "two", "three"},
			},
			{
				input:     "single",
				separator: ",",
				expected:  []string{"single"},
			},
			{
				input:     "",
				separator: ",",
				expected:  []string{},
			},
			{
				input:     "  ",
				separator: ",",
				expected:  []string{},
			},
		}

		for _, test := range tests {
			result := splitAndTrim(test.input, test.separator)
			assert.Equal(t, test.expected, result, "input: %s, separator: %s", test.input, test.separator)
		}
	})

	t.Run("splitString", func(t *testing.T) {
		tests := []struct {
			input     string
			separator string
			expected  []string
		}{
			{
				input:     "a,b,c",
				separator: ",",
				expected:  []string{"a", "b", "c"},
			},
			{
				input:     "hello world",
				separator: " ",
				expected:  []string{"hello", "world"},
			},
			{
				input:     "no-separator",
				separator: ",",
				expected:  []string{"no-separator"},
			},
			{
				input:     "",
				separator: ",",
				expected:  []string{}, // Empty input returns empty slice
			},
		}

		for _, test := range tests {
			result := splitString(test.input, test.separator)
			assert.Equal(t, test.expected, result, "input: %s, separator: %s", test.input, test.separator)
		}
	})

	t.Run("trimWhitespace", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{
				input:    "  hello world  ",
				expected: "hello world",
			},
			{
				input:    "\t\n  trimmed  \t\n",
				expected: "trimmed",
			},
			{
				input:    "no-spaces",
				expected: "no-spaces",
			},
			{
				input:    "   ",
				expected: "",
			},
			{
				input:    "",
				expected: "",
			},
			{
				input:    " a ",
				expected: "a",
			},
		}

		for _, test := range tests {
			result := trimWhitespace(test.input)
			assert.Equal(t, test.expected, result, "input: %q", test.input)
		}
	})

	t.Run("isWhitespace", func(t *testing.T) {
		tests := []struct {
			input    byte
			expected bool
		}{
			{' ', true},
			{'\t', true},
			{'\n', true},
			{'\r', true},
			{'a', false},
			{'1', false},
			{'_', false},
			{'-', false},
			{0, false},
		}

		for _, test := range tests {
			result := isWhitespace(test.input)
			assert.Equal(t, test.expected, result, "input: %q (%d)", test.input, test.input)
		}
	})
}

func TestClientIntegration(t *testing.T) {
	t.Run("full client workflow", func(t *testing.T) {
		logger := slog.Default()

		// Create configurations
		dataDogConfig := &DataDogConfig{
			Enabled:    true,
			APIKey:     "test-key",
			APMEnabled: true,
			LogEnabled: true,
		}

		elasticConfig := &ElasticConfig{
			Enabled:        true,
			Host:           "localhost",
			Port:           9200,
			LogsEnabled:    true,
			MetricsEnabled: true,
			TracesEnabled:  true,
			Index:          "test-logs",
			IndexRotation:  "daily",
			Timeout:        30 * time.Second,
		}

		prometheusConfig := &PrometheusConfig{
			Enabled:     true,
			PushEnabled: true,
			PushGateway: "http://localhost:9091",
			PushTimeout: 10 * time.Second,
		}

		otlpConfig := &OTLPConfig{
			Enabled:  true,
			Endpoint: "http://localhost:4317",
			Timeout:  10 * time.Second,
		}

		// Create clients
		dataDogClient, err := NewDataDogClient(dataDogConfig, logger)
		require.NoError(t, err)

		elasticClient, err := NewElasticClient(elasticConfig, logger)
		require.NoError(t, err)

		prometheusClient, err := NewPrometheusClient(prometheusConfig, logger)
		require.NoError(t, err)

		otlpClient, err := NewOTLPClient(otlpConfig, logger)
		require.NoError(t, err)

		// Test metrics
		metricName := "integration.test.metric"
		value := 42.0
		labels := map[string]string{
			"test": "integration",
			"env":  "test",
		}

		assert.NotPanics(t, func() {
			dataDogClient.RecordMetric(metricName, value, labels)
			elasticClient.RecordMetric(metricName, value, labels)
			prometheusClient.RecordMetric(metricName, value, labels)
			otlpClient.RecordMetric(metricName, value, labels)
		})

		// Test traces
		tracer := otel.Tracer("integration-test")
		ctx, span := tracer.Start(context.Background(), "integration-test-span")

		assert.NotPanics(t, func() {
			dataDogClient.SendTrace(span)
			elasticClient.SendTrace(span)
			otlpClient.SendTrace(span)
		})

		span.End()

		// Test logs
		logEntry := LogEntry{
			Timestamp:   time.Now().UTC(),
			Level:       "info",
			Message:     "integration test completed",
			Service:     "test-service",
			Environment: "test",
			Fields: map[string]interface{}{
				"test_type": "integration",
				"success":   true,
			},
		}

		assert.NotPanics(t, func() {
			dataDogClient.SendLog(logEntry)
			elasticClient.SendLog(logEntry)
		})

		// Close all clients
		assert.NoError(t, dataDogClient.Close())
		assert.NoError(t, elasticClient.Close())
		assert.NoError(t, prometheusClient.Close())
		assert.NoError(t, otlpClient.Close())

		_ = ctx
	})
}

func TestClientEdgeCases(t *testing.T) {
	logger := slog.Default()

	t.Run("clients with nil configurations", func(t *testing.T) {
		// Test with minimal/empty configs
		dataDogClient, err := NewDataDogClient(&DataDogConfig{}, logger)
		assert.NoError(t, err)
		assert.NotNil(t, dataDogClient)

		elasticClient, err := NewElasticClient(&ElasticConfig{Timeout: 1 * time.Second}, logger)
		assert.NoError(t, err)
		assert.NotNil(t, elasticClient)

		otlpClient, err := NewOTLPClient(&OTLPConfig{Timeout: 1 * time.Second}, logger)
		assert.NoError(t, err)
		assert.NotNil(t, otlpClient)
	})

	t.Run("operations with empty/nil data", func(t *testing.T) {
		dataDogConfig := &DataDogConfig{Enabled: true}
		client, err := NewDataDogClient(dataDogConfig, logger)
		require.NoError(t, err)

		// Test with empty metric name and labels
		assert.NotPanics(t, func() {
			client.RecordMetric("", 0, nil)
			client.RecordMetric("test", 0, map[string]string{})
		})

		// Test with empty log entry
		emptyLog := LogEntry{}
		assert.NotPanics(t, func() {
			client.SendLog(emptyLog)
		})
	})

	t.Run("large values and edge cases", func(t *testing.T) {
		elasticConfig := &ElasticConfig{
			Enabled: true,
			Timeout: 1 * time.Second,
		}
		client, err := NewElasticClient(elasticConfig, logger)
		require.NoError(t, err)

		// Test with very large metric value
		assert.NotPanics(t, func() {
			client.RecordMetric("large.metric", 1e10, map[string]string{
				"type": "large-number",
			})
		})

		// Test with very long strings
		longString := string(make([]byte, 10000))
		for i := range longString {
			longString = string(rune('a' + (i % 26)))
		}

		logEntry := LogEntry{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   longString,
		}

		assert.NotPanics(t, func() {
			client.SendLog(logEntry)
		})
	})
}
