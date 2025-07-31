# Metrics and Monitoring

## Overview

The elevator control system implements comprehensive Prometheus-based metrics and health monitoring to ensure operational visibility and reliability.

## Architecture

- **Metrics Collection**: Prometheus-based metrics with custom collectors
- **Health Monitoring**: Multi-level health checks with caching
- **HTTP Middleware**: Automatic request tracking and performance monitoring
- **External Integration**: Support for DataDog, Elasticsearch, OpenTelemetry, and other observability platforms

## Core Metrics

### Request Processing
- `elevator_request_duration_seconds` - Request processing time (histogram)
- `elevator_requests_total` - Total requests by elevator, direction, and status (counter)
- `elevator_wait_time_seconds` - Passenger wait times (histogram)
- `elevator_travel_time_seconds` - Journey completion times (histogram)

### System Performance
- `elevator_efficiency_ratio` - Success rate per elevator (gauge)
- `elevator_current_floor` - Real-time floor position (gauge)
- `elevator_pending_requests` - Pending requests by direction (gauge)
- `elevator_memory_usage_bytes` - Memory utilization (gauge)
- `elevator_active_connections` - WebSocket connections (gauge)

### Circuit Breaker
- `elevator_circuit_breaker_state` - Circuit breaker status (gauge: 0=closed, 1=half-open, 2=open)
- `elevator_circuit_breaker_failures_total` - Circuit breaker failures (counter)

### HTTP Metrics
- `elevator_http_request_duration_seconds` - HTTP response times (histogram)
- `elevator_http_requests_total` - HTTP request count by method and status (counter)

### System Health
- `elevator_system_health` - Component health status (gauge: 1=healthy, 0=unhealthy)
- `elevator_errors_total` - Error count by type and component (counter)

## Health Monitoring

### Endpoints
- `/v1/health/live` - Basic liveness check
- `/v1/health/ready` - Readiness check with dependency validation
- `/v1/health/detailed` - Comprehensive health overview with metrics

### Health Checkers
- **System Resources**: Memory, goroutines, GC cycles (85% memory threshold, 1000 goroutine limit)
- **Manager**: Elevator manager operational status
- **Liveness**: Application uptime and responsiveness

### Health Status Levels
- **Healthy**: Component functioning normally
- **Degraded**: Issues present but functional
- **Unhealthy**: Component not functioning
- **Unknown**: Status cannot be determined

## Monitoring Endpoints

- `/metrics` - Prometheus metrics exposition
- `/v1/metrics` - JSON operational metrics from manager
- `/v1/health/*` - Health check endpoints

## Configuration

### Key Settings
- Health cache TTL: 30 seconds
- Memory threshold: 85%
- Goroutine threshold: 1000
- Health check timeout: 2 seconds

### External Integrations

#### Supported Platforms
- **DataDog**: Full metrics, traces, and logs
- **Elasticsearch/ELK**: Structured logging and storage
- **Prometheus**: Pull and push-based metrics
- **OpenTelemetry**: OTLP protocol support

#### Agent Auto-Detection
The system automatically detects and configures for observability agents:
- DataDog Agent (via `DD_API_KEY`)
- OpenTelemetry Collector (via `OTEL_EXPORTER_OTLP_ENDPOINT`)
- FluentBit (via `FLUENTD_HOST`)

#### Configuration Examples
```bash
# DataDog
DATADOG_ENABLED=true
DATADOG_API_KEY=your_api_key

# OpenTelemetry
OTLP_ENABLED=true
OTLP_ENDPOINT=http://otel-collector:4317

# Prometheus Push
PROMETHEUS_PUSH_ENABLED=true
PROMETHEUS_PUSH_GATEWAY=http://pushgateway:9091
```

## Best Practices

- Use consistent metric labels and avoid high cardinality
- Implement lightweight health checks with appropriate timeouts
- Cache health check results to minimize overhead
- Classify errors appropriately with structured logging
- Configure histogram buckets for your specific use case

## Testing

Comprehensive test coverage includes:
- Health check endpoint validation
- Metrics collection and recording
- HTTP performance monitoring
- Error rate tracking and panic recovery
- Benchmark tests for performance validation

The monitoring system provides complete observability for the elevator control system, enabling proactive monitoring, performance optimization, and reliable production operation. 