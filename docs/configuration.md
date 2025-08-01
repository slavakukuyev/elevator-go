# Configuration & Environment Management

## Overview

The Elevator Control System uses a comprehensive configuration management system that supports environment-specific settings, extensive validation, and flexible deployment configurations. The system is built on environment variables with strong typing and comprehensive validation.

## Architecture

The configuration system is organized into logical groups:
- **Environment & Logging**: Basic environment settings
- **Server Configuration**: HTTP server and networking settings  
- **Elevator System**: Core elevator operation parameters
- **HTTP & Middleware**: Request handling and rate limiting
- **Monitoring & Observability**: Metrics, health checks, logging
- **Circuit Breaker**: Fault tolerance configuration
- **WebSocket**: Real-time communication settings

## Environment Variables

### Environment & Basic Settings
| Variable | Default | Description |
|----------|---------|-------------|
| `ENV` | `development` | Environment type: development, testing, production |
| `LOG_LEVEL` | `INFO` | Logging level: DEBUG, INFO, WARN, ERROR |

### Server Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `6660` | HTTP server port |
| `SERVER_READ_TIMEOUT` | `30s` | Maximum duration for reading the entire request |
| `SERVER_WRITE_TIMEOUT` | `30s` | Maximum duration before timing out writes |
| `SERVER_IDLE_TIMEOUT` | `120s` | Maximum idle time before closing connections |
| `SERVER_SHUTDOWN_TIMEOUT` | `30s` | Maximum time to wait for server shutdown |
| `SERVER_SHUTDOWN_GRACE` | `2s` | Grace period after shutdown |

### Elevator System Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `DEFAULT_MAX_FLOOR` | `9` | Default maximum floor for new elevators |
| `DEFAULT_MIN_FLOOR` | `0` | Default minimum floor for new elevators |
| `DEFAULT_OVERLOAD_THRESHOLD` | `12` | Default overload threshold for new elevators (1-100) |
| `EACH_FLOOR_DURATION` | `500ms` | Time to travel between floors |
| `OPEN_DOOR_DURATION` | `2s` | Duration to keep doors open |
| `ELEVATOR_OPERATION_TIMEOUT` | `30s` | Timeout for elevator operations |
| `CREATE_ELEVATOR_TIMEOUT` | `10s` | Timeout for elevator creation |
| `ELEVATOR_REQUEST_TIMEOUT` | `5s` | Timeout for processing requests |
| `STATUS_UPDATE_TIMEOUT` | `3s` | Timeout for status updates |
| `HEALTH_CHECK_TIMEOUT` | `2s` | Timeout for health checks |
| `MAX_ELEVATORS` | `100` | Maximum number of elevators in system |
| `DEFAULT_ELEVATOR_COUNT` | `0` | Number of elevators to create at startup |
| `ELEVATOR_NAME_PREFIX` | `Elevator` | Prefix for auto-generated elevator names |
| `SWITCH_ON_CHANNEL_BUFFER` | `10` | Buffer size for elevator event channels |

### HTTP & Middleware Configuration  
| Variable | Default | Description |
|----------|---------|-------------|
| `RATE_LIMIT_RPM` | `100` | Rate limit in requests per minute per IP |
| `RATE_LIMIT_WINDOW` | `1m` | Time window for rate limiting |
| `RATE_LIMIT_CLEANUP` | `5m` | Cleanup interval for rate limiter |
| `MAX_REQUEST_SIZE` | `1048576` | Maximum request body size (1MB) |
| `HTTP_REQUEST_TIMEOUT` | `30s` | Timeout for HTTP requests |
| `CORS_ENABLED` | `true` | Enable CORS middleware |
| `CORS_MAX_AGE` | `12h` | CORS preflight cache duration |
| `CORS_ALLOWED_ORIGINS` | `*` | Allowed CORS origins (comma-separated) |

### Monitoring & Observability
| Variable | Default | Description |
|----------|---------|-------------|
| `METRICS_ENABLED` | `true` | Enable Prometheus metrics |
| `METRICS_PATH` | `/metrics` | Metrics endpoint path |
| `STATUS_UPDATE_INTERVAL` | `1s` | WebSocket status update frequency |
| `HEALTH_ENABLED` | `true` | Enable health check endpoint |
| `HEALTH_PATH` | `/health` | Health check endpoint path |
| `STRUCTURED_LOGGING` | `true` | Use structured JSON logging |
| `LOG_REQUEST_DETAILS` | `false` | Log detailed request information |
| `CORRELATION_ID_HEADER` | `X-Request-ID` | Header name for request correlation |

### Circuit Breaker Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `CIRCUIT_BREAKER_ENABLED` | `true` | Enable circuit breaker protection |
| `CIRCUIT_BREAKER_MAX_FAILURES` | `5` | Failures before opening circuit |
| `CIRCUIT_BREAKER_RESET_TIMEOUT` | `30s` | Time before trying to close circuit |
| `CIRCUIT_BREAKER_HALF_OPEN_LIMIT` | `3` | Requests allowed in half-open state |
| `CIRCUIT_BREAKER_FAILURE_THRESHOLD` | `0.6` | Failure rate threshold (0.0-1.0) |

### WebSocket Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `WEBSOCKET_ENABLED` | `true` | Enable WebSocket status updates |
| `WEBSOCKET_PATH` | `/ws/status` | WebSocket endpoint path |
| `WEBSOCKET_CONNECTION_TIMEOUT` | `10m` | Maximum connection duration |
| `WEBSOCKET_WRITE_TIMEOUT` | `5s` | Write operation timeout |
| `WEBSOCKET_READ_TIMEOUT` | `60s` | Read operation timeout |
| `WEBSOCKET_PING_INTERVAL` | `30s` | Ping interval for keep-alive |
| `WEBSOCKET_MAX_CONNECTIONS` | `1000` | Maximum concurrent connections |
| `WEBSOCKET_BUFFER_SIZE` | `1024` | WebSocket buffer size |

## Environment Configuration Matrix

The following table shows how key settings differ across environments:

| **Setting** | **Development** | **Testing** | **Production** | **Base Default** |
|-------------|-----------------|-------------|----------------|------------------|
| **LogLevel** | `DEBUG` | `WARN` | `WARN` | `INFO` |
| **LogRequestDetails** | `true` | `false` | `false` | `false` |
| **EachFloorDuration** | `500ms` | `10ms` | `200ms` | `500ms` |
| **OpenDoorDuration** | `2s` | `10ms` | `2s` | `2s` |
| **DefaultOverloadThreshold** | `12` | `5` | `15` | `12` |
| **ElevatorOperationTimeout** | `30s` | `500ms` | `15s` | `30s` |
| **RateLimitRPM** | `100` | `1000` | `30` | `100` |
| **ReadTimeout** | `30s` | `5s` | `30s` | `30s` |
| **WriteTimeout** | `30s` | `5s` | `30s` | `30s` |
| **IdleTimeout** | `120s` | `30s` | `120s` | `120s` |
| **WebSocketEnabled** | `true` | `false` | `true` | `true` |
| **MetricsEnabled** | `true` | `false` | `true` | `true` |
| **WebSocketMaxConnections** | `1000` | `0` | `5000` | `1000` |
| **CircuitBreakerMaxFailures** | `5` | `1` | `2` | `5` |
| **MaxElevators** | `100` | `5` | `200` | `100` |
| **CORSAllowedOrigins** | `*` | `*` | `specific-domain` | `*` |
| **MaxRequestSize** | `1MB` | `1MB` | `1MB` | `1MB` |

## Environment-Specific Configurations

### Development Environment
- **Default configuration** with minimal changes for consistency
- **Enhanced debugging** with `DEBUG` log level for detailed troubleshooting  
- **Detailed request logging** enabled for API debugging
- **Unrestricted CORS** (`*`) for local development convenience
- **Standard settings** for most parameters to match production behavior
- **Focus on debugging** rather than performance optimization

### Testing Environment (Strictest - Catches Issues Early)
- **Aggressive timeouts** (500ms operation timeout vs 30s default) to catch slow operations
- **Very fast operations** (10ms floor duration) for rapid test execution
- **Minimal resources** (5 elevators max) for controlled testing environment
- **Lower overload threshold** (5 vs 12 default) to test capacity limits early
- **Aggressive circuit breaker** (1 failure triggers) to test fault handling
- **Disabled non-essential features** (metrics, WebSocket) to avoid test interference
- **High rate limiting** (1000 RPM) for acceptance tests while still being restricted
- **Focused logging** (`WARN` level) to surface issues without noise

### Production Environment (High-Performance & Secure)
- **Optimized performance** with faster operations (200ms floor duration vs 500ms default)
- **Minimal logging** (`WARN` level) for performance and security
- **Optimized timeouts** (15s operation timeout vs 30s default) for better responsiveness
- **Higher overload threshold** (15 vs 12 default) for increased throughput capacity
- **Strict rate limiting** (30 RPM) to prevent abuse while allowing legitimate use
- **High capacity** (200 elevators, 5000 WebSocket connections) for scalability
- **Aggressive circuit breaker** (2 failures) for fast fault detection
- **Security-focused** with specific CORS origins (no wildcards)

## Recommended Environment Variables

### Development Environment
```bash
ENV=development
LOG_LEVEL=DEBUG
LOG_REQUEST_DETAILS=true
# All other settings use defaults for consistency
```

### Testing Environment
```bash
ENV=testing
LOG_LEVEL=WARN
EACH_FLOOR_DURATION=10ms
OPEN_DOOR_DURATION=10ms
DEFAULT_OVERLOAD_THRESHOLD=5
ELEVATOR_OPERATION_TIMEOUT=500ms
SERVER_READ_TIMEOUT=5s
SERVER_WRITE_TIMEOUT=5s
SERVER_IDLE_TIMEOUT=30s
RATE_LIMIT_RPM=1000
CIRCUIT_BREAKER_MAX_FAILURES=1
MAX_ELEVATORS=5
WEBSOCKET_ENABLED=false
METRICS_ENABLED=false
WEBSOCKET_MAX_CONNECTIONS=0
```

### Production Environment
```bash
ENV=production
LOG_LEVEL=WARN
LOG_REQUEST_DETAILS=false
CORS_ALLOWED_ORIGINS=https://app.example.com,https://admin.example.com
RATE_LIMIT_RPM=30
EACH_FLOOR_DURATION=200ms
DEFAULT_OVERLOAD_THRESHOLD=15
ELEVATOR_OPERATION_TIMEOUT=15s
WEBSOCKET_MAX_CONNECTIONS=5000
CIRCUIT_BREAKER_MAX_FAILURES=2
MAX_ELEVATORS=200
```

## Configuration Files

Environment-specific configuration files can be created in the `configs/` directory:

- `configs/development.env` - Development settings
- `configs/testing.env` - Testing/CI settings  
- `configs/production.env` - Production settings

To use a specific environment configuration:

```bash
# Load development configuration
export $(cat configs/development.env | xargs)

# Or set the environment directly
export ENV=production
```

## Configuration Validation

The configuration system includes comprehensive validation at startup with environment-specific checks:

### Basic Configuration Validation

#### Floor Configuration
- `min_floor` must be less than `max_floor`
- Floor ranges must be within system limits (-100 to 200)
- Floor values must be integers

#### Elevator Capacity Configuration
- `default_overload_threshold` must be between 1 and 100
- Value must be a positive integer
- Used for individual elevator capacity management

#### Timeout Configuration
- All timeout values must be positive durations
- Timeouts are validated at startup
- Invalid timeouts cause startup failure

#### Server Configuration
- Port must be between 1 and 65535
- Timeout durations must be positive
- Request size limits are enforced

#### Rate Limiting
- RPM must be between 1 and 100,000
- Window duration must be positive
- Cleanup intervals are validated

#### Circuit Breaker
- Max failures must be between 1 and 100
- Reset timeout must be positive
- Half-open limit must be between 1 and 50
- Failure threshold must be between 0.0 and 1.0

#### WebSocket
- Connection limits must be reasonable (1-10,000)
- Buffer sizes must be positive
- Timeout values are validated

### Environment-Specific Validation

#### Production Environment Security Checks
- **CORS Origins**: Wildcard (`*`) is not allowed in production
- **Request Logging**: Detailed request logging must be disabled for performance
- **Rate Limiting**: Rate limit must not exceed 500 RPM to prevent abuse
- **Configuration Drift**: Validates production-specific security settings

#### Testing Environment Integrity Checks
- **WebSocket**: Must be disabled to prevent test interference
- **Metrics**: Must be disabled to avoid overhead during testing
- **Resource Limits**: Validates minimal resource allocations for testing

#### Development Environment Warnings
- **Request Logging**: Warning if detailed logging is disabled (recommended for debugging)
- **Security**: Relaxed security settings are noted but allowed

### Validation Error Examples

```go
// Production validation error
"CORS wildcard not allowed in production: environment=production, cors_origins=*"

// Rate limit validation error  
"rate limit too high for production: environment=production, rate_limit=600"

// Testing environment validation error
"WebSocket should be disabled in testing environment: environment=testing"
```

## Usage Examples

### Environment Detection

```go
cfg, err := config.InitConfig()
if err != nil {
    log.Fatal(err)
}

if cfg.IsProduction() {
    // Production-specific logic
} else if cfg.IsDevelopment() {
    // Development-specific logic
}
```

### Accessing Configuration

```go
// Server settings
port := cfg.Server.Port
readTimeout := cfg.Server.ReadTimeout

// Elevator settings
maxElevators := cfg.Elevator.MaxElevators
operationTimeout := cfg.Elevator.OperationTimeout

// Monitoring settings
if cfg.Monitoring.MetricsEnabled {
    // Enable metrics collection
}
```

### Environment Information

```go
envInfo := cfg.GetEnvironmentInfo()
// Returns map with environment summary for logging
```

## Best Practices

### Development
1. **Minimal Configuration**: Use default settings with only debug logging changes for consistency
2. **Debugging**: Enable debug logging (`LOG_LEVEL=DEBUG`) for troubleshooting
3. **Request Logging**: Enable detailed request logging (`LOG_REQUEST_DETAILS=true`) for API debugging
4. **Security**: Accept that security is relaxed (CORS wildcard) for convenience
5. **Consistency**: Keep most settings at defaults to match production behavior patterns

### Testing (Strictest Environment)
1. **Aggressive Validation**: Use very fast timeouts (500ms) to catch slow operations early
2. **Feature Control**: Disable unnecessary features (metrics, WebSocket) to avoid test interference
3. **Resource Limits**: Use minimal resources (5 elevators max) for controlled testing
4. **Fault Testing**: Aggressive circuit breaker (1 failure) to test fault handling thoroughly
5. **Fast Execution**: Ultra-fast operations (10ms floor duration) for rapid test cycles
6. **Issue Detection**: Use `WARN` level logging to surface problems without noise

### Production (High-Performance & Secure)
1. **Security First**: Use specific CORS origins, disable detailed request logging
2. **Performance Optimization**: Faster operations (200ms floor duration) and optimized timeouts (15s)
3. **Fault Tolerance**: Aggressive circuit breaker (2 failures) for fast fault detection
4. **Rate Limiting**: Strict rate limits (30 RPM) to prevent abuse while allowing legitimate use
5. **Scalability**: High capacity (200 elevators, 5000 WebSocket connections) for production load
6. **Monitoring**: Full observability with structured logging (`WARN` level for performance)

### Security & Compliance
1. **CORS Configuration**: Never use wildcard (`*`) in production environments
2. **Request Logging**: Disable detailed request logging in production for performance and privacy
3. **Rate Limiting**: Set appropriate limits based on expected load and abuse prevention
4. **Structured Logging**: Enable structured logging for audit trails and compliance
5. **Correlation IDs**: Use correlation IDs for request tracking and debugging
6. **Environment Validation**: Rely on automatic environment-specific validation

### Configuration Management
1. **Environment Variables**: Use environment-specific variable sets from recommendations
2. **Validation**: Test configuration validation during deployment
3. **Monitoring**: Monitor configuration drift and unauthorized changes
4. **Documentation**: Keep configuration documentation updated with changes
5. **Rollback Plan**: Have rollback procedures for configuration changes

## Troubleshooting

### Common Issues

**Configuration Loading Fails**
- Check environment variable syntax and formatting
- Verify all required variables are set
- Check for typos in variable names
- Ensure environment files have proper permissions

**Validation Errors**
- Review validation constraints and ranges in documentation
- Check numeric ranges and positive values
- Verify duration format (e.g., "30s", "5m", "1h")
- Check environment-specific validation rules

**Environment-Specific Issues**
- Ensure `ENV` variable is set to valid value (development, testing, production)
- Check environment file is being loaded properly
- Verify environment-specific overrides are applied
- Review environment-specific validation warnings/errors

**Production Environment Errors**
```bash
# CORS wildcard not allowed
"CORS wildcard not allowed in production: environment=production, cors_origins=*"
# Solution: Set specific origins
export CORS_ALLOWED_ORIGINS=https://app.example.com

# Rate limit too high
"rate limit too high for production: environment=production, rate_limit=600"
# Solution: Reduce rate limit
export RATE_LIMIT_RPM=60
```

**Testing Environment Issues**
```bash
# WebSocket should be disabled
"WebSocket should be disabled in testing environment: environment=testing"
# Solution: Disable WebSocket
export WEBSOCKET_ENABLED=false

# Metrics should be disabled
"Metrics should be disabled in testing environment: environment=testing"  
# Solution: Disable metrics
export METRICS_ENABLED=false
```

### Debugging Configuration

**Enable Comprehensive Logging**
```bash
export LOG_LEVEL=DEBUG
export LOG_REQUEST_DETAILS=true
```

**Verify Configuration Loading**
```bash
# Check health endpoint for environment info
curl http://localhost:6660/health

# Check current configuration with debug output
LOG_LEVEL=DEBUG ./server
```

**Validate Environment Variables**
```bash
# List all environment variables
env | grep -E "(ENV|LOG_|SERVER_|ELEVATOR_|RATE_|WEBSOCKET_|CIRCUIT_|CORS_)" | sort

# Test configuration parsing
ENV=testing go run cmd/server/main.go --validate-config
```

### Configuration Migration Issues

**Nested to Flat Structure Migration**
If you're migrating from the old nested configuration structure:

```go
// Old structure (no longer supported)
cfg.Server.ReadTimeout
cfg.Elevator.EachFloorDuration  
cfg.WebSocket.ConnectionTimeout
cfg.CircuitBreaker.MaxFailures

// New flat structure (current)
cfg.ReadTimeout
cfg.EachFloorDuration
cfg.WebSocketConnectionTimeout  
cfg.CircuitBreakerMaxFailures
```

**Environment Variable Migration**
```bash
# Old variables (deprecated)
READ_TIMEOUT -> SERVER_READ_TIMEOUT
CONNECTION_TIMEOUT -> WEBSOCKET_CONNECTION_TIMEOUT

# New variables (current)
SERVER_READ_TIMEOUT=30s
WEBSOCKET_CONNECTION_TIMEOUT=10m
```

## Migration from Previous Versions

### Breaking Changes
- Configuration structure is now nested (e.g., `cfg.Port` → `cfg.Server.Port`)
- Some environment variables have been renamed for consistency
- New required configuration fields with sensible defaults

### Migration Steps
1. Update application code to use nested configuration structure
2. Review and update environment variable names
3. Test with new configuration files
4. Update deployment scripts and infrastructure

## Monitoring Configuration Changes

### Production Monitoring
1. **Configuration Drift Detection**: Monitor for unauthorized configuration changes
2. **Performance Impact**: Track metrics when configuration changes are deployed
3. **Security Validation**: Alert on security-sensitive configuration changes (CORS, rate limits)
4. **Environment Consistency**: Compare configurations across environments

### Recommended Alerts
```yaml
# Rate limit exceeded
- alert: RateLimitExceeded
  expr: rate_limit_rpm > 500 AND environment == "production"
  
# CORS wildcard in production  
- alert: CORSWildcardProduction
  expr: cors_origins == "*" AND environment == "production"
  
# Configuration validation failure
- alert: ConfigValidationFailure
  expr: config_validation_errors > 0
```

## Summary of Key Features

### ✅ **Environment-Specific Configuration**
- **Development**: Default settings + enhanced debugging, minimal changes for consistency
- **Testing**: Strictest settings, catches issues early, aggressive validation, fast execution  
- **Production**: High-performance optimization, strict security, scalable configuration

### ✅ **Comprehensive Validation**
- **Basic Validation**: Data types, ranges, required fields
- **Environment-Specific**: Security checks for production, performance checks for testing
- **Startup Validation**: Fail-fast on invalid configuration
- **Runtime Monitoring**: Detect configuration drift

### ✅ **Security Enhancements**
- **CORS Protection**: Environment-specific origins, wildcard prevention in production
- **Rate Limiting**: Environment-appropriate limits with abuse prevention
- **Request Logging**: Privacy-aware logging control
- **Timeout Management**: Performance vs security balance

### ✅ **Operational Excellence**
- **Environment Detection**: Automatic environment-specific defaults
- **Configuration Matrix**: Clear comparison across environments
- **Migration Support**: Smooth transition from nested to flat structure
- **Troubleshooting**: Comprehensive debugging and validation tools

## Related Documentation
- [API Documentation](api_improvements.md)
- [Logging & Error Handling](logging_and_error_handling.md)
- [Monitoring & Observability](../README.md#monitoring--observability)
- [Type Safety](type_safety.md)
- [Concurrency & Performance](CONCURRENCY_PERFORMANCE.md) 