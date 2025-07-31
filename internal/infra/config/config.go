package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env"
	"github.com/slavakukuyev/elevator-go/internal/constants"
	"github.com/slavakukuyev/elevator-go/internal/domain"
)

// Config represents the application configuration with comprehensive options
type Config struct {
	// Environment and basic settings
	Environment string `env:"ENV" envDefault:"development"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"INFO"`

	// Server configuration
	Port            int           `env:"PORT" envDefault:"6660"`
	ReadTimeout     time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"30s"`
	WriteTimeout    time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"30s"`
	IdleTimeout     time.Duration `env:"SERVER_IDLE_TIMEOUT" envDefault:"120s"`
	ShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT" envDefault:"30s"`
	ShutdownGrace   time.Duration `env:"SERVER_SHUTDOWN_GRACE" envDefault:"2s"`

	// Elevator system configuration
	MaxFloor                 int           `env:"DEFAULT_MAX_FLOOR" envDefault:"9"`
	MinFloor                 int           `env:"DEFAULT_MIN_FLOOR" envDefault:"0"`
	DefaultOverloadThreshold int           `env:"DEFAULT_OVERLOAD_THRESHOLD" envDefault:"12"`
	EachFloorDuration        time.Duration `env:"EACH_FLOOR_DURATION" envDefault:"500ms"`
	OpenDoorDuration         time.Duration `env:"OPEN_DOOR_DURATION" envDefault:"2s"`
	OperationTimeout         time.Duration `env:"ELEVATOR_OPERATION_TIMEOUT" envDefault:"30s"`
	CreateElevatorTimeout    time.Duration `env:"CREATE_ELEVATOR_TIMEOUT" envDefault:"10s"`
	RequestTimeout           time.Duration `env:"ELEVATOR_REQUEST_TIMEOUT" envDefault:"5s"`
	StatusUpdateTimeout      time.Duration `env:"STATUS_UPDATE_TIMEOUT" envDefault:"3s"`
	HealthCheckTimeout       time.Duration `env:"HEALTH_CHECK_TIMEOUT" envDefault:"2s"`
	MaxElevators             int           `env:"MAX_ELEVATORS" envDefault:"100"`
	DefaultElevatorCount     int           `env:"DEFAULT_ELEVATOR_COUNT" envDefault:"0"`
	NamePrefix               string        `env:"ELEVATOR_NAME_PREFIX" envDefault:"Elevator"`
	SwitchOnChannelBuffer    int           `env:"SWITCH_ON_CHANNEL_BUFFER" envDefault:"10"`

	// HTTP Configuration
	RateLimitRPM       int           `env:"RATE_LIMIT_RPM" envDefault:"100"`
	RateLimitWindow    time.Duration `env:"RATE_LIMIT_WINDOW" envDefault:"1m"`
	RateLimitCleanup   time.Duration `env:"RATE_LIMIT_CLEANUP" envDefault:"5m"`
	MaxRequestSize     int64         `env:"MAX_REQUEST_SIZE" envDefault:"1048576"`
	RequestTimeoutHTTP time.Duration `env:"HTTP_REQUEST_TIMEOUT" envDefault:"30s"`
	CORSEnabled        bool          `env:"CORS_ENABLED" envDefault:"true"`
	CORSMaxAge         time.Duration `env:"CORS_MAX_AGE" envDefault:"12h"`
	CORSAllowedOrigins string        `env:"CORS_ALLOWED_ORIGINS" envDefault:"*"`

	// Monitoring
	MetricsEnabled       bool          `env:"METRICS_ENABLED" envDefault:"true"`
	MetricsPath          string        `env:"METRICS_PATH" envDefault:"/metrics"`
	StatusUpdateInterval time.Duration `env:"STATUS_UPDATE_INTERVAL" envDefault:"1s"`
	HealthEnabled        bool          `env:"HEALTH_ENABLED" envDefault:"true"`
	HealthPath           string        `env:"HEALTH_PATH" envDefault:"/health"`
	StructuredLogging    bool          `env:"STRUCTURED_LOGGING" envDefault:"true"`
	LogRequestDetails    bool          `env:"LOG_REQUEST_DETAILS" envDefault:"false"`
	CorrelationIDHeader  string        `env:"CORRELATION_ID_HEADER" envDefault:"X-Request-ID"`

	// Circuit Breaker
	CircuitBreakerEnabled          bool          `env:"CIRCUIT_BREAKER_ENABLED" envDefault:"true"`
	CircuitBreakerMaxFailures      int           `env:"CIRCUIT_BREAKER_MAX_FAILURES" envDefault:"5"`
	CircuitBreakerResetTimeout     time.Duration `env:"CIRCUIT_BREAKER_RESET_TIMEOUT" envDefault:"30s"`
	CircuitBreakerHalfOpenLimit    int           `env:"CIRCUIT_BREAKER_HALF_OPEN_LIMIT" envDefault:"3"`
	CircuitBreakerFailureThreshold float64       `env:"CIRCUIT_BREAKER_FAILURE_THRESHOLD" envDefault:"0.6"`

	// WebSocket
	WebSocketEnabled           bool          `env:"WEBSOCKET_ENABLED" envDefault:"true"`
	WebSocketPath              string        `env:"WEBSOCKET_PATH" envDefault:"/ws/status"`
	WebSocketConnectionTimeout time.Duration `env:"WEBSOCKET_CONNECTION_TIMEOUT" envDefault:"10m"`
	WebSocketWriteTimeout      time.Duration `env:"WEBSOCKET_WRITE_TIMEOUT" envDefault:"5s"`
	WebSocketReadTimeout       time.Duration `env:"WEBSOCKET_READ_TIMEOUT" envDefault:"60s"`
	WebSocketPingInterval      time.Duration `env:"WEBSOCKET_PING_INTERVAL" envDefault:"30s"`
	WebSocketMaxConnections    int           `env:"WEBSOCKET_MAX_CONNECTIONS" envDefault:"1000"`
	WebSocketBufferSize        int           `env:"WEBSOCKET_BUFFER_SIZE" envDefault:"1024"`
}

// ServerConfig contains HTTP server specific configuration
type ServerConfig struct {
	Port         int           `env:"PORT" envDefault:"6660"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT" envDefault:"30s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" envDefault:"30s"`
	IdleTimeout  time.Duration `env:"IDLE_TIMEOUT" envDefault:"120s"`

	// Graceful shutdown configuration
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"30s"`
	ShutdownGrace   time.Duration `env:"SHUTDOWN_GRACE" envDefault:"2s"`
}

// ElevatorConfig contains elevator system specific configuration
type ElevatorConfig struct {
	// Default floor ranges
	MaxFloor                 int `env:"DEFAULT_MAX_FLOOR" envDefault:"9"`
	MinFloor                 int `env:"DEFAULT_MIN_FLOOR" envDefault:"0"`
	DefaultOverloadThreshold int `env:"DEFAULT_OVERLOAD_THRESHOLD" envDefault:"12"`

	// Timing configuration
	EachFloorDuration time.Duration `env:"EACH_FLOOR_DURATION" envDefault:"500ms"`
	OpenDoorDuration  time.Duration `env:"OPEN_DOOR_DURATION" envDefault:"2s"`

	// Operation timeouts
	OperationTimeout      time.Duration `env:"ELEVATOR_OPERATION_TIMEOUT" envDefault:"30s"`
	CreateElevatorTimeout time.Duration `env:"CREATE_ELEVATOR_TIMEOUT" envDefault:"10s"`
	RequestTimeout        time.Duration `env:"ELEVATOR_REQUEST_TIMEOUT" envDefault:"5s"`
	StatusUpdateTimeout   time.Duration `env:"STATUS_UPDATE_TIMEOUT" envDefault:"3s"`
	HealthCheckTimeout    time.Duration `env:"HEALTH_CHECK_TIMEOUT" envDefault:"2s"`

	// Elevator system limits
	MaxElevators         int    `env:"MAX_ELEVATORS" envDefault:"100"`
	DefaultElevatorCount int    `env:"DEFAULT_ELEVATOR_COUNT" envDefault:"0"`
	NamePrefix           string `env:"ELEVATOR_NAME_PREFIX" envDefault:"Elevator"`

	// Performance settings
	SwitchOnChannelBuffer int `env:"SWITCH_ON_CHANNEL_BUFFER" envDefault:"10"`
}

// HTTPConfig contains HTTP client and middleware configuration
type HTTPConfig struct {
	// Rate limiting
	RateLimitRPM     int           `env:"RATE_LIMIT_RPM" envDefault:"100"`
	RateLimitWindow  time.Duration `env:"RATE_LIMIT_WINDOW" envDefault:"1m"`
	RateLimitCleanup time.Duration `env:"RATE_LIMIT_CLEANUP" envDefault:"5m"`

	// Request processing
	MaxRequestSize int64         `env:"MAX_REQUEST_SIZE" envDefault:"1048576"` // 1MB
	RequestTimeout time.Duration `env:"HTTP_REQUEST_TIMEOUT" envDefault:"30s"`

	// CORS configuration
	CORSEnabled        bool          `env:"CORS_ENABLED" envDefault:"true"`
	CORSMaxAge         time.Duration `env:"CORS_MAX_AGE" envDefault:"12h"`
	CORSAllowedOrigins string        `env:"CORS_ALLOWED_ORIGINS" envDefault:"*"`
}

// MonitoringConfig contains monitoring and metrics configuration
type MonitoringConfig struct {
	// Metrics
	MetricsEnabled       bool          `env:"METRICS_ENABLED" envDefault:"true"`
	MetricsPath          string        `env:"METRICS_PATH" envDefault:"/metrics"`
	StatusUpdateInterval time.Duration `env:"STATUS_UPDATE_INTERVAL" envDefault:"1s"`

	// Health checks
	HealthEnabled bool   `env:"HEALTH_ENABLED" envDefault:"true"`
	HealthPath    string `env:"HEALTH_PATH" envDefault:"/health"`

	// Logging
	StructuredLogging   bool   `env:"STRUCTURED_LOGGING" envDefault:"true"`
	LogRequestDetails   bool   `env:"LOG_REQUEST_DETAILS" envDefault:"false"`
	CorrelationIDHeader string `env:"CORRELATION_ID_HEADER" envDefault:"X-Request-ID"`
}

// CircuitBreakerConfig contains circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled          bool          `env:"CIRCUIT_BREAKER_ENABLED" envDefault:"true"`
	MaxFailures      int           `env:"CIRCUIT_BREAKER_MAX_FAILURES" envDefault:"5"`
	ResetTimeout     time.Duration `env:"CIRCUIT_BREAKER_RESET_TIMEOUT" envDefault:"30s"`
	HalfOpenLimit    int           `env:"CIRCUIT_BREAKER_HALF_OPEN_LIMIT" envDefault:"3"`
	FailureThreshold float64       `env:"CIRCUIT_BREAKER_FAILURE_THRESHOLD" envDefault:"0.6"`
}

// WebSocketConfig contains WebSocket specific configuration
type WebSocketConfig struct {
	Enabled           bool          `env:"WEBSOCKET_ENABLED" envDefault:"true"`
	Path              string        `env:"WEBSOCKET_PATH" envDefault:"/ws/status"`
	ConnectionTimeout time.Duration `env:"WEBSOCKET_CONNECTION_TIMEOUT" envDefault:"10m"`
	WriteTimeout      time.Duration `env:"WEBSOCKET_WRITE_TIMEOUT" envDefault:"5s"`
	ReadTimeout       time.Duration `env:"WEBSOCKET_READ_TIMEOUT" envDefault:"60s"`
	PingInterval      time.Duration `env:"WEBSOCKET_PING_INTERVAL" envDefault:"30s"`
	MaxConnections    int           `env:"WEBSOCKET_MAX_CONNECTIONS" envDefault:"1000"`
	BufferSize        int           `env:"WEBSOCKET_BUFFER_SIZE" envDefault:"1024"`
}

// InitConfig initializes the configuration from environment variables with comprehensive validation
func InitConfig() (*Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	// Apply environment-specific defaults
	if err := applyEnvironmentDefaults(&cfg); err != nil {
		return nil, fmt.Errorf("failed to apply environment defaults: %w", err)
	}

	// Comprehensive validation
	if err := validateConfiguration(&cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// applyEnvironmentDefaults applies environment-specific default values
func applyEnvironmentDefaults(cfg *Config) error {
	switch cfg.Environment {
	case "development", "dev":
		applyDevelopmentDefaults(cfg)
	case "testing", "test":
		applyTestingDefaults(cfg)
	case "production", "prod":
		applyProductionDefaults(cfg)
	default:
		// Keep current defaults for unknown environments
	}

	return nil
}

// applyDevelopmentDefaults applies minimal changes for development (mostly defaults + debug)
func applyDevelopmentDefaults(cfg *Config) {
	// Only change: Enable debug logging for development
	if cfg.LogLevel == "INFO" {
		cfg.LogLevel = "DEBUG"
	}

	// Enable detailed request logging for debugging
	cfg.LogRequestDetails = true

	// All other values remain at their defaults for consistency
}

// applyTestingDefaults applies strict settings for testing (stricter than production)
func applyTestingDefaults(cfg *Config) {
	// Minimal logging for tests but catch errors
	cfg.LogLevel = "WARN"

	// Very fast operations for rigorous testing
	cfg.EachFloorDuration = 10 * time.Millisecond
	cfg.OpenDoorDuration = 10 * time.Millisecond

	// Very aggressive timeouts to catch timing issues early
	cfg.OperationTimeout = 500 * time.Millisecond
	cfg.CreateElevatorTimeout = 500 * time.Millisecond
	cfg.RequestTimeout = 200 * time.Millisecond
	cfg.StatusUpdateTimeout = 200 * time.Millisecond
	cfg.HealthCheckTimeout = 200 * time.Millisecond

	// Stricter HTTP timeouts than production
	cfg.ReadTimeout = 2 * time.Second
	cfg.WriteTimeout = 2 * time.Second
	cfg.IdleTimeout = 10 * time.Second
	cfg.RequestTimeoutHTTP = 1 * time.Second

	// Disable non-essential features for testing
	cfg.MetricsEnabled = false
	cfg.WebSocketEnabled = false
	cfg.LogRequestDetails = false

	// Higher rate limiting for acceptance tests but still controlled
	cfg.RateLimitRPM = 1000

	// Minimal resource limits for testing
	cfg.MaxElevators = 5
	cfg.WebSocketMaxConnections = 5
	cfg.MaxRequestSize = 256 * 1024 // 256KB

	// Very aggressive circuit breaker to catch failures fast
	cfg.CircuitBreakerMaxFailures = 1
	cfg.CircuitBreakerFailureThreshold = 0.1
	cfg.CircuitBreakerResetTimeout = 5 * time.Second
}

// applyProductionDefaults applies high-performance and strict settings for production
func applyProductionDefaults(cfg *Config) {
	// Minimal logging for performance, only critical events
	cfg.LogLevel = "WARN"
	cfg.LogRequestDetails = false

	// High-performance elevator operations
	cfg.EachFloorDuration = 200 * time.Millisecond // Faster than default for performance
	cfg.OpenDoorDuration = 1 * time.Second         // Faster door operations

	// Strict rate limiting for security
	cfg.RateLimitRPM = 30

	// Optimized timeouts for high performance
	cfg.ReadTimeout = 15 * time.Second
	cfg.WriteTimeout = 15 * time.Second
	cfg.IdleTimeout = 60 * time.Second
	cfg.RequestTimeoutHTTP = 10 * time.Second

	// Performance-oriented operation timeouts
	cfg.OperationTimeout = 15 * time.Second
	cfg.CreateElevatorTimeout = 5 * time.Second
	cfg.RequestTimeout = 3 * time.Second
	cfg.StatusUpdateTimeout = 1 * time.Second
	cfg.HealthCheckTimeout = 1 * time.Second

	// High-performance WebSocket settings
	cfg.WebSocketConnectionTimeout = 10 * time.Minute
	cfg.WebSocketMaxConnections = 5000
	cfg.WebSocketWriteTimeout = 2 * time.Second
	cfg.WebSocketReadTimeout = 30 * time.Second
	cfg.WebSocketPingInterval = 15 * time.Second

	// Aggressive circuit breaker for fast failure detection
	cfg.CircuitBreakerMaxFailures = 2
	cfg.CircuitBreakerFailureThreshold = 0.3
	cfg.CircuitBreakerResetTimeout = 10 * time.Second

	// Production security settings
	cfg.CORSAllowedOrigins = "https://app.example.com"
	cfg.MaxRequestSize = 512 * 1024 // 512KB for security and performance

	// High-capacity elevator settings for production load
	cfg.MaxElevators = 200
	cfg.SwitchOnChannelBuffer = 50 // Larger buffer for performance
}

// validateConfiguration performs comprehensive configuration validation
func validateConfiguration(cfg *Config) error {
	// Validate floor configuration
	if cfg.MinFloor >= cfg.MaxFloor {
		return domain.NewValidationError("min floor must be less than max floor", nil).
			WithContext("min_floor", cfg.MinFloor).
			WithContext("max_floor", cfg.MaxFloor)
	}

	if cfg.MinFloor < constants.MinAllowedFloor {
		return domain.NewValidationError("min floor is below system minimum", nil).
			WithContext("min_floor", cfg.MinFloor).
			WithContext("system_minimum", constants.MinAllowedFloor)
	}

	if cfg.MaxFloor > constants.MaxAllowedFloor {
		return domain.NewValidationError("max floor exceeds system maximum", nil).
			WithContext("max_floor", cfg.MaxFloor).
			WithContext("system_maximum", constants.MaxAllowedFloor)
	}

	// Validate basic configuration values
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return domain.NewValidationError("port must be between 1 and 65535", nil).
			WithContext("port", cfg.Port)
	}

	if cfg.EachFloorDuration <= 0 {
		return domain.NewValidationError("each floor duration must be positive", nil).
			WithContext("duration", cfg.EachFloorDuration)
	}

	if cfg.MaxElevators <= 0 || cfg.MaxElevators > 1000 {
		return domain.NewValidationError("max elevators must be between 1 and 1000", nil).
			WithContext("max_elevators", cfg.MaxElevators)
	}

	if cfg.DefaultOverloadThreshold <= 0 || cfg.DefaultOverloadThreshold > 100 {
		return domain.NewValidationError("default overload threshold must be between 1 and 100", nil).
			WithContext("default_overload_threshold", cfg.DefaultOverloadThreshold)
	}

	// Environment-specific validations
	if err := validateEnvironmentSpecificConfig(cfg); err != nil {
		return err
	}

	return nil
}

// validateEnvironmentSpecificConfig validates environment-specific configuration issues
func validateEnvironmentSpecificConfig(cfg *Config) error {
	// Production environment validations
	if cfg.IsProduction() {
		if cfg.CORSAllowedOrigins == "*" {
			return domain.NewValidationError("CORS wildcard not allowed in production", nil).
				WithContext("environment", cfg.Environment).
				WithContext("cors_origins", cfg.CORSAllowedOrigins)
		}

		if cfg.LogRequestDetails {
			return domain.NewValidationError("request logging should be disabled in production for performance", nil).
				WithContext("environment", cfg.Environment)
		}

		if cfg.RateLimitRPM > 100 {
			return domain.NewValidationError("rate limit too high for production", nil).
				WithContext("environment", cfg.Environment).
				WithContext("rate_limit", cfg.RateLimitRPM)
		}
	}

	// Development environment warnings (would be logged, not errors)
	if cfg.IsDevelopment() && !cfg.LogRequestDetails {
		// This would be a warning in logs: "Consider enabling LogRequestDetails in development"
		// For now, we just note this configuration without logging
	}

	// Testing environment validations
	if cfg.IsTesting() {
		if cfg.WebSocketEnabled {
			return domain.NewValidationError("WebSocket should be disabled in testing environment", nil).
				WithContext("environment", cfg.Environment)
		}

		if cfg.MetricsEnabled {
			return domain.NewValidationError("Metrics should be disabled in testing environment", nil).
				WithContext("environment", cfg.Environment)
		}
	}

	return nil
}

// validateElevatorConfiguration validates elevator-specific configuration
func validateElevatorConfiguration(config ElevatorConfig) error {
	// Validate floor configuration
	if config.MinFloor >= config.MaxFloor {
		return domain.NewValidationError("min floor must be less than max floor", nil).
			WithContext("min_floor", config.MinFloor).
			WithContext("max_floor", config.MaxFloor)
	}

	if config.MinFloor < constants.MinAllowedFloor {
		return domain.NewValidationError("min floor is below system minimum", nil).
			WithContext("min_floor", config.MinFloor).
			WithContext("system_minimum", constants.MinAllowedFloor)
	}

	if config.MaxFloor > constants.MaxAllowedFloor {
		return domain.NewValidationError("max floor exceeds system maximum", nil).
			WithContext("max_floor", config.MaxFloor).
			WithContext("system_maximum", constants.MaxAllowedFloor)
	}

	// Validate timing configuration
	if config.EachFloorDuration <= 0 {
		return domain.NewValidationError("each floor duration must be positive", nil).
			WithContext("duration", config.EachFloorDuration)
	}

	if config.OpenDoorDuration <= 0 {
		return domain.NewValidationError("open door duration must be positive", nil).
			WithContext("duration", config.OpenDoorDuration)
	}

	// Validate timeout configuration
	if config.OperationTimeout <= 0 {
		return domain.NewValidationError("operation timeout must be positive", nil).
			WithContext("timeout", config.OperationTimeout)
	}

	// Validate limits
	if config.MaxElevators <= 0 || config.MaxElevators > 1000 {
		return domain.NewValidationError("max elevators must be between 1 and 1000", nil).
			WithContext("max_elevators", config.MaxElevators)
	}

	if config.DefaultElevatorCount < 0 || config.DefaultElevatorCount > config.MaxElevators {
		return domain.NewValidationError("default elevator count must be between 0 and max elevators", nil).
			WithContext("default_count", config.DefaultElevatorCount).
			WithContext("max_elevators", config.MaxElevators)
	}

	if config.SwitchOnChannelBuffer <= 0 || config.SwitchOnChannelBuffer > 1000 {
		return domain.NewValidationError("switch on channel buffer must be between 1 and 1000", nil).
			WithContext("buffer_size", config.SwitchOnChannelBuffer)
	}

	if config.DefaultOverloadThreshold <= 0 || config.DefaultOverloadThreshold > 100 {
		return domain.NewValidationError("default overload threshold must be between 1 and 100", nil).
			WithContext("default_overload_threshold", config.DefaultOverloadThreshold)
	}

	return nil
}

// validateServerConfiguration validates server-specific configuration
func validateServerConfiguration(config ServerConfig) error {
	if config.Port <= 0 || config.Port > 65535 {
		return domain.NewValidationError("port must be between 1 and 65535", nil).
			WithContext("port", config.Port)
	}

	if config.ReadTimeout <= 0 {
		return domain.NewValidationError("read timeout must be positive", nil).
			WithContext("timeout", config.ReadTimeout)
	}

	if config.WriteTimeout <= 0 {
		return domain.NewValidationError("write timeout must be positive", nil).
			WithContext("timeout", config.WriteTimeout)
	}

	if config.IdleTimeout <= 0 {
		return domain.NewValidationError("idle timeout must be positive", nil).
			WithContext("timeout", config.IdleTimeout)
	}

	return nil
}

// validateHTTPConfiguration validates HTTP-specific configuration
func validateHTTPConfiguration(config HTTPConfig) error {
	if config.RateLimitRPM <= 0 || config.RateLimitRPM > 100000 {
		return domain.NewValidationError("rate limit RPM must be between 1 and 100000", nil).
			WithContext("rpm", config.RateLimitRPM)
	}

	if config.MaxRequestSize <= 0 || config.MaxRequestSize > 100*1024*1024 { // 100MB max
		return domain.NewValidationError("max request size must be between 1 byte and 100MB", nil).
			WithContext("size", config.MaxRequestSize)
	}

	return nil
}



// validateCircuitBreakerConfiguration validates circuit breaker configuration
func validateCircuitBreakerConfiguration(config CircuitBreakerConfig) error {
	if config.MaxFailures <= 0 || config.MaxFailures > 100 {
		return domain.NewValidationError("max failures must be between 1 and 100", nil).
			WithContext("max_failures", config.MaxFailures)
	}

	if config.ResetTimeout <= 0 {
		return domain.NewValidationError("reset timeout must be positive", nil).
			WithContext("timeout", config.ResetTimeout)
	}

	if config.HalfOpenLimit <= 0 || config.HalfOpenLimit > 50 {
		return domain.NewValidationError("half open limit must be between 1 and 50", nil).
			WithContext("limit", config.HalfOpenLimit)
	}

	if config.FailureThreshold <= 0 || config.FailureThreshold > 1 {
		return domain.NewValidationError("failure threshold must be between 0 and 1", nil).
			WithContext("threshold", config.FailureThreshold)
	}

	return nil
}

// validateWebSocketConfiguration validates WebSocket configuration
func validateWebSocketConfiguration(config WebSocketConfig) error {
	if config.ConnectionTimeout <= 0 {
		return domain.NewValidationError("connection timeout must be positive", nil).
			WithContext("timeout", config.ConnectionTimeout)
	}

	if config.MaxConnections <= 0 || config.MaxConnections > 10000 {
		return domain.NewValidationError("max connections must be between 1 and 10000", nil).
			WithContext("max_connections", config.MaxConnections)
	}

	if config.BufferSize <= 0 || config.BufferSize > 65536 {
		return domain.NewValidationError("buffer size must be between 1 and 65536", nil).
			WithContext("buffer_size", config.BufferSize)
	}

	return nil
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production" || c.Environment == "prod"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development" || c.Environment == "dev"
}

// IsTesting returns true if running in testing environment
func (c *Config) IsTesting() bool {
	return c.Environment == "testing" || c.Environment == "test"
}

// GetEnvironmentInfo returns environment information for logging/debugging
func (c *Config) GetEnvironmentInfo() map[string]interface{} {
	return map[string]interface{}{
		"environment":             c.Environment,
		"log_level":               c.LogLevel,
		"port":                    c.Port,
		"metrics_enabled":         c.MetricsEnabled,
		"websocket_enabled":       c.WebSocketEnabled,
		"circuit_breaker_enabled": c.CircuitBreakerEnabled,
	}
}
