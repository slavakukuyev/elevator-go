package config

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/slavakukuyev/elevator-go/internal/constants"
	"github.com/slavakukuyev/elevator-go/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitConfig_DefaultValues(t *testing.T) {
	// Clear environment variables to test defaults
	cleanupEnv := clearEnvVars()
	defer cleanupEnv()

	cfg, err := InitConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Test default values with development environment defaults applied
	assert.Equal(t, "development", cfg.Environment)
	assert.Equal(t, "DEBUG", cfg.LogLevel)            // Only change: DEBUG logging
	assert.Equal(t, 6660, cfg.Port)                   // Default value
	assert.Equal(t, 30*time.Second, cfg.ReadTimeout)  // Default value
	assert.Equal(t, 30*time.Second, cfg.WriteTimeout) // Default value
	assert.Equal(t, 120*time.Second, cfg.IdleTimeout) // Default value
	assert.Equal(t, 9, cfg.MaxFloor)
	assert.Equal(t, 0, cfg.MinFloor)
	assert.Equal(t, 500*time.Millisecond, cfg.EachFloorDuration) // Default value
	assert.Equal(t, 2*time.Second, cfg.OpenDoorDuration)         // Default value
	assert.Equal(t, 100, cfg.MaxElevators)
	assert.Equal(t, "Elevator", cfg.NamePrefix)
	assert.Equal(t, 100, cfg.RateLimitRPM) // Default value
	assert.True(t, cfg.LogRequestDetails)  // Enabled in development
}

func TestInitConfig_EnvironmentVariables(t *testing.T) {
	cleanupEnv := clearEnvVars()
	defer cleanupEnv()

	// Set specific environment variables
	envVars := map[string]string{
		"ENV":                     "production",
		"LOG_LEVEL":               "ERROR",
		"PORT":                    "8080",
		"DEFAULT_MAX_FLOOR":       "20",
		"DEFAULT_MIN_FLOOR":       "-5",
		"EACH_FLOOR_DURATION":     "1s",
		"OPEN_DOOR_DURATION":      "3s",
		"MAX_ELEVATORS":           "50",
		"ELEVATOR_NAME_PREFIX":    "Lift",
		"RATE_LIMIT_RPM":          "200",
		"WEBSOCKET_ENABLED":       "false",
		"CIRCUIT_BREAKER_ENABLED": "false",
	}

	for key, value := range envVars {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("Failed to set environment variable %s: %v", key, err)
		}
	}

	cfg, err := InitConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify environment variables are parsed correctly
	assert.Equal(t, "production", cfg.Environment)
	assert.Equal(t, "WARN", cfg.LogLevel) // Should be overridden to WARN in production defaults
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, 20, cfg.MaxFloor)
	assert.Equal(t, -5, cfg.MinFloor)
	assert.Equal(t, 200*time.Millisecond, cfg.EachFloorDuration) // Overridden by production defaults
	assert.Equal(t, 1*time.Second, cfg.OpenDoorDuration)         // Overridden by production defaults
	assert.Equal(t, 200, cfg.MaxElevators)                       // Overridden by production defaults
	assert.Equal(t, "Lift", cfg.NamePrefix)
	assert.Equal(t, 30, cfg.RateLimitRPM) // Should be overridden to 30 in production defaults
	assert.False(t, cfg.WebSocketEnabled)
	assert.False(t, cfg.CircuitBreakerEnabled)
}

func TestEnvironmentDefaults_Development(t *testing.T) {
	cleanupEnv := clearEnvVars()
	defer cleanupEnv()

	if err := os.Setenv("ENV", "development"); err != nil {
		t.Fatalf("Failed to set ENV variable: %v", err)
	}

	cfg, err := InitConfig()
	require.NoError(t, err)

	assert.Equal(t, "development", cfg.Environment)
	assert.Equal(t, "DEBUG", cfg.LogLevel)
	// All other values should be defaults
	assert.Equal(t, 500*time.Millisecond, cfg.EachFloorDuration) // Default value
	assert.Equal(t, 2*time.Second, cfg.OpenDoorDuration)         // Default value
	assert.Equal(t, 100, cfg.RateLimitRPM)                       // Default value
	assert.True(t, cfg.LogRequestDetails)                        // Only other change
}

func TestEnvironmentDefaults_Testing(t *testing.T) {
	cleanupEnv := clearEnvVars()
	defer cleanupEnv()

	if err := os.Setenv("ENV", "testing"); err != nil {
		t.Fatalf("Failed to set ENV variable: %v", err)
	}

	cfg, err := InitConfig()
	require.NoError(t, err)

	assert.Equal(t, "testing", cfg.Environment)
	assert.Equal(t, "WARN", cfg.LogLevel)
	// Very strict settings for testing
	assert.Equal(t, 10*time.Millisecond, cfg.EachFloorDuration)
	assert.Equal(t, 10*time.Millisecond, cfg.OpenDoorDuration)
	assert.Equal(t, 500*time.Millisecond, cfg.OperationTimeout)      // Stricter
	assert.Equal(t, 500*time.Millisecond, cfg.CreateElevatorTimeout) // Stricter
	assert.Equal(t, 200*time.Millisecond, cfg.RequestTimeout)        // Stricter
	assert.Equal(t, 2*time.Second, cfg.ReadTimeout)                  // Stricter
	assert.Equal(t, 2*time.Second, cfg.WriteTimeout)                 // Stricter
	assert.Equal(t, 10*time.Second, cfg.IdleTimeout)                 // Stricter
	assert.False(t, cfg.MetricsEnabled)
	assert.False(t, cfg.WebSocketEnabled)
	assert.False(t, cfg.LogRequestDetails)            // Disabled
	assert.Equal(t, 1000, cfg.RateLimitRPM)           // Controlled but higher for acceptance tests
	assert.Equal(t, 5, cfg.MaxElevators)              // Minimal
	assert.Equal(t, 1, cfg.CircuitBreakerMaxFailures) // Aggressive
}

func TestEnvironmentDefaults_Production(t *testing.T) {
	cleanupEnv := clearEnvVars()
	defer cleanupEnv()

	if err := os.Setenv("ENV", "production"); err != nil {
		t.Fatalf("Failed to set ENV variable: %v", err)
	}

	cfg, err := InitConfig()
	require.NoError(t, err)

	assert.Equal(t, "production", cfg.Environment)
	// High-performance and strict production settings
	assert.Equal(t, "WARN", cfg.LogLevel)                              // Minimal logging
	assert.Equal(t, 30, cfg.RateLimitRPM)                              // Strict rate limiting
	assert.False(t, cfg.LogRequestDetails)                             // No detailed logging
	assert.Equal(t, 15*time.Second, cfg.ReadTimeout)                   // Optimized timeouts
	assert.Equal(t, 15*time.Second, cfg.WriteTimeout)                  // Optimized timeouts
	assert.Equal(t, 60*time.Second, cfg.IdleTimeout)                   // Optimized timeouts
	assert.Equal(t, 200*time.Millisecond, cfg.EachFloorDuration)       // Faster operations
	assert.Equal(t, 1*time.Second, cfg.OpenDoorDuration)               // Faster operations
	assert.Equal(t, 200, cfg.MaxElevators)                             // High capacity
	assert.Equal(t, 5000, cfg.WebSocketMaxConnections)                 // High capacity
	assert.Equal(t, 2, cfg.CircuitBreakerMaxFailures)                  // Aggressive
	assert.Equal(t, "https://app.example.com", cfg.CORSAllowedOrigins) // Secure
}

func TestConfigValidation_ValidConfiguration(t *testing.T) {
	cleanupEnv := clearEnvVars()
	defer cleanupEnv()

	envVars := map[string]string{
		"ENV":                               "development",
		"PORT":                              "8080",
		"DEFAULT_MAX_FLOOR":                 "10",
		"DEFAULT_MIN_FLOOR":                 "0",
		"EACH_FLOOR_DURATION":               "500ms",
		"MAX_ELEVATORS":                     "50",
		"RATE_LIMIT_RPM":                    "100",
		"MAX_REQUEST_SIZE":                  "2097152", // 2MB
		"CIRCUIT_BREAKER_MAX_FAILURES":      "3",
		"CIRCUIT_BREAKER_FAILURE_THRESHOLD": "0.5",
		"WEBSOCKET_MAX_CONNECTIONS":         "500",
		"WEBSOCKET_BUFFER_SIZE":             "2048",
	}

	for key, value := range envVars {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("Failed to set environment variable %s: %v", key, err)
		}
	}

	cfg, err := InitConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

func TestConfigValidation_InvalidFloorConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		minFloor string
		maxFloor string
		wantErr  string
	}{
		{
			name:     "min floor equals max floor",
			minFloor: "5",
			maxFloor: "5",
			wantErr:  "min floor must be less than max floor",
		},
		{
			name:     "min floor greater than max floor",
			minFloor: "10",
			maxFloor: "5",
			wantErr:  "min floor must be less than max floor",
		},
		{
			name:     "min floor below system minimum",
			minFloor: "-150",
			maxFloor: "10",
			wantErr:  "min floor is below system minimum",
		},
		{
			name:     "max floor exceeds system maximum",
			minFloor: "0",
			maxFloor: "250",
			wantErr:  "max floor exceeds system maximum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupEnv := clearEnvVars()
			defer cleanupEnv()

			if err := os.Setenv("DEFAULT_MIN_FLOOR", tt.minFloor); err != nil {
				t.Fatalf("Failed to set DEFAULT_MIN_FLOOR variable: %v", err)
			}
			if err := os.Setenv("DEFAULT_MAX_FLOOR", tt.maxFloor); err != nil {
				t.Fatalf("Failed to set DEFAULT_MAX_FLOOR variable: %v", err)
			}

			cfg, err := InitConfig()
			require.Error(t, err)
			assert.Nil(t, cfg)
			assert.Contains(t, err.Error(), tt.wantErr)

			// Verify it's a validation error
			var domainErr *domain.DomainError
			require.ErrorAs(t, err, &domainErr)
			assert.Equal(t, domain.ErrTypeValidation, domainErr.Type)
		})
	}
}

func TestConfigValidation_InvalidPortConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		wantErr string
	}{
		{
			name:    "port zero",
			port:    "0",
			wantErr: "port must be between 1 and 65535",
		},
		{
			name:    "negative port",
			port:    "-1",
			wantErr: "port must be between 1 and 65535",
		},
		{
			name:    "port too high",
			port:    "70000",
			wantErr: "port must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupEnv := clearEnvVars()
			defer cleanupEnv()

			if err := os.Setenv("PORT", tt.port); err != nil {
				t.Fatalf("Failed to set PORT variable: %v", err)
			}

			cfg, err := InitConfig()
			require.Error(t, err)
			assert.Nil(t, cfg)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestConfigValidation_InvalidDurationConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		envVar  string
		value   string
		wantErr string
	}{
		{
			name:    "negative each floor duration",
			envVar:  "EACH_FLOOR_DURATION",
			value:   "-1s",
			wantErr: "each floor duration must be positive",
		},
		{
			name:    "zero each floor duration",
			envVar:  "EACH_FLOOR_DURATION",
			value:   "0s",
			wantErr: "each floor duration must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupEnv := clearEnvVars()
			defer cleanupEnv()

			if err := os.Setenv(tt.envVar, tt.value); err != nil {
				t.Fatalf("Failed to set environment variable %s: %v", tt.envVar, err)
			}

			cfg, err := InitConfig()
			require.Error(t, err)
			assert.Nil(t, cfg)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestConfigValidation_InvalidMaxElevators(t *testing.T) {
	tests := []struct {
		name         string
		maxElevators string
		wantErr      string
	}{
		{
			name:         "zero max elevators",
			maxElevators: "0",
			wantErr:      "max elevators must be between 1 and 1000",
		},
		{
			name:         "negative max elevators",
			maxElevators: "-1",
			wantErr:      "max elevators must be between 1 and 1000",
		},
		{
			name:         "too many max elevators",
			maxElevators: "1001",
			wantErr:      "max elevators must be between 1 and 1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupEnv := clearEnvVars()
			defer cleanupEnv()

			if err := os.Setenv("MAX_ELEVATORS", tt.maxElevators); err != nil {
				t.Fatalf("Failed to set MAX_ELEVATORS variable: %v", err)
			}

			cfg, err := InitConfig()
			require.Error(t, err)
			assert.Nil(t, cfg)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestValidateElevatorConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		config  ElevatorConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid configuration",
			config: ElevatorConfig{
				MinFloor:                 0,
				MaxFloor:                 10,
				DefaultOverloadThreshold: 12,
				EachFloorDuration:        500 * time.Millisecond,
				OpenDoorDuration:         2 * time.Second,
				OperationTimeout:         30 * time.Second,
				MaxElevators:             100,
				DefaultElevatorCount:     5,
				SwitchOnChannelBuffer:    10,
			},
			wantErr: false,
		},
		{
			name: "invalid floor range",
			config: ElevatorConfig{
				MinFloor: 10,
				MaxFloor: 5,
			},
			wantErr: true,
			errMsg:  "min floor must be less than max floor",
		},
		{
			name: "default elevator count exceeds max",
			config: ElevatorConfig{
				MinFloor:                 0,
				MaxFloor:                 10,
				DefaultOverloadThreshold: 12,
				EachFloorDuration:        500 * time.Millisecond,
				OpenDoorDuration:         2 * time.Second,
				OperationTimeout:         30 * time.Second,
				MaxElevators:             10,
				DefaultElevatorCount:     15,
			},
			wantErr: true,
			errMsg:  "default elevator count must be between 0 and max elevators",
		},
		{
			name: "invalid overload threshold - too low",
			config: ElevatorConfig{
				MinFloor:                 0,
				MaxFloor:                 10,
				DefaultOverloadThreshold: 0,
				EachFloorDuration:        500 * time.Millisecond,
				OpenDoorDuration:         2 * time.Second,
				OperationTimeout:         30 * time.Second,
				MaxElevators:             10,
				DefaultElevatorCount:     5,
				SwitchOnChannelBuffer:    10,
			},
			wantErr: true,
			errMsg:  "default overload threshold must be between 1 and 100",
		},
		{
			name: "invalid overload threshold - too high",
			config: ElevatorConfig{
				MinFloor:                 0,
				MaxFloor:                 10,
				DefaultOverloadThreshold: 150,
				EachFloorDuration:        500 * time.Millisecond,
				OpenDoorDuration:         2 * time.Second,
				OperationTimeout:         30 * time.Second,
				MaxElevators:             10,
				DefaultElevatorCount:     5,
				SwitchOnChannelBuffer:    10,
			},
			wantErr: true,
			errMsg:  "default overload threshold must be between 1 and 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateElevatorConfiguration(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateServerConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		config  ServerConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid configuration",
			config: ServerConfig{
				Port:         8080,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  120 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "invalid port",
			config: ServerConfig{
				Port: 0,
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name: "negative timeout",
			config: ServerConfig{
				Port:        8080,
				ReadTimeout: -1 * time.Second,
			},
			wantErr: true,
			errMsg:  "read timeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateServerConfiguration(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateHTTPConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		config  HTTPConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid configuration",
			config: HTTPConfig{
				RateLimitRPM:   100,
				MaxRequestSize: 1024 * 1024, // 1MB
			},
			wantErr: false,
		},
		{
			name: "rate limit too high",
			config: HTTPConfig{
				RateLimitRPM: 200000,
			},
			wantErr: true,
			errMsg:  "rate limit RPM must be between 1 and 100000",
		},
		{
			name: "request size too large",
			config: HTTPConfig{
				RateLimitRPM:   100,
				MaxRequestSize: 200 * 1024 * 1024, // 200MB
			},
			wantErr: true,
			errMsg:  "max request size must be between 1 byte and 100MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHTTPConfiguration(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateCircuitBreakerConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		config  CircuitBreakerConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid configuration",
			config: CircuitBreakerConfig{
				MaxFailures:      5,
				ResetTimeout:     30 * time.Second,
				HalfOpenLimit:    3,
				FailureThreshold: 0.6,
			},
			wantErr: false,
		},
		{
			name: "invalid failure threshold",
			config: CircuitBreakerConfig{
				MaxFailures:      5,
				ResetTimeout:     30 * time.Second,
				HalfOpenLimit:    3,
				FailureThreshold: 1.5,
			},
			wantErr: true,
			errMsg:  "failure threshold must be between 0 and 1",
		},
		{
			name: "too many max failures",
			config: CircuitBreakerConfig{
				MaxFailures: 150,
			},
			wantErr: true,
			errMsg:  "max failures must be between 1 and 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCircuitBreakerConfiguration(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateWebSocketConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		config  WebSocketConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid configuration",
			config: WebSocketConfig{
				ConnectionTimeout: 10 * time.Minute,
				MaxConnections:    1000,
				BufferSize:        1024,
			},
			wantErr: false,
		},
		{
			name: "too many connections",
			config: WebSocketConfig{
				ConnectionTimeout: 10 * time.Minute,
				MaxConnections:    15000,
				BufferSize:        1024,
			},
			wantErr: true,
			errMsg:  "max connections must be between 1 and 10000",
		},
		{
			name: "buffer size too large",
			config: WebSocketConfig{
				ConnectionTimeout: 10 * time.Minute,
				MaxConnections:    1000,
				BufferSize:        100000,
			},
			wantErr: true,
			errMsg:  "buffer size must be between 1 and 65536",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWebSocketConfiguration(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfig_EnvironmentMethods(t *testing.T) {
	tests := []struct {
		name          string
		environment   string
		isProduction  bool
		isDevelopment bool
		isTesting     bool
	}{
		{
			name:          "production environment",
			environment:   "production",
			isProduction:  true,
			isDevelopment: false,
			isTesting:     false,
		},
		{
			name:          "prod environment",
			environment:   "prod",
			isProduction:  true,
			isDevelopment: false,
			isTesting:     false,
		},
		{
			name:          "development environment",
			environment:   "development",
			isProduction:  false,
			isDevelopment: true,
			isTesting:     false,
		},
		{
			name:          "dev environment",
			environment:   "dev",
			isProduction:  false,
			isDevelopment: true,
			isTesting:     false,
		},
		{
			name:          "testing environment",
			environment:   "testing",
			isProduction:  false,
			isDevelopment: false,
			isTesting:     true,
		},
		{
			name:          "test environment",
			environment:   "test",
			isProduction:  false,
			isDevelopment: false,
			isTesting:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Environment: tt.environment}

			assert.Equal(t, tt.isProduction, cfg.IsProduction())
			assert.Equal(t, tt.isDevelopment, cfg.IsDevelopment())
			assert.Equal(t, tt.isTesting, cfg.IsTesting())
		})
	}
}

func TestConfig_GetEnvironmentInfo(t *testing.T) {
	cfg := &Config{
		Environment:           "development",
		LogLevel:              "DEBUG",
		Port:                  8080,
		MetricsEnabled:        true,
		WebSocketEnabled:      true,
		CircuitBreakerEnabled: false,
	}

	info := cfg.GetEnvironmentInfo()

	expected := map[string]interface{}{
		"environment":             "development",
		"log_level":               "DEBUG",
		"port":                    8080,
		"metrics_enabled":         true,
		"websocket_enabled":       true,
		"circuit_breaker_enabled": false,
	}

	assert.Equal(t, expected, info)
}

func TestConfigBoundaryValues(t *testing.T) {
	cleanupEnv := clearEnvVars()
	defer cleanupEnv()

	// Test boundary values that should be valid
	envVars := map[string]string{
		"DEFAULT_MIN_FLOOR":                 "-100", // Minimum allowed
		"DEFAULT_MAX_FLOOR":                 "200",  // Maximum allowed
		"PORT":                              "1",    // Minimum port
		"MAX_ELEVATORS":                     "1000", // Maximum elevators
		"RATE_LIMIT_RPM":                    "1",    // Minimum rate limit
		"MAX_REQUEST_SIZE":                  "1",    // Minimum request size
		"CIRCUIT_BREAKER_MAX_FAILURES":      "1",    // Minimum failures
		"CIRCUIT_BREAKER_FAILURE_THRESHOLD": "0.01", // Very low threshold
		"WEBSOCKET_MAX_CONNECTIONS":         "1",    // Minimum connections
		"WEBSOCKET_BUFFER_SIZE":             "1",    // Minimum buffer size
	}

	for key, value := range envVars {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("Failed to set environment variable %s: %v", key, err)
		}
	}

	cfg, err := InitConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, constants.MinAllowedFloor, cfg.MinFloor)
	assert.Equal(t, constants.MaxAllowedFloor, cfg.MaxFloor)
	assert.Equal(t, 1, cfg.Port)
	assert.Equal(t, 1000, cfg.MaxElevators)
}

func TestConfigWithAlternativeEnvironmentNames(t *testing.T) {
	environments := []struct {
		envName      string
		expectedType string
	}{
		{"dev", "development"},
		{"development", "development"},
		{"test", "testing"},
		{"testing", "testing"},

		{"prod", "production"},
		{"production", "production"},
	}

	for _, env := range environments {
		t.Run(env.envName, func(t *testing.T) {
			cleanupEnv := clearEnvVars()
			defer cleanupEnv()

			if err := os.Setenv("ENV", env.envName); err != nil {
				t.Fatalf("Failed to set ENV variable: %v", err)
			}

			cfg, err := InitConfig()
			require.NoError(t, err)

			switch env.expectedType {
			case "development":
				assert.True(t, cfg.IsDevelopment())
				assert.False(t, cfg.IsProduction())
				assert.False(t, cfg.IsTesting())
			case "testing":
				assert.False(t, cfg.IsDevelopment())
				assert.False(t, cfg.IsProduction())
				assert.True(t, cfg.IsTesting())
			case "production":
				assert.False(t, cfg.IsDevelopment())
				assert.True(t, cfg.IsProduction())
				assert.False(t, cfg.IsTesting())
			}
		})
	}
}

// Helper function to clear environment variables used by config
func clearEnvVars() func() {
	envVars := []string{
		"ENV", "LOG_LEVEL", "PORT", "SERVER_READ_TIMEOUT", "SERVER_WRITE_TIMEOUT",
		"SERVER_IDLE_TIMEOUT", "SERVER_SHUTDOWN_TIMEOUT", "SERVER_SHUTDOWN_GRACE",
		"DEFAULT_MAX_FLOOR", "DEFAULT_MIN_FLOOR", "EACH_FLOOR_DURATION",
		"OPEN_DOOR_DURATION", "ELEVATOR_OPERATION_TIMEOUT", "CREATE_ELEVATOR_TIMEOUT",
		"ELEVATOR_REQUEST_TIMEOUT", "STATUS_UPDATE_TIMEOUT", "HEALTH_CHECK_TIMEOUT",
		"MAX_ELEVATORS", "DEFAULT_ELEVATOR_COUNT", "ELEVATOR_NAME_PREFIX",
		"SWITCH_ON_CHANNEL_BUFFER", "RATE_LIMIT_RPM", "RATE_LIMIT_WINDOW",
		"RATE_LIMIT_CLEANUP", "MAX_REQUEST_SIZE", "HTTP_REQUEST_TIMEOUT",
		"CORS_ENABLED", "CORS_MAX_AGE", "CORS_ALLOWED_ORIGINS", "METRICS_ENABLED",
		"METRICS_PATH", "STATUS_UPDATE_INTERVAL", "HEALTH_ENABLED", "HEALTH_PATH",
		"STRUCTURED_LOGGING", "LOG_REQUEST_DETAILS", "CORRELATION_ID_HEADER",
		"CIRCUIT_BREAKER_ENABLED", "CIRCUIT_BREAKER_MAX_FAILURES",
		"CIRCUIT_BREAKER_RESET_TIMEOUT", "CIRCUIT_BREAKER_HALF_OPEN_LIMIT",
		"CIRCUIT_BREAKER_FAILURE_THRESHOLD", "WEBSOCKET_ENABLED", "WEBSOCKET_PATH",
		"WEBSOCKET_CONNECTION_TIMEOUT", "WEBSOCKET_WRITE_TIMEOUT",
		"WEBSOCKET_READ_TIMEOUT", "WEBSOCKET_PING_INTERVAL",
		"WEBSOCKET_MAX_CONNECTIONS", "WEBSOCKET_BUFFER_SIZE",
	}

	// Store original values
	originalValues := make(map[string]string)
	for _, envVar := range envVars {
		originalValues[envVar] = os.Getenv(envVar)
		if err := os.Unsetenv(envVar); err != nil {
			// Log but don't fail, as this is cleanup code
			fmt.Printf("Failed to unset environment variable %s: %v\n", envVar, err)
		}
	}

	// Return cleanup function
	return func() {
		for _, envVar := range envVars {
			if originalValue, exists := originalValues[envVar]; exists && originalValue != "" {
				os.Setenv(envVar, originalValue)
			} else {
				if err := os.Unsetenv(envVar); err != nil {
					// Log but don't fail, as this is cleanup code
					fmt.Printf("Failed to unset environment variable %s: %v\n", envVar, err)
				}
			}
		}
	}
}
