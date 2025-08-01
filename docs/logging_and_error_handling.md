# Logging and Error Handling Improvements

## Overview

This document describes the logging and error handling improvements implemented in **step-2-logging** branch, building upon the architectural foundation established in **v2-major-refactor**. These changes implement structured logging and standardised error handling patterns across the elevator system.

## Logging System

### Structured JSON Logging

**Implementation**: Native Go `slog` with JSON output format for production observability.

```go
// Configuration via environment variable
type Config struct {
    LogLevel string `env:"LOG_LEVEL" envDefault:"INFO"`
}

// Logger initialisation with configurable levels
func InitLogger(logLevel string) {
    level := parseLogLevel(logLevel)
    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
    slog.SetDefault(slog.New(handler))
}
```

### Log Levels

| Level | Environment | Description |
|-------|-------------|-------------|
| `DEBUG` | Development | All messages including debug information |
| `INFO` | Production (default) | Informational messages and above |
| `WARN` | Production | Warning and error messages only |
| `ERROR` | Production | Error messages only |

### Correlation and Context

**Request correlation** for distributed tracing:

```go
func NewContextWithCorrelation(ctx context.Context) context.Context
func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
    ctx := logging.NewContextWithCorrelation(r.Context())
    slog.InfoContext(ctx, "request processed", slog.String("elevator", name))
}
```

**Component identification** for structured filtering:

```go
logger := slog.With(slog.String("component", "manager"))
logger.InfoContext(ctx, "elevator added", slog.String("elevator", name))
```

### JSON Output Format

```json
{
  "timestamp": "2025-06-23T22:53:19.399005+03:00",
  "level": "INFO",
  "message": "elevator added to management pool",
  "component": "manager",
  "elevator": "A",
  "min_floor": 0,
  "max_floor": 9,
  "correlation_id": "abc123..."
}
```

## Error Handling System

### Domain Error Types

**Structured error categorisation** with context support:

```go
type ErrType string
const (
    ErrTypeValidation ErrType = "validation"
    ErrTypeNotFound   ErrType = "not_found" 
    ErrTypeConflict   ErrType = "conflict"
    ErrTypeInternal   ErrType = "internal"
    ErrTypeExternal   ErrType = "external"
)

type DomainError struct {
    Type    ErrType
    Message string
    Err     error
    Context map[string]interface{}
}
```

### Predefined Domain Errors

```go
var (
    ErrElevatorNameEmpty  = NewValidationError("elevator name cannot be empty", nil)
    ErrElevatorFloorsSame = NewValidationError("minFloor and maxFloor cannot be equal", nil)
    ErrNoElevatorFound    = NewNotFoundError("no suitable elevator found for request", nil)
)
```

### Error Context Enrichment

```go
// Before: return fmt.Errorf("elevator %s not found", name)
// After:
return domain.ErrNoElevatorFound.WithContext("elevator_name", name)

// Before: return fmt.Errorf("floors must be different: %d", floor)  
// After:
return domain.ErrFloorsSame.WithContext("floor", floor)
```

### HTTP Status Mapping

**Automatic status code mapping** based on error types:

```go
statusCode := http.StatusInternalServerError
if domainErr, ok := err.(*domain.DomainError); ok {
    switch domainErr.Type {
    case domain.ErrTypeValidation: statusCode = http.StatusBadRequest
    case domain.ErrTypeNotFound:   statusCode = http.StatusNotFound
    case domain.ErrTypeConflict:   statusCode = http.StatusConflict
    }
}
```

## Migration from v2-major-refactor

### Key Changes Applied

1. **Configuration**: Replaced panic-based config loading with proper error handling
2. **Logging**: Added structured JSON logging with correlation IDs throughout
3. **Errors**: Implemented domain error system with type categorisation
4. **Context**: Added context.Context propagation for request tracing
5. **HTTP**: Added smart error status mapping for REST API responses

### Pattern Changes

```go
// Old patterns (v2-major-refactor)
return fmt.Errorf("name can't be empty")
slog.Error("elevator not created") 
panic("error on parsing env")

// New patterns (step-2-logging)
return domain.ErrElevatorNameEmpty
slog.ErrorContext(ctx, "failed to create elevator", slog.String("error", err.Error()))
if err := config.InitConfig(); err != nil { /* handle gracefully */ }
```

## Benefits

### Observability
- **Structured logs**: Compatible with ELK, Splunk, and other log aggregation systems
- **Request correlation**: End-to-end request tracing across components
- **Component isolation**: Easy filtering by service component

### Error Handling  
- **Type safety**: Compile-time error categorisation
- **Context preservation**: Rich error information through call chains
- **HTTP compatibility**: Proper REST API error responses

### Production Readiness
- **Configurable logging**: Environment-specific log levels
- **JSON format**: Machine-readable log output
- **Graceful degradation**: No application crashes on configuration errors

## Testing

Comprehensive test coverage includes:
- **Log level parsing**: All supported levels and edge cases
- **Domain errors**: Error creation, wrapping, and context enrichment  
- **HTTP mapping**: Status code assignment per error type

```bash
go test ./internal/infra/logging/... -v
go test ./internal/domain/... -v
go test ./internal/http/... -v
```

This implementation provides production-ready observability and error handling while maintaining the clean architecture established in the v2-major-refactor foundation. 