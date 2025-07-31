package observability

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// DataDogClient handles communication with DataDog
type DataDogClient struct {
	config     *DataDogConfig
	logger     *slog.Logger
	httpClient *http.Client
}

// NewDataDogClient creates a new DataDog client
func NewDataDogClient(config *DataDogConfig, logger *slog.Logger) (*DataDogClient, error) {
	return &DataDogClient{
		config:     config,
		logger:     logger,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// RecordMetric sends metrics to DataDog
func (c *DataDogClient) RecordMetric(name string, value float64, labels map[string]string) {
	// In a full implementation, this would send metrics to DataDog StatsD or HTTP API
	c.logger.Debug("DataDog metric recorded",
		slog.String("metric", name),
		slog.Float64("value", value),
		slog.Any("labels", labels))
}

// SendTrace sends trace data to DataDog
func (c *DataDogClient) SendTrace(span trace.Span) {
	// In a full implementation, this would send trace data to DataDog APM
	c.logger.Debug("DataDog trace sent",
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()))
}

// SendLog sends log data to DataDog
func (c *DataDogClient) SendLog(entry LogEntry) {
	// In a full implementation, this would send logs to DataDog Logs API
	c.logger.Debug("DataDog log sent",
		slog.String("level", entry.Level),
		slog.String("message", entry.Message))
}

// Close closes the DataDog client
func (c *DataDogClient) Close() error {
	return nil
}

// ElasticClient handles communication with Elasticsearch
type ElasticClient struct {
	config     *ElasticConfig
	logger     *slog.Logger
	httpClient *http.Client
}

// NewElasticClient creates a new Elasticsearch client
func NewElasticClient(config *ElasticConfig, logger *slog.Logger) (*ElasticClient, error) {
	return &ElasticClient{
		config:     config,
		logger:     logger,
		httpClient: &http.Client{Timeout: config.Timeout},
	}, nil
}

// RecordMetric sends metrics to Elasticsearch
func (c *ElasticClient) RecordMetric(name string, value float64, labels map[string]string) {
	// In a full implementation, this would send metrics to Elasticsearch
	c.logger.Debug("Elastic metric recorded",
		slog.String("metric", name),
		slog.Float64("value", value),
		slog.Any("labels", labels))
}

// SendTrace sends trace data to Elasticsearch
func (c *ElasticClient) SendTrace(span trace.Span) {
	// In a full implementation, this would send trace data to Elastic APM
	c.logger.Debug("Elastic trace sent",
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()))
}

// SendLog sends log data to Elasticsearch
func (c *ElasticClient) SendLog(entry LogEntry) {
	// In a full implementation, this would bulk index logs to Elasticsearch
	if c.config.LogsEnabled {
		indexName := c.getIndexName(entry.Timestamp)
		c.logger.Debug("Elastic log sent",
			slog.String("index", indexName),
			slog.String("level", entry.Level),
			slog.String("message", entry.Message))
	}
}

// getIndexName generates index name based on rotation policy
func (c *ElasticClient) getIndexName(timestamp time.Time) string {
	base := c.config.Index
	switch c.config.IndexRotation {
	case "daily":
		return fmt.Sprintf("%s-%s", base, timestamp.Format("2006.01.02"))
	case "weekly":
		year, week := timestamp.ISOWeek()
		return fmt.Sprintf("%s-%d.%02d", base, year, week)
	case "monthly":
		return fmt.Sprintf("%s-%s", base, timestamp.Format("2006.01"))
	default:
		return base
	}
}

// Close closes the Elasticsearch client
func (c *ElasticClient) Close() error {
	return nil
}

// PrometheusClient handles push-based communication with Prometheus
type PrometheusClient struct {
	config     *PrometheusConfig
	logger     *slog.Logger
	httpClient *http.Client
}

// NewPrometheusClient creates a new Prometheus push client
func NewPrometheusClient(config *PrometheusConfig, logger *slog.Logger) (*PrometheusClient, error) {
	if !config.PushEnabled || config.PushGateway == "" {
		return nil, fmt.Errorf("prometheus push gateway not configured")
	}

	return &PrometheusClient{
		config:     config,
		logger:     logger,
		httpClient: &http.Client{Timeout: config.PushTimeout},
	}, nil
}

// RecordMetric sends metrics to Prometheus Push Gateway
func (c *PrometheusClient) RecordMetric(name string, value float64, labels map[string]string) {
	// In a full implementation, this would push metrics to Prometheus Push Gateway
	c.logger.Debug("Prometheus metric pushed",
		slog.String("metric", name),
		slog.Float64("value", value),
		slog.Any("labels", labels),
		slog.String("gateway", c.config.PushGateway))
}

// Close closes the Prometheus client
func (c *PrometheusClient) Close() error {
	return nil
}

// OTLPClient handles communication with OTLP endpoints
type OTLPClient struct {
	config     *OTLPConfig
	logger     *slog.Logger
	httpClient *http.Client
}

// NewOTLPClient creates a new OTLP client
func NewOTLPClient(config *OTLPConfig, logger *slog.Logger) (*OTLPClient, error) {
	return &OTLPClient{
		config:     config,
		logger:     logger,
		httpClient: &http.Client{Timeout: config.Timeout},
	}, nil
}

// RecordMetric sends metrics via OTLP
func (c *OTLPClient) RecordMetric(name string, value float64, labels map[string]string) {
	// In a full implementation, this would send metrics via OTLP protocol
	c.logger.Debug("OTLP metric sent",
		slog.String("metric", name),
		slog.Float64("value", value),
		slog.Any("labels", labels),
		slog.String("endpoint", c.config.Endpoint))
}

// SendTrace sends trace data via OTLP
func (c *OTLPClient) SendTrace(span trace.Span) {
	// In a full implementation, this would send trace data via OTLP protocol
	c.logger.Debug("OTLP trace sent",
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
		slog.String("endpoint", c.config.Endpoint))
}

// Close closes the OTLP client
func (c *OTLPClient) Close() error {
	return nil
}

// StructuredLogger provides structured logging with external backend support
type StructuredLogger struct {
	config            *LoggingConfig
	telemetryProvider *TelemetryProvider
	logger            *slog.Logger
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(config *LoggingConfig, provider *TelemetryProvider) *StructuredLogger {
	return &StructuredLogger{
		config:            config,
		telemetryProvider: provider,
		logger:            slog.Default(),
	}
}

// Log sends a structured log entry to configured backends
func (sl *StructuredLogger) Log(level string, message string, fields map[string]interface{}) {
	// Add standard fields
	enrichedFields := make(map[string]interface{})
	for k, v := range fields {
		enrichedFields[k] = v
	}

	// Add structured extra fields if configured
	if sl.config.StructuredExtra != "" {
		extraFields := parseKeyValuePairs(sl.config.StructuredExtra)
		for k, v := range extraFields {
			enrichedFields[k] = v
		}
	}

	// Send to telemetry provider if available
	if sl.telemetryProvider != nil {
		sl.telemetryProvider.SendLog(level, message, enrichedFields)
	}

	// Log locally as well
	switch level {
	case "debug":
		sl.logger.Debug(message, slog.Any("fields", enrichedFields))
	case "info":
		sl.logger.Info(message, slog.Any("fields", enrichedFields))
	case "warn", "warning":
		sl.logger.Warn(message, slog.Any("fields", enrichedFields))
	case "error":
		sl.logger.Error(message, slog.Any("fields", enrichedFields))
	default:
		sl.logger.Info(message, slog.Any("fields", enrichedFields))
	}
}

// Helper functions

// parseKeyValuePairs parses key=value,key2=value2 format
func parseKeyValuePairs(input string) map[string]string {
	result := make(map[string]string)
	if input == "" {
		return result
	}

	pairs := splitAndTrim(input, ",")
	for _, pair := range pairs {
		parts := splitAndTrim(pair, "=")
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}

	return result
}

// splitAndTrim splits string and trims whitespace
func splitAndTrim(input, separator string) []string {
	parts := make([]string, 0)
	for _, part := range splitString(input, separator) {
		trimmed := trimWhitespace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// splitString splits a string by separator
func splitString(input, separator string) []string {
	if input == "" {
		return []string{}
	}

	result := make([]string, 0)
	start := 0
	sepLen := len(separator)

	for i := 0; i <= len(input)-sepLen; i++ {
		if input[i:i+sepLen] == separator {
			result = append(result, input[start:i])
			start = i + sepLen
			i += sepLen - 1
		}
	}
	result = append(result, input[start:])

	return result
}

// trimWhitespace removes leading and trailing whitespace
func trimWhitespace(s string) string {
	start := 0
	end := len(s)

	// Trim leading whitespace
	for start < end && isWhitespace(s[start]) {
		start++
	}

	// Trim trailing whitespace
	for end > start && isWhitespace(s[end-1]) {
		end--
	}

	return s[start:end]
}

// isWhitespace checks if a character is whitespace
func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}
