// Package observability provides a comprehensive observability system for the elevator control system
// supporting multiple backends, agents, and both pull/push patterns using OpenTelemetry standards.
package observability

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// ObservabilityConfig contains configuration for all observability components
type ObservabilityConfig struct {
	// Core settings
	Enabled     bool   `env:"OBSERVABILITY_ENABLED" envDefault:"true"`
	ServiceName string `env:"SERVICE_NAME" envDefault:"elevator-control-system"`
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
	Version     string `env:"SERVICE_VERSION" envDefault:"1.0.0"`

	// OpenTelemetry settings
	OTel OTelConfig `envPrefix:"OTEL_"`

	// Metrics configuration
	Metrics MetricsConfig `envPrefix:"METRICS_"`

	// Logging configuration
	Logging LoggingConfig `envPrefix:"LOGGING_"`

	// Tracing configuration
	Tracing TracingConfig `envPrefix:"TRACING_"`

	// External agents detection
	Agents AgentConfig `envPrefix:"AGENTS_"`

	// Platform-specific configurations
	DataDog    DataDogConfig    `envPrefix:"DATADOG_"`
	Prometheus PrometheusConfig `envPrefix:"PROMETHEUS_"`
	Elastic    ElasticConfig    `envPrefix:"ELASTIC_"`
	OTLP       OTLPConfig       `envPrefix:"OTLP_"`
}

// OTelConfig contains OpenTelemetry-specific configuration
type OTelConfig struct {
	Enabled            bool          `env:"ENABLED" envDefault:"true"`
	ExporterType       string        `env:"EXPORTER_TYPE" envDefault:"auto"`        // auto, otlp, prometheus, datadog, jaeger
	ExporterEndpoint   string        `env:"EXPORTER_ENDPOINT" envDefault:""`        // OTLP endpoint
	ExporterTimeout    time.Duration `env:"EXPORTER_TIMEOUT" envDefault:"10s"`      // Export timeout
	BatchTimeout       time.Duration `env:"BATCH_TIMEOUT" envDefault:"5s"`          // Batch export timeout
	MaxExportBatchSize int           `env:"MAX_EXPORT_BATCH_SIZE" envDefault:"512"` // Max batch size
	ExportInterval     time.Duration `env:"EXPORT_INTERVAL" envDefault:"5s"`        // Export interval
	ResourceAttributes string        `env:"RESOURCE_ATTRIBUTES" envDefault:""`      // Additional resource attributes
	Headers            string        `env:"HEADERS" envDefault:""`                  // Additional headers for OTLP
	Insecure           bool          `env:"INSECURE" envDefault:"false"`            // Use insecure connection
	Compression        string        `env:"COMPRESSION" envDefault:"gzip"`          // Compression: gzip, none
	SamplingRatio      float64       `env:"SAMPLING_RATIO" envDefault:"1.0"`        // Trace sampling ratio (0.0-1.0)
}

// MetricsConfig contains metrics-specific configuration
type MetricsConfig struct {
	Enabled         bool          `env:"ENABLED" envDefault:"true"`
	PushEnabled     bool          `env:"PUSH_ENABLED" envDefault:"false"`    // Enable push-based metrics
	PushInterval    time.Duration `env:"PUSH_INTERVAL" envDefault:"15s"`     // Push interval
	PullEnabled     bool          `env:"PULL_ENABLED" envDefault:"true"`     // Enable pull-based metrics (Prometheus)
	Port            int           `env:"PORT" envDefault:"8080"`             // Metrics server port
	Path            string        `env:"PATH" envDefault:"/metrics"`         // Metrics endpoint path
	Namespace       string        `env:"NAMESPACE" envDefault:"elevator"`    // Metrics namespace
	CustomLabels    string        `env:"CUSTOM_LABELS" envDefault:""`        // Additional custom labels (key1=value1,key2=value2)
	HistogramBounds string        `env:"HISTOGRAM_BOUNDS" envDefault:""`     // Custom histogram bounds
	DisableBuiltIn  bool          `env:"DISABLE_BUILTIN" envDefault:"false"` // Disable built-in metrics
}

// LoggingConfig contains logging-specific configuration
type LoggingConfig struct {
	Enabled         bool          `env:"ENABLED" envDefault:"true"`
	Level           string        `env:"LEVEL" envDefault:"info"`                      // debug, info, warn, error
	Format          string        `env:"FORMAT" envDefault:"json"`                     // json, text, console
	Output          string        `env:"OUTPUT" envDefault:"stdout"`                   // stdout, stderr, file, agent
	FilePath        string        `env:"FILE_PATH" envDefault:"/var/log/elevator.log"` // Log file path when output=file
	MaxSize         int           `env:"MAX_SIZE" envDefault:"100"`                    // Max log file size in MB
	MaxBackups      int           `env:"MAX_BACKUPS" envDefault:"3"`                   // Max number of backup files
	MaxAge          int           `env:"MAX_AGE" envDefault:"28"`                      // Max age in days
	Compress        bool          `env:"COMPRESS" envDefault:"true"`                   // Compress old log files
	AddSource       bool          `env:"ADD_SOURCE" envDefault:"false"`                // Add source code location
	SampleRate      int           `env:"SAMPLE_RATE" envDefault:"1"`                   // Log sampling rate (1 = no sampling)
	BufferSize      int           `env:"BUFFER_SIZE" envDefault:"1000"`                // Log buffer size
	FlushInterval   time.Duration `env:"FLUSH_INTERVAL" envDefault:"5s"`               // Log flush interval
	StructuredExtra string        `env:"STRUCTURED_EXTRA" envDefault:""`               // Extra structured fields (key1=value1,key2=value2)
}

// TracingConfig contains tracing-specific configuration
type TracingConfig struct {
	Enabled       bool          `env:"ENABLED" envDefault:"true"`
	SamplingRatio float64       `env:"SAMPLING_RATIO" envDefault:"1.0"` // Sampling ratio (0.0-1.0)
	MaxTagLength  int           `env:"MAX_TAG_LENGTH" envDefault:"256"` // Maximum tag value length
	MaxLogs       int           `env:"MAX_LOGS" envDefault:"10"`        // Maximum logs per span
	MaxAttributes int           `env:"MAX_ATTRIBUTES" envDefault:"64"`  // Maximum attributes per span
	Timeout       time.Duration `env:"TIMEOUT" envDefault:"10s"`        // Span export timeout
	BatchSize     int           `env:"BATCH_SIZE" envDefault:"128"`     // Batch export size
}

// AgentConfig contains configuration for external agent detection
type AgentConfig struct {
	AutoDetect        bool `env:"AUTO_DETECT" envDefault:"true"`         // Auto-detect agents
	FluentBitEnabled  bool `env:"FLUENTBIT_ENABLED" envDefault:"false"`  // FluentBit agent
	DataDogEnabled    bool `env:"DATADOG_ENABLED" envDefault:"false"`    // DataDog agent
	OTelAgentEnabled  bool `env:"OTEL_AGENT_ENABLED" envDefault:"false"` // OpenTelemetry Collector
	FilebeatEnabled   bool `env:"FILEBEAT_ENABLED" envDefault:"false"`   // Filebeat agent
	PrometheusEnabled bool `env:"PROMETHEUS_ENABLED" envDefault:"true"`  // Prometheus scraping
	FluentBitPort     int  `env:"FLUENTBIT_PORT" envDefault:"24224"`     // FluentBit forward port
	DataDogPort       int  `env:"DATADOG_PORT" envDefault:"8125"`        // DataDog StatsD port
	OTelAgentPort     int  `env:"OTEL_AGENT_PORT" envDefault:"4317"`     // OTLP gRPC port
}

// DataDogConfig contains DataDog-specific configuration
type DataDogConfig struct {
	Enabled     bool   `env:"ENABLED" envDefault:"false"`
	APIKey      string `env:"API_KEY" envDefault:""`           // DataDog API key
	Site        string `env:"SITE" envDefault:"datadoghq.com"` // DataDog site (datadoghq.com, datadoghq.eu, etc.)
	Host        string `env:"HOST" envDefault:"localhost"`     // DataDog agent host
	Port        int    `env:"PORT" envDefault:"8125"`          // DataDog agent port
	Namespace   string `env:"NAMESPACE" envDefault:"elevator"` // Metrics namespace
	Tags        string `env:"TAGS" envDefault:""`              // Additional tags (key1:value1,key2:value2)
	APMEnabled  bool   `env:"APM_ENABLED" envDefault:"false"`  // Enable APM tracing
	APMHost     string `env:"APM_HOST" envDefault:"localhost"` // APM agent host
	APMPort     int    `env:"APM_PORT" envDefault:"8126"`      // APM agent port
	LogEnabled  bool   `env:"LOG_ENABLED" envDefault:"false"`  // Enable log forwarding
	LogEndpoint string `env:"LOG_ENDPOINT" envDefault:""`      // Log intake endpoint
}

// PrometheusConfig contains Prometheus-specific configuration
type PrometheusConfig struct {
	Enabled       bool          `env:"ENABLED" envDefault:"true"`
	PushEnabled   bool          `env:"PUSH_ENABLED" envDefault:"false"`        // Enable push gateway
	PushGateway   string        `env:"PUSH_GATEWAY" envDefault:""`             // Push gateway URL
	PushInterval  time.Duration `env:"PUSH_INTERVAL" envDefault:"15s"`         // Push interval
	PushJob       string        `env:"PUSH_JOB" envDefault:"elevator-metrics"` // Push job name
	PushTimeout   time.Duration `env:"PUSH_TIMEOUT" envDefault:"10s"`          // Push timeout
	ExtraLabels   string        `env:"EXTRA_LABELS" envDefault:""`             // Extra labels for all metrics
	ScrapePort    int           `env:"SCRAPE_PORT" envDefault:"8080"`          // Port for Prometheus scraping
	ScrapePath    string        `env:"SCRAPE_PATH" envDefault:"/metrics"`      // Path for Prometheus scraping
	ScrapeTimeout time.Duration `env:"SCRAPE_TIMEOUT" envDefault:"10s"`        // Scrape timeout
}

// ElasticConfig contains Elasticsearch/ELK stack configuration
type ElasticConfig struct {
	Enabled        bool          `env:"ENABLED" envDefault:"false"`
	Host           string        `env:"HOST" envDefault:"localhost"`        // Elasticsearch host
	Port           int           `env:"PORT" envDefault:"9200"`             // Elasticsearch port
	Username       string        `env:"USERNAME" envDefault:""`             // Authentication username
	Password       string        `env:"PASSWORD" envDefault:""`             // Authentication password
	Index          string        `env:"INDEX" envDefault:"elevator-logs"`   // Log index pattern
	IndexRotation  string        `env:"INDEX_ROTATION" envDefault:"daily"`  // daily, weekly, monthly
	BulkSize       int           `env:"BULK_SIZE" envDefault:"100"`         // Bulk insert size
	FlushInterval  time.Duration `env:"FLUSH_INTERVAL" envDefault:"5s"`     // Bulk flush interval
	Timeout        time.Duration `env:"TIMEOUT" envDefault:"30s"`           // Request timeout
	TLS            bool          `env:"TLS" envDefault:"false"`             // Enable TLS
	TLSSkipVerify  bool          `env:"TLS_SKIP_VERIFY" envDefault:"false"` // Skip TLS verification
	LogsEnabled    bool          `env:"LOGS_ENABLED" envDefault:"true"`     // Send logs to Elasticsearch
	MetricsEnabled bool          `env:"METRICS_ENABLED" envDefault:"false"` // Send metrics to Elasticsearch
	TracesEnabled  bool          `env:"TRACES_ENABLED" envDefault:"false"`  // Send traces to Elasticsearch
}

// OTLPConfig contains OpenTelemetry Protocol configuration
type OTLPConfig struct {
	Enabled      bool          `env:"ENABLED" envDefault:"false"`
	Endpoint     string        `env:"ENDPOINT" envDefault:"http://localhost:4317"` // OTLP gRPC endpoint
	HTTPEndpoint string        `env:"HTTP_ENDPOINT" envDefault:""`                 // OTLP HTTP endpoint
	Insecure     bool          `env:"INSECURE" envDefault:"true"`                  // Use insecure connection
	Headers      string        `env:"HEADERS" envDefault:""`                       // Additional headers
	Timeout      time.Duration `env:"TIMEOUT" envDefault:"10s"`                    // Request timeout
	Compression  string        `env:"COMPRESSION" envDefault:"gzip"`               // Compression: gzip, none
	TLS          bool          `env:"TLS" envDefault:"false"`                      // Enable TLS
	TLSCert      string        `env:"TLS_CERT" envDefault:""`                      // TLS certificate path
	TLSKey       string        `env:"TLS_KEY" envDefault:""`                       // TLS key path
	TLSCA        string        `env:"TLS_CA" envDefault:""`                        // TLS CA certificate path
}

// LoadObservabilityConfig loads observability configuration from environment variables
func LoadObservabilityConfig() (*ObservabilityConfig, error) {
	config := &ObservabilityConfig{
		// Set defaults
		Enabled:     getBoolEnv("OBSERVABILITY_ENABLED", true),
		ServiceName: getStringEnv("SERVICE_NAME", "elevator-control-system"),
		Environment: getStringEnv("ENVIRONMENT", "development"),
		Version:     getStringEnv("SERVICE_VERSION", "1.0.0"),
	}

	// Load OpenTelemetry configuration
	if err := loadOTelConfig(&config.OTel); err != nil {
		return nil, fmt.Errorf("failed to load OpenTelemetry config: %w", err)
	}

	// Load metrics configuration
	if err := loadMetricsConfig(&config.Metrics); err != nil {
		return nil, fmt.Errorf("failed to load metrics config: %w", err)
	}

	// Load logging configuration
	if err := loadLoggingConfig(&config.Logging); err != nil {
		return nil, fmt.Errorf("failed to load logging config: %w", err)
	}

	// Load tracing configuration
	if err := loadTracingConfig(&config.Tracing); err != nil {
		return nil, fmt.Errorf("failed to load tracing config: %w", err)
	}

	// Load agents configuration
	if err := loadAgentsConfig(&config.Agents); err != nil {
		return nil, fmt.Errorf("failed to load agents config: %w", err)
	}

	// Load platform-specific configurations
	if err := loadDataDogConfig(&config.DataDog); err != nil {
		return nil, fmt.Errorf("failed to load DataDog config: %w", err)
	}

	if err := loadPrometheusConfig(&config.Prometheus); err != nil {
		return nil, fmt.Errorf("failed to load Prometheus config: %w", err)
	}

	if err := loadElasticConfig(&config.Elastic); err != nil {
		return nil, fmt.Errorf("failed to load Elastic config: %w", err)
	}

	if err := loadOTLPConfig(&config.OTLP); err != nil {
		return nil, fmt.Errorf("failed to load OTLP config: %w", err)
	}

	// Auto-detect agents if enabled
	if config.Agents.AutoDetect {
		autoDetectAgents(&config.Agents)
	}

	// Apply intelligent defaults based on detected agents
	applyIntelligentDefaults(config)

	return config, nil
}

// Helper functions for loading specific configurations
func loadOTelConfig(cfg *OTelConfig) error {
	cfg.Enabled = getBoolEnv("OTEL_ENABLED", true)
	cfg.ExporterType = getStringEnv("OTEL_EXPORTER_TYPE", "auto")
	cfg.ExporterEndpoint = getStringEnv("OTEL_EXPORTER_ENDPOINT", "")
	cfg.ExporterTimeout = getDurationEnv("OTEL_EXPORTER_TIMEOUT", 10*time.Second)
	cfg.BatchTimeout = getDurationEnv("OTEL_BATCH_TIMEOUT", 5*time.Second)
	cfg.MaxExportBatchSize = getIntEnv("OTEL_MAX_EXPORT_BATCH_SIZE", 512)
	cfg.ExportInterval = getDurationEnv("OTEL_EXPORT_INTERVAL", 5*time.Second)
	cfg.ResourceAttributes = getStringEnv("OTEL_RESOURCE_ATTRIBUTES", "")
	cfg.Headers = getStringEnv("OTEL_HEADERS", "")
	cfg.Insecure = getBoolEnv("OTEL_INSECURE", false)
	cfg.Compression = getStringEnv("OTEL_COMPRESSION", "gzip")
	cfg.SamplingRatio = getFloat64Env("OTEL_SAMPLING_RATIO", 1.0)
	return nil
}

func loadMetricsConfig(cfg *MetricsConfig) error {
	cfg.Enabled = getBoolEnv("METRICS_ENABLED", true)
	cfg.PushEnabled = getBoolEnv("METRICS_PUSH_ENABLED", false)
	cfg.PushInterval = getDurationEnv("METRICS_PUSH_INTERVAL", 15*time.Second)
	cfg.PullEnabled = getBoolEnv("METRICS_PULL_ENABLED", true)
	cfg.Port = getIntEnv("METRICS_PORT", 8080)
	cfg.Path = getStringEnv("METRICS_PATH", "/metrics")
	cfg.Namespace = getStringEnv("METRICS_NAMESPACE", "elevator")
	cfg.CustomLabels = getStringEnv("METRICS_CUSTOM_LABELS", "")
	cfg.HistogramBounds = getStringEnv("METRICS_HISTOGRAM_BOUNDS", "")
	cfg.DisableBuiltIn = getBoolEnv("METRICS_DISABLE_BUILTIN", false)
	return nil
}

func loadLoggingConfig(cfg *LoggingConfig) error {
	cfg.Enabled = getBoolEnv("LOGGING_ENABLED", true)
	cfg.Level = getStringEnv("LOGGING_LEVEL", "info")
	cfg.Format = getStringEnv("LOGGING_FORMAT", "json")
	cfg.Output = getStringEnv("LOGGING_OUTPUT", "stdout")
	cfg.FilePath = getStringEnv("LOGGING_FILE_PATH", "/var/log/elevator.log")
	cfg.MaxSize = getIntEnv("LOGGING_MAX_SIZE", 100)
	cfg.MaxBackups = getIntEnv("LOGGING_MAX_BACKUPS", 3)
	cfg.MaxAge = getIntEnv("LOGGING_MAX_AGE", 28)
	cfg.Compress = getBoolEnv("LOGGING_COMPRESS", true)
	cfg.AddSource = getBoolEnv("LOGGING_ADD_SOURCE", false)
	cfg.SampleRate = getIntEnv("LOGGING_SAMPLE_RATE", 1)
	cfg.BufferSize = getIntEnv("LOGGING_BUFFER_SIZE", 1000)
	cfg.FlushInterval = getDurationEnv("LOGGING_FLUSH_INTERVAL", 5*time.Second)
	cfg.StructuredExtra = getStringEnv("LOGGING_STRUCTURED_EXTRA", "")
	return nil
}

func loadTracingConfig(cfg *TracingConfig) error {
	cfg.Enabled = getBoolEnv("TRACING_ENABLED", true)
	cfg.SamplingRatio = getFloat64Env("TRACING_SAMPLING_RATIO", 1.0)
	cfg.MaxTagLength = getIntEnv("TRACING_MAX_TAG_LENGTH", 256)
	cfg.MaxLogs = getIntEnv("TRACING_MAX_LOGS", 10)
	cfg.MaxAttributes = getIntEnv("TRACING_MAX_ATTRIBUTES", 64)
	cfg.Timeout = getDurationEnv("TRACING_TIMEOUT", 10*time.Second)
	cfg.BatchSize = getIntEnv("TRACING_BATCH_SIZE", 128)
	return nil
}

func loadAgentsConfig(cfg *AgentConfig) error {
	cfg.AutoDetect = getBoolEnv("AGENTS_AUTO_DETECT", true)
	cfg.FluentBitEnabled = getBoolEnv("AGENTS_FLUENTBIT_ENABLED", false)
	cfg.DataDogEnabled = getBoolEnv("AGENTS_DATADOG_ENABLED", false)
	cfg.OTelAgentEnabled = getBoolEnv("AGENTS_OTEL_AGENT_ENABLED", false)
	cfg.FilebeatEnabled = getBoolEnv("AGENTS_FILEBEAT_ENABLED", false)
	cfg.PrometheusEnabled = getBoolEnv("AGENTS_PROMETHEUS_ENABLED", true)
	cfg.FluentBitPort = getIntEnv("AGENTS_FLUENTBIT_PORT", 24224)
	cfg.DataDogPort = getIntEnv("AGENTS_DATADOG_PORT", 8125)
	cfg.OTelAgentPort = getIntEnv("AGENTS_OTEL_AGENT_PORT", 4317)
	return nil
}

func loadDataDogConfig(cfg *DataDogConfig) error {
	cfg.Enabled = getBoolEnv("DATADOG_ENABLED", false)
	cfg.APIKey = getStringEnv("DATADOG_API_KEY", "")
	cfg.Site = getStringEnv("DATADOG_SITE", "datadoghq.com")
	cfg.Host = getStringEnv("DATADOG_HOST", "localhost")
	cfg.Port = getIntEnv("DATADOG_PORT", 8125)
	cfg.Namespace = getStringEnv("DATADOG_NAMESPACE", "elevator")
	cfg.Tags = getStringEnv("DATADOG_TAGS", "")
	cfg.APMEnabled = getBoolEnv("DATADOG_APM_ENABLED", false)
	cfg.APMHost = getStringEnv("DATADOG_APM_HOST", "localhost")
	cfg.APMPort = getIntEnv("DATADOG_APM_PORT", 8126)
	cfg.LogEnabled = getBoolEnv("DATADOG_LOG_ENABLED", false)
	cfg.LogEndpoint = getStringEnv("DATADOG_LOG_ENDPOINT", "")
	return nil
}

func loadPrometheusConfig(cfg *PrometheusConfig) error {
	cfg.Enabled = getBoolEnv("PROMETHEUS_ENABLED", true)
	cfg.PushEnabled = getBoolEnv("PROMETHEUS_PUSH_ENABLED", false)
	cfg.PushGateway = getStringEnv("PROMETHEUS_PUSH_GATEWAY", "")
	cfg.PushInterval = getDurationEnv("PROMETHEUS_PUSH_INTERVAL", 15*time.Second)
	cfg.PushJob = getStringEnv("PROMETHEUS_PUSH_JOB", "elevator-metrics")
	cfg.PushTimeout = getDurationEnv("PROMETHEUS_PUSH_TIMEOUT", 10*time.Second)
	cfg.ExtraLabels = getStringEnv("PROMETHEUS_EXTRA_LABELS", "")
	cfg.ScrapePort = getIntEnv("PROMETHEUS_SCRAPE_PORT", 8080)
	cfg.ScrapePath = getStringEnv("PROMETHEUS_SCRAPE_PATH", "/metrics")
	cfg.ScrapeTimeout = getDurationEnv("PROMETHEUS_SCRAPE_TIMEOUT", 10*time.Second)
	return nil
}

func loadElasticConfig(cfg *ElasticConfig) error {
	cfg.Enabled = getBoolEnv("ELASTIC_ENABLED", false)
	cfg.Host = getStringEnv("ELASTIC_HOST", "localhost")
	cfg.Port = getIntEnv("ELASTIC_PORT", 9200)
	cfg.Username = getStringEnv("ELASTIC_USERNAME", "")
	cfg.Password = getStringEnv("ELASTIC_PASSWORD", "")
	cfg.Index = getStringEnv("ELASTIC_INDEX", "elevator-logs")
	cfg.IndexRotation = getStringEnv("ELASTIC_INDEX_ROTATION", "daily")
	cfg.BulkSize = getIntEnv("ELASTIC_BULK_SIZE", 100)
	cfg.FlushInterval = getDurationEnv("ELASTIC_FLUSH_INTERVAL", 5*time.Second)
	cfg.Timeout = getDurationEnv("ELASTIC_TIMEOUT", 30*time.Second)
	cfg.TLS = getBoolEnv("ELASTIC_TLS", false)
	cfg.TLSSkipVerify = getBoolEnv("ELASTIC_TLS_SKIP_VERIFY", false)
	cfg.LogsEnabled = getBoolEnv("ELASTIC_LOGS_ENABLED", true)
	cfg.MetricsEnabled = getBoolEnv("ELASTIC_METRICS_ENABLED", false)
	cfg.TracesEnabled = getBoolEnv("ELASTIC_TRACES_ENABLED", false)
	return nil
}

func loadOTLPConfig(cfg *OTLPConfig) error {
	cfg.Enabled = getBoolEnv("OTLP_ENABLED", false)
	cfg.Endpoint = getStringEnv("OTLP_ENDPOINT", "http://localhost:4317")
	cfg.HTTPEndpoint = getStringEnv("OTLP_HTTP_ENDPOINT", "")
	cfg.Insecure = getBoolEnv("OTLP_INSECURE", true)
	cfg.Headers = getStringEnv("OTLP_HEADERS", "")
	cfg.Timeout = getDurationEnv("OTLP_TIMEOUT", 10*time.Second)
	cfg.Compression = getStringEnv("OTLP_COMPRESSION", "gzip")
	cfg.TLS = getBoolEnv("OTLP_TLS", false)
	cfg.TLSCert = getStringEnv("OTLP_TLS_CERT", "")
	cfg.TLSKey = getStringEnv("OTLP_TLS_KEY", "")
	cfg.TLSCA = getStringEnv("OTLP_TLS_CA", "")
	return nil
}

// Utility functions for environment variable parsing
func getStringEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getFloat64Env(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// autoDetectAgents detects available observability agents in the environment
func autoDetectAgents(cfg *AgentConfig) {
	// Auto-detect DataDog agent
	if !cfg.DataDogEnabled {
		cfg.DataDogEnabled = os.Getenv("DD_API_KEY") != "" ||
			os.Getenv("DATADOG_API_KEY") != "" ||
			os.Getenv("DD_AGENT_HOST") != ""
	}

	// Auto-detect OpenTelemetry Collector
	if !cfg.OTelAgentEnabled {
		cfg.OTelAgentEnabled = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" ||
			os.Getenv("OTEL_COLLECTOR_HOST") != ""
	}

	// Auto-detect FluentBit by checking environment variables
	if !cfg.FluentBitEnabled {
		cfg.FluentBitEnabled = os.Getenv("FLUENTD_HOST") != "" ||
			os.Getenv("FLUENT_HOST") != "" ||
			os.Getenv("FLUENTBIT_CONFIG") != ""
	}

	// Auto-detect Filebeat by checking common environment variables
	if !cfg.FilebeatEnabled {
		cfg.FilebeatEnabled = os.Getenv("FILEBEAT_CONFIG") != "" ||
			os.Getenv("ELASTIC_BEATS_CONFIG") != "" ||
			os.Getenv("ELASTIC_CLOUD_ID") != ""
	}
}

// applyIntelligentDefaults applies intelligent defaults based on detected configuration
func applyIntelligentDefaults(config *ObservabilityConfig) {
	// If DataDog agent is detected, configure DataDog-specific settings
	if config.Agents.DataDogEnabled && !config.DataDog.Enabled {
		config.DataDog.Enabled = true
		config.OTel.ExporterType = "datadog"
		config.Metrics.PushEnabled = true
		config.Tracing.Enabled = true
	}

	// If OTLP collector is detected, configure OTLP settings
	if config.Agents.OTelAgentEnabled && !config.OTLP.Enabled {
		config.OTLP.Enabled = true
		config.OTel.ExporterType = "otlp"
		config.Metrics.PushEnabled = true
	}

	// If FluentBit is detected, configure structured logging
	if config.Agents.FluentBitEnabled {
		config.Logging.Format = "json"
		config.Logging.Output = "stdout" // FluentBit will collect from stdout
	}

	// If Elastic is enabled, configure appropriate settings
	if config.Elastic.Enabled {
		config.Logging.Format = "json"
		config.OTel.ExporterType = "otlp" // Use OTLP to send to Elastic APM
	}

	// Auto-configure exporter type if set to "auto"
	if config.OTel.ExporterType == "auto" {
		switch {
		case config.DataDog.Enabled:
			config.OTel.ExporterType = "datadog"
		case config.OTLP.Enabled:
			config.OTel.ExporterType = "otlp"
		case config.Prometheus.Enabled:
			config.OTel.ExporterType = "prometheus"
		default:
			config.OTel.ExporterType = "prometheus" // Default fallback
		}
	}
}

// GetActiveExporters returns a list of active exporters based on configuration
func (c *ObservabilityConfig) GetActiveExporters() []string {
	var exporters []string

	if c.Prometheus.Enabled {
		exporters = append(exporters, "prometheus")
	}
	if c.DataDog.Enabled {
		exporters = append(exporters, "datadog")
	}
	if c.OTLP.Enabled {
		exporters = append(exporters, "otlp")
	}
	if c.Elastic.Enabled {
		exporters = append(exporters, "elastic")
	}

	return exporters
}

// GetResourceAttributes returns OpenTelemetry resource attributes
func (c *ObservabilityConfig) GetResourceAttributes() map[string]string {
	attrs := map[string]string{
		"service.name":           c.ServiceName,
		"service.version":        c.Version,
		"deployment.environment": c.Environment,
	}

	// Parse additional resource attributes from configuration
	if c.OTel.ResourceAttributes != "" {
		pairs := strings.Split(c.OTel.ResourceAttributes, ",")
		for _, pair := range pairs {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				attrs[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	return attrs
}

// GetCustomLabels returns custom labels as a map
func (c *MetricsConfig) GetCustomLabels() map[string]string {
	labels := make(map[string]string)

	if c.CustomLabels != "" {
		pairs := strings.Split(c.CustomLabels, ",")
		for _, pair := range pairs {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				labels[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	return labels
}

// Validate validates the observability configuration
func (c *ObservabilityConfig) Validate() error {
	if !c.Enabled {
		return nil // Skip validation if observability is disabled
	}

	// Validate service name
	if c.ServiceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	// Validate OTLP configuration
	if c.OTLP.Enabled && c.OTLP.Endpoint == "" {
		return fmt.Errorf("OTLP endpoint must be specified when OTLP is enabled")
	}

	// Validate DataDog configuration
	if c.DataDog.Enabled && c.DataDog.APIKey == "" {
		return fmt.Errorf("DataDog API key must be specified when DataDog is enabled")
	}

	// Validate Elasticsearch configuration
	if c.Elastic.Enabled && c.Elastic.Host == "" {
		return fmt.Errorf("Elasticsearch host must be specified when Elastic is enabled")
	}

	// Validate sampling ratios
	if c.OTel.SamplingRatio < 0.0 || c.OTel.SamplingRatio > 1.0 {
		return fmt.Errorf("OpenTelemetry sampling ratio must be between 0.0 and 1.0")
	}

	if c.Tracing.SamplingRatio < 0.0 || c.Tracing.SamplingRatio > 1.0 {
		return fmt.Errorf("tracing sampling ratio must be between 0.0 and 1.0")
	}

	return nil
}
