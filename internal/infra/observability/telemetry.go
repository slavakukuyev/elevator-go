// Package observability provides telemetry infrastructure
// with support for multiple backends including Prometheus, DataDog, Elasticsearch, and OTLP.
package observability

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// TelemetryProvider provides a unified interface for all telemetry operations
type TelemetryProvider struct {
	config        *ObservabilityConfig
	logger        *slog.Logger
	tracer        trace.Tracer
	meter         metric.Meter
	shutdownFuncs []func(context.Context) error

	// External integrations
	dataDogClient    *DataDogClient
	elasticClient    *ElasticClient
	prometheusClient *PrometheusClient
	otlpClient       *OTLPClient
}

// NewTelemetryProvider creates a new telemetry provider with the given configuration
func NewTelemetryProvider(config *ObservabilityConfig, logger *slog.Logger) (*TelemetryProvider, error) {
	if !config.Enabled {
		return &TelemetryProvider{
			config: config,
			logger: logger,
		}, nil
	}

	provider := &TelemetryProvider{
		config: config,
		logger: logger,
	}

	// Initialize basic OpenTelemetry components
	provider.tracer = otel.Tracer("elevator-control-system")
	provider.meter = otel.Meter("elevator-control-system")

	// Set global propagator for distributed tracing
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Initialize external clients based on configuration
	if err := provider.initExternalClients(); err != nil {
		return nil, fmt.Errorf("failed to initialize external clients: %w", err)
	}

	provider.logger.Info("telemetry provider initialized",
		slog.String("service", config.ServiceName),
		slog.String("version", config.Version),
		slog.String("environment", config.Environment),
		slog.Any("exporters", config.GetActiveExporters()))

	return provider, nil
}

// initExternalClients initializes external observability clients
func (tp *TelemetryProvider) initExternalClients() error {
	var err error

	// Initialize DataDog client if enabled
	if tp.config.DataDog.Enabled {
		tp.dataDogClient, err = NewDataDogClient(&tp.config.DataDog, tp.logger)
		if err != nil {
			tp.logger.Error("failed to initialize DataDog client", slog.String("error", err.Error()))
		} else {
			tp.logger.Info("DataDog client initialized")
		}
	}

	// Initialize Elastic client if enabled
	if tp.config.Elastic.Enabled {
		tp.elasticClient, err = NewElasticClient(&tp.config.Elastic, tp.logger)
		if err != nil {
			tp.logger.Error("failed to initialize Elastic client", slog.String("error", err.Error()))
		} else {
			tp.logger.Info("Elastic client initialized")
		}
	}

	// Initialize Prometheus client if push is enabled
	if tp.config.Prometheus.Enabled && tp.config.Prometheus.PushEnabled {
		tp.prometheusClient, err = NewPrometheusClient(&tp.config.Prometheus, tp.logger)
		if err != nil {
			tp.logger.Error("failed to initialize Prometheus push client", slog.String("error", err.Error()))
		} else {
			tp.logger.Info("Prometheus push client initialized")
		}
	}

	// Initialize OTLP client if enabled
	if tp.config.OTLP.Enabled {
		tp.otlpClient, err = NewOTLPClient(&tp.config.OTLP, tp.logger)
		if err != nil {
			tp.logger.Error("failed to initialize OTLP client", slog.String("error", err.Error()))
		} else {
			tp.logger.Info("OTLP client initialized")
		}
	}

	return nil
}

// GetTracer returns the configured tracer
func (tp *TelemetryProvider) GetTracer() trace.Tracer {
	if tp.tracer == nil {
		return noop.NewTracerProvider().Tracer("noop")
	}
	return tp.tracer
}

// GetMeter returns the configured meter
func (tp *TelemetryProvider) GetMeter() metric.Meter {
	if tp.meter == nil {
		// Return a basic meter instead of noop
		return otel.Meter("noop")
	}
	return tp.meter
}

// CreateSpan creates a new span with the given name and options
func (tp *TelemetryProvider) CreateSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if tp.tracer == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	return tp.tracer.Start(ctx, name, opts...)
}

// RecordMetric records a metric value and sends to all configured backends
func (tp *TelemetryProvider) RecordMetric(ctx context.Context, name string, value float64, labels map[string]string) {
	// Send to DataDog if enabled
	if tp.dataDogClient != nil {
		tp.dataDogClient.RecordMetric(name, value, labels)
	}

	// Send to Elastic if enabled
	if tp.elasticClient != nil {
		tp.elasticClient.RecordMetric(name, value, labels)
	}

	// Send to Prometheus push gateway if enabled
	if tp.prometheusClient != nil {
		tp.prometheusClient.RecordMetric(name, value, labels)
	}

	// Send to OTLP if enabled
	if tp.otlpClient != nil {
		tp.otlpClient.RecordMetric(name, value, labels)
	}
}

// SendTrace sends trace data to configured backends
func (tp *TelemetryProvider) SendTrace(ctx context.Context, span trace.Span) {
	// Extract span information
	spanContext := span.SpanContext()
	if !spanContext.IsValid() {
		return
	}

	// Send to DataDog if enabled
	if tp.dataDogClient != nil && tp.config.DataDog.APMEnabled {
		tp.dataDogClient.SendTrace(span)
	}

	// Send to Elastic if enabled
	if tp.elasticClient != nil && tp.config.Elastic.TracesEnabled {
		tp.elasticClient.SendTrace(span)
	}

	// Send to OTLP if enabled
	if tp.otlpClient != nil {
		tp.otlpClient.SendTrace(span)
	}
}

// SendLog sends structured log data to configured backends
func (tp *TelemetryProvider) SendLog(level string, message string, fields map[string]interface{}) {
	logEntry := LogEntry{
		Timestamp:   time.Now().UTC(),
		Level:       level,
		Message:     message,
		Fields:      fields,
		Service:     tp.config.ServiceName,
		Version:     tp.config.Version,
		Environment: tp.config.Environment,
	}

	// Send to DataDog if enabled
	if tp.dataDogClient != nil && tp.config.DataDog.LogEnabled {
		tp.dataDogClient.SendLog(logEntry)
	}

	// Send to Elastic if enabled
	if tp.elasticClient != nil && tp.config.Elastic.LogsEnabled {
		tp.elasticClient.SendLog(logEntry)
	}
}

// TelemetryMiddleware provides HTTP middleware for automatic instrumentation
func (tp *TelemetryProvider) TelemetryMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create span for request
			ctx, span := tp.CreateSpan(r.Context(), "http_request",
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.url", r.URL.String()),
					attribute.String("http.user_agent", r.UserAgent()),
				),
			)
			defer func() {
				tp.SendTrace(ctx, span)
				span.End()
			}()

			// Add request to context
			r = r.WithContext(ctx)

			// Record request metric
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			// Record metrics
			duration := time.Since(start).Seconds()
			labels := map[string]string{
				"method":      r.Method,
				"status_code": fmt.Sprintf("%d", wrapped.statusCode),
				"endpoint":    sanitizeEndpoint(r.URL.Path),
			}

			tp.RecordMetric(ctx, "http_request_duration_seconds", duration, labels)
			tp.RecordMetric(ctx, "http_requests_total", 1, labels)

			// Add span attributes
			span.SetAttributes(
				attribute.Int("http.status_code", wrapped.statusCode),
				attribute.Float64("http.duration_seconds", duration),
			)

			// Log request completion
			logFields := map[string]interface{}{
				"method":           r.Method,
				"path":             r.URL.Path,
				"status_code":      wrapped.statusCode,
				"duration_seconds": duration,
				"remote_addr":      r.RemoteAddr,
			}

			if wrapped.statusCode >= 400 {
				tp.SendLog("error", "HTTP request failed", logFields)
			} else {
				tp.SendLog("info", "HTTP request completed", logFields)
			}
		})
	}
}

// Shutdown gracefully shuts down the telemetry provider
func (tp *TelemetryProvider) Shutdown(ctx context.Context) error {
	var errors []error

	// Execute all shutdown functions
	for _, shutdownFunc := range tp.shutdownFuncs {
		if err := shutdownFunc(ctx); err != nil {
			errors = append(errors, err)
			tp.logger.Error("error during telemetry shutdown", slog.String("error", err.Error()))
		}
	}

	// Shutdown external clients
	if tp.dataDogClient != nil {
		if err := tp.dataDogClient.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if tp.elasticClient != nil {
		if err := tp.elasticClient.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if tp.prometheusClient != nil {
		if err := tp.prometheusClient.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if tp.otlpClient != nil {
		if err := tp.otlpClient.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("telemetry shutdown errors: %v", errors)
	}

	tp.logger.Info("telemetry provider shutdown completed")
	return nil
}

// Helper types and functions

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	Service     string                 `json:"service"`
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Hijack implements http.Hijacker interface for WebSocket support
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("ResponseWriter does not implement http.Hijacker")
}

// sanitizeEndpoint sanitizes URL path for metrics
func sanitizeEndpoint(path string) string {
	// Remove query parameters
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	// Replace numeric IDs with placeholder
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if len(part) > 0 && isNumeric(part) {
			parts[i] = "{id}"
		}
	}

	return strings.Join(parts, "/")
}

// isNumeric checks if a string is numeric
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// AgentDetector provides functionality to detect and configure agents
type AgentDetector struct {
	logger *slog.Logger
}

// NewAgentDetector creates a new agent detector
func NewAgentDetector(logger *slog.Logger) *AgentDetector {
	return &AgentDetector{logger: logger}
}

// DetectAgents detects available observability agents in the environment
func (ad *AgentDetector) DetectAgents() *AgentConfig {
	config := &AgentConfig{}

	// Detect DataDog agent
	config.DataDogEnabled = ad.detectDataDogAgent()

	// Detect FluentBit/Fluent
	config.FluentBitEnabled = ad.detectFluentAgent()

	// Detect OpenTelemetry Collector
	config.OTelAgentEnabled = ad.detectOTelAgent()

	// Detect Filebeat
	config.FilebeatEnabled = ad.detectFilebeatAgent()

	ad.logger.Info("agent detection completed",
		slog.Bool("datadog", config.DataDogEnabled),
		slog.Bool("fluentbit", config.FluentBitEnabled),
		slog.Bool("otel_agent", config.OTelAgentEnabled),
		slog.Bool("filebeat", config.FilebeatEnabled))

	return config
}

func (ad *AgentDetector) detectDataDogAgent() bool {
	// Check for DataDog environment variables
	if os.Getenv("DD_API_KEY") != "" || os.Getenv("DATADOG_API_KEY") != "" {
		return true
	}

	// Check for DataDog agent process or configuration
	if os.Getenv("DD_AGENT_HOST") != "" || os.Getenv("DD_TRACE_AGENT_URL") != "" {
		return true
	}

	return false
}

func (ad *AgentDetector) detectFluentAgent() bool {
	// Check for Fluent environment variables
	if os.Getenv("FLUENTD_HOST") != "" || os.Getenv("FLUENT_HOST") != "" {
		return true
	}

	// Check for FluentBit configuration
	if os.Getenv("FLUENT_CONF") != "" || os.Getenv("FLUENTBIT_CONFIG") != "" {
		return true
	}

	return false
}

func (ad *AgentDetector) detectOTelAgent() bool {
	// Check for OpenTelemetry Collector environment variables
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		return true
	}

	if os.Getenv("OTEL_COLLECTOR_HOST") != "" {
		return true
	}

	return false
}

func (ad *AgentDetector) detectFilebeatAgent() bool {
	// Check for Filebeat environment variables
	if os.Getenv("FILEBEAT_CONFIG") != "" || os.Getenv("ELASTIC_BEATS_CONFIG") != "" {
		return true
	}

	// Check for Elastic Cloud configuration
	if os.Getenv("ELASTIC_CLOUD_ID") != "" || os.Getenv("ELASTIC_CLOUD_AUTH") != "" {
		return true
	}

	return false
}
