# Concurrency and Performance Architecture

This document provides a comprehensive overview of the concurrency mechanisms and performance optimizations implemented in the `concurrency-performance` branch of the elevator system.

## Table of Contents

1. [Overview](#overview)
2. [Concurrency Architecture](#concurrency-architecture)
3. [Performance Optimizations](#performance-optimizations)
4. [Circuit Breaker Pattern](#circuit-breaker-pattern)
5. [Synchronization Mechanisms](#synchronization-mechanisms)
6. [Performance Testing](#performance-testing)
7. [Monitoring and Metrics](#monitoring-and-metrics)
8. [Best Practices](#best-practices)

## Overview

The elevator system is designed as a highly concurrent, fault-tolerant application that can handle multiple elevator operations simultaneously while maintaining data consistency and optimal performance. The architecture leverages Go's built-in concurrency primitives and implements several performance optimization techniques.

### Key Design Principles

- **Concurrent by Default**: Every component is designed to handle concurrent operations safely
- **Non-Blocking Operations**: Minimize blocking operations to maintain system responsiveness
- **Fault Tolerance**: Circuit breaker pattern protects against cascading failures
- **Performance Monitoring**: Comprehensive metrics collection for performance analysis
- **Graceful Degradation**: System continues operating even when individual components fail

## Concurrency Architecture

### 1. Goroutine Management

#### Elevator Goroutines
```go
// Each elevator runs in its own goroutine with context cancellation
func (e *Elevator) switchOn() {
    for {
        select {
        case <-e.ctx.Done():
            e.logger.Info("elevator stopped due to context cancellation")
            return
        case <-e.switchOnChan:
            if e.directionsManager.UpDirectionLength() > 0 || e.directionsManager.DownDirectionLength() > 0 {
                e.runWithTimeout()
            }
        }
    }
}
```

**Features:**
- Each elevator operates independently in its own goroutine
- Context-based cancellation for graceful shutdown
- Buffered channels prevent blocking operations
- Timeout protection for long-running operations

#### Manager Concurrency
```go
// Concurrent elevator request processing with timeout protection
func (m *Manager) chooseElevatorWithTimeout(ctx context.Context, elevators []*elevator.Elevator, requestedDirection domain.Direction, fromFloor, toFloor domain.Floor) (*elevator.Elevator, error) {
    type result struct {
        elevator *elevator.Elevator
        err      error
    }

    resultCh := make(chan result, 1)
    go func() {
        e, err := m.chooseElevator(elevators, requestedDirection, fromFloor, toFloor)
        resultCh <- result{elevator: e, err: err}
    }()

    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    case result := <-resultCh:
        return result.elevator, result.err
    }
}
```

### 2. Channel Communication

#### Optimized Channel Usage
- **Zero-Memory Channels**: Using `chan struct{}` for signaling to minimize memory usage
- **Buffered Channels**: Prevent blocking in high-frequency operations
- **Timeout Channels**: Implement operation timeouts to prevent deadlocks

```go
// Buffered channel using struct{} for zero memory overhead
switchOnChan: make(chan struct{}, 10)

// Non-blocking push with context awareness
func (e *Elevator) pushWithContext() {
    select {
    case e.switchOnChan <- struct{}{}:
        // Signal sent successfully
    case <-e.ctx.Done():
        // Context was cancelled, don't block
        return
    default:
        // Channel is full, elevator is already processing, don't block
        e.logger.Debug("elevator is busy, skipping push signal")
    }
}
```

### 3. Context Management

#### Hierarchical Context Structure
- **Root Context**: Application-level cancellation
- **Operation Contexts**: Individual operation timeouts
- **Request Contexts**: HTTP request-scoped operations

```go
// Timeout context for elevator operations
func (e *Elevator) runWithTimeout() {
    ctx, cancel := context.WithTimeout(e.ctx, 30*time.Second)
    defer cancel()
    
    done := make(chan struct{})
    go func() {
        defer close(done)
        e.circuitBreaker.Execute(ctx, func() error {
            e.Run()
            return nil
        })
    }()
    
    select {
    case <-ctx.Done():
        e.logger.Warn("elevator operation timed out")
    case <-done:
        // Operation completed successfully
    }
}
```

## Performance Optimizations

### 1. Lock Optimization

#### Read-Write Mutexes
```go
type Manager struct {
    mu        sync.RWMutex  // Allows concurrent reads, exclusive writes
    elevators []*elevator.Elevator
    // ...
}

// Optimized read operations allow concurrent access
func (m *Manager) GetElevators() []*elevator.Elevator {
    m.mu.RLock()
    defer m.mu.RUnlock()
    elevators := make([]*elevator.Elevator, len(m.elevators))
    copy(elevators, m.elevators)  // Return a copy to prevent race conditions
    return elevators
}
```

#### Minimal Lock Scope
```go
// Reduced lock scope - only protecting critical sections
func (m *Manager) AddElevator(...) error {
    // Validation outside the lock
    if m.elevatorExists(name) {
        return domain.NewValidationError("elevator with this name already exists", nil)
    }
    
    // Create elevator outside the lock
    e, err := m.factory.CreateElevator(...)
    if err != nil {
        return err
    }
    
    // Minimal lock time for collection modification
    m.mu.Lock()
    m.elevators = append(m.elevators, e)
    m.mu.Unlock()
    
    return nil
}
```

### 2. Memory Optimization

#### Zero-Allocation Patterns
```go
// Using struct{} for signaling channels (zero memory)
switchOnChan: make(chan struct{}, 10)

// Slice pre-allocation to prevent repeated allocations
func (m *Manager) GetElevators() []*elevator.Elevator {
    m.mu.RLock()
    defer m.mu.RUnlock()
    elevators := make([]*elevator.Elevator, len(m.elevators))  // Pre-allocate exact size
    copy(elevators, m.elevators)
    return elevators
}
```

#### Efficient Data Structures
```go
// Optimized directions manager with efficient lookups
type Manager struct {
    mu   sync.RWMutex
    up   map[int][]int  // Floor -> destination floors
    down map[int][]int  // Floor -> destination floors
}

// O(1) lookup for request existence
func (d *Manager) IsRequestExisting(direction domain.Direction, from domain.Floor, to domain.Floor) bool {
    d.mu.RLock()
    defer d.mu.RUnlock()
    
    fromVal := from.Value()
    toVal := to.Value()
    
    if direction == domain.DirectionUp {
        return isValueInMapSlice(d.up, fromVal, toVal)
    }
    return isValueInMapSlice(d.down, fromVal, toVal)
}
```

### 3. HTTP Server Optimizations

#### Concurrent Request Handling
```go
// Each HTTP request is handled in its own goroutine by default
// Additional optimizations for WebSocket connections
func (s *Server) statusWebSocketHandler(w http.ResponseWriter, r *http.Request) {
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    defer ws.Close()
    
    // Separate goroutine for status updates with timeout protection
    wsCtx, wsCancel := context.WithTimeout(ctx, 10*time.Minute)
    defer wsCancel()
    
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-wsCtx.Done():
            return
        case <-ticker.C:
            // Non-blocking status update with timeout
            updateCtx, updateCancel := context.WithTimeout(wsCtx, 3*time.Second)
            statusCh := make(chan statusResult, 1)
            
            go func() {
                st, errS := s.manager.GetStatus()
                statusCh <- statusResult{status: st, err: errS}
            }()
            
            select {
            case <-updateCtx.Done():
                s.logger.Warn("status update timed out")
                updateCancel()
                continue
            case result := <-statusCh:
                // Process result
            }
            updateCancel()
        }
    }
}
```

## Circuit Breaker Pattern

### Implementation Overview

The circuit breaker pattern protects the elevator system from cascading failures by monitoring operation success/failure rates and temporarily blocking requests when the failure rate becomes too high.

### Circuit Breaker States

```go
type CircuitBreakerState int

const (
    StateClosed    CircuitBreakerState = iota  // Normal operation
    StateOpen                                  // Blocking requests
    StateHalfOpen                             // Testing recovery
)
```

#### 1. CLOSED State (Normal Operation)
- All requests are allowed through
- Failures are counted, successes reset the failure counter
- Transitions to OPEN when failure threshold is exceeded

#### 2. OPEN State (Failure Protection)
- All requests are immediately rejected without execution
- Prevents further load on the failing system
- After a timeout period, transitions to HALF-OPEN to test recovery

#### 3. HALF-OPEN State (Recovery Testing)
- Limited number of requests are allowed through to test system recovery
- If requests succeed, transitions back to CLOSED
- If any request fails, immediately returns to OPEN state

### Configuration and Usage

```go
func NewCircuitBreaker() *CircuitBreaker {
    return &CircuitBreaker{
        state:         StateClosed,
        maxFailures:   5,                // Open after 5 consecutive failures
        resetTimeout:  30 * time.Second, // Try to reset after 30 seconds
        halfOpenLimit: 3,                // Allow 3 requests in half-open state
    }
}

// Execute with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, operation func() error) error {
    if !cb.allowRequest() {
        return fmt.Errorf("circuit breaker is open - request rejected")
    }
    
    err := operation()
    
    if err != nil {
        cb.recordFailure()
        return err
    }
    
    cb.recordSuccess()
    return nil
}
```

### Integration with Elevator Operations

```go
func (e *Elevator) runWithTimeout() {
    ctx, cancel := context.WithTimeout(e.ctx, 30*time.Second)
    defer cancel()
    
    var operationErr error
    done := make(chan struct{})
    
    go func() {
        defer close(done)
        // Wrap elevator operation with circuit breaker protection
        operationErr = e.circuitBreaker.Execute(ctx, func() error {
            e.Run()
            return nil
        })
    }()
    
    select {
    case <-ctx.Done():
        e.logger.Warn("elevator operation timed out")
    case <-done:
        if operationErr != nil {
            state, failures, _ := e.circuitBreaker.GetMetrics()
            e.logger.Warn("elevator operation failed via circuit breaker",
                slog.String("circuit_breaker_state", e.getCircuitBreakerStateName(state)),
                slog.Int("failure_count", failures))
        }
    }
}
```

## Synchronization Mechanisms

### 1. Read-Write Mutexes

Used extensively for protecting data structures that have frequent reads and infrequent writes:

```go
// Directions manager uses RWMutex for optimal concurrent access
func (d *Manager) UpDirectionLength() int {
    d.mu.RLock()         // Multiple readers can access simultaneously
    defer d.mu.RUnlock()
    return len(d.up)
}

func (d *Manager) Append(direction domain.Direction, from, to domain.Floor) {
    d.mu.Lock()          // Exclusive access for writes
    defer d.mu.Unlock()
    // Modify data structure
}
```

### 2. Atomic Operations

For simple counters and flags where possible:

```go
// Metrics collection using atomic operations (when applicable)
var requestCount int64
atomic.AddInt64(&requestCount, 1)
```

### 3. Context-Based Cancellation

```go
// Hierarchical cancellation system
type Manager struct {
    ctx    context.Context
    cancel context.CancelFunc
}

func New(cfg *config.Config, factory factory.ElevatorFactory) *Manager {
    ctx, cancel := context.WithCancel(context.Background())
    return &Manager{
        ctx:    ctx,
        cancel: cancel,
        // ...
    }
}

// Graceful shutdown propagates through context
func (m *Manager) Shutdown() {
    elevators := m.GetElevators()
    for _, e := range elevators {
        e.Shutdown()  // Each elevator cancels its own context
    }
    
    if m.cancel != nil {
        m.cancel()    // Cancel manager context
    }
}
```

## Performance Testing

### 1. Benchmark Test Structure

The benchmark tests are organized to test different aspects of performance:

#### Elevator Benchmarks (`tests/benchmarks/elevator/`)

```go
// Memory usage benchmarks
func BenchmarkElevator_MemoryUsage(b *testing.B) {
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        elev, err := elevator.New("MemoryBenchmarkElevator", 0, 50, 
            time.Millisecond*10, time.Millisecond*10)
        if err != nil {
            b.Fatal(err)
        }
        
        // Add multiple requests to simulate real usage
        for j := 0; j < 10; j++ {
            elev.Request(domain.DirectionUp, domain.NewFloor(j), domain.NewFloor(j+5))
        }
        
        // Access various properties
        _ = elev.CurrentFloor()
        _ = elev.CurrentDirection()
        _ = elev.Directions()
    }
}

// Concurrent access benchmarks
func BenchmarkElevator_ConcurrentStateAccess(b *testing.B) {
    elev, err := elevator.New("ConcurrentStateBenchmarkElevator", 0, 100, 
        time.Millisecond*10, time.Millisecond*10)
    if err != nil {
        b.Fatal(err)
    }
    
    b.ResetTimer()
    b.ReportAllocs()
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            // Simulate concurrent reads of elevator state
            _ = elev.CurrentFloor()
            _ = elev.CurrentDirection()
            directions := elev.Directions()
            _ = directions.UpDirectionLength()
            _ = directions.DownDirectionLength()
        }
    })
}
```

#### Manager Benchmarks (`tests/benchmarks/manager/`)

```go
// Concurrent request processing benchmarks
func BenchmarkManager_ConcurrentRequests(b *testing.B) {
    ctx := context.Background()
    cfg := buildManagerTestConfig()
    elevatorFactory := &factory.StandardElevatorFactory{}
    mgr := manager.New(cfg, elevatorFactory)
    
    // Setup elevators
    for i := 0; i < 10; i++ {
        elevatorName := fmt.Sprintf("ConcurrentBenchmarkElevator%d", i)
        err := mgr.AddElevator(ctx, cfg, elevatorName, 0, 100, 
            time.Millisecond*10, time.Millisecond*10)
        if err != nil {
            b.Fatal(err)
        }
    }
    
    b.ResetTimer()
    b.ReportAllocs()
    
    b.RunParallel(func(pb *testing.PB) {
        counter := 0
        for pb.Next() {
            from := counter % 90
            to := from + 10
            _, err := mgr.RequestElevator(ctx, from, to)
            if err != nil {
                b.Logf("Request failed: %v", err)
            }
            counter++
        }
    })
}
```

### 2. Concurrent Testing

Race condition detection is built into the testing process:

```go
// Race detection test
func TestManager_ConcurrentOperations(t *testing.T) {
    const numGoroutines = 20
    const requestsPerGoroutine = 10
    
    var wg sync.WaitGroup
    var successCount int64
    var mu sync.Mutex
    
    wg.Add(numGoroutines)
    
    for i := 0; i < numGoroutines; i++ {
        go func(routineID int) {
            defer wg.Done()
            
            for j := 0; j < requestsPerGoroutine; j++ {
                from := routineID % 15
                to := from + 3
                
                _, err := manager.RequestElevator(ctx, from, to)
                if err == nil {
                    mu.Lock()
                    successCount++
                    mu.Unlock()
                }
            }
        }(i)
    }
    
    wg.Wait()
    
    // Verify no race conditions and good success rate
    assert.Greater(t, successCount, int64(numGoroutines*requestsPerGoroutine/2))
}
```

### 3. Test Execution

```bash
# Run all tests with race detection
make test/race

# Run specific benchmark tests
make test/benchmarks/elevator
make test/benchmarks/manager

# Run all benchmarks
make test/benchmarks
```

## Monitoring and Metrics

### 1. Health Metrics

Each elevator provides comprehensive health metrics including circuit breaker status:

```go
func (e *Elevator) GetHealthMetrics() map[string]interface{} {
    state, failures, successes := e.circuitBreaker.GetMetrics()
    
    return map[string]interface{}{
        "name":                      e.Name(),
        "current_floor":             e.CurrentFloor().Value(),
        "direction":                 string(e.CurrentDirection()),
        "pending_requests":          e.directionsManager.DirectionsLength(),
        "circuit_breaker_state":     e.getCircuitBreakerStateName(state),
        "circuit_breaker_failures":  failures,
        "circuit_breaker_successes": successes,
        "is_healthy":                state != StateOpen,
        "min_floor":                 e.state.MinFloor().Value(),
        "max_floor":                 e.state.MaxFloor().Value(),
    }
}
```

### 2. System Metrics

Manager-level metrics provide system-wide performance insights:

```go
func (m *Manager) GetMetrics() map[string]interface{} {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    totalRequests := 0
    totalUpRequests := 0
    totalDownRequests := 0
    
    for _, e := range m.elevators {
        directions := e.Directions()
        totalUpRequests += directions.UpDirectionLength()
        totalDownRequests += directions.DownDirectionLength()
    }
    totalRequests = totalUpRequests + totalDownRequests
    
    return map[string]interface{}{
        "total_elevators":     len(m.elevators),
        "total_requests":      totalRequests,
        "total_up_requests":   totalUpRequests,
        "total_down_requests": totalDownRequests,
        "average_load":        float64(totalRequests) / float64(max(len(m.elevators), 1)),
        "timestamp":           time.Now().Unix(),
    }
}
```

### 3. HTTP Metrics

Performance metrics are collected for HTTP operations:

```go
func requestDuration(elevatorName string, start time.Time) {
    duration := time.Since(start)
    metrics.HTTPRequestDuration.WithLabelValues(elevatorName).Observe(duration.Seconds())
}

func (s *Server) floorHandler(w http.ResponseWriter, r *http.Request) {
    startTime := time.Now()
    var elevatorName string
    defer func() {
        requestDuration(elevatorName, startTime)
    }()
    
    // Request processing...
}
```

### 4. Prometheus Integration

The system exposes Prometheus-compatible metrics:

```go
// Metrics endpoint
mux.Handle("/metrics", promhttp.Handler())

// Custom metrics endpoint
mux.HandleFunc("/metrics/system", s.systemMetricsHandler)
```

## Best Practices

### 1. Concurrency Guidelines

- **Prefer Channels over Shared Memory**: Use channels for communication between goroutines
- **Context Everywhere**: Always pass context for cancellation and timeouts
- **Minimal Lock Scope**: Keep critical sections as small as possible
- **Read-Write Locks**: Use RWMutex for read-heavy operations
- **Buffered Channels**: Use buffered channels to prevent blocking

### 2. Performance Guidelines

- **Pre-allocate Slices**: When the size is known, pre-allocate slices to prevent multiple allocations
- **Zero-Value Structs**: Use `struct{}` for signaling channels to minimize memory usage
- **Copy on Read**: Return copies of internal data structures to prevent race conditions
- **Circuit Breakers**: Protect against cascading failures in distributed operations
- **Timeout Everything**: Add timeouts to all potentially blocking operations

### 3. Testing Guidelines

- **Race Detection**: Always run tests with `-race` flag during development
- **Benchmark All Paths**: Create benchmarks for both sequential and concurrent operations
- **Load Testing**: Test with realistic concurrent loads
- **Memory Profiling**: Use `b.ReportAllocs()` in benchmarks to track memory usage
- **Timeout Tests**: Include timeout scenarios in testing

### 4. Monitoring Guidelines

- **Health Checks**: Implement comprehensive health checks including circuit breaker status
- **Metrics Collection**: Collect metrics at multiple levels (request, elevator, system)
- **Structured Logging**: Use structured logging with context for better observability
- **Error Tracking**: Track and categorize different types of errors
- **Performance Monitoring**: Monitor response times and throughput continuously

### 5. Error Handling Guidelines

- **Context Errors**: Handle context cancellation and timeouts appropriately
- **Circuit Breaker Integration**: Use circuit breakers for external dependencies
- **Graceful Degradation**: System should continue operating even when some components fail
- **Error Classification**: Classify errors (validation, internal, timeout, etc.) for appropriate handling
- **Logging and Metrics**: Log errors appropriately and update relevant metrics

## Conclusion

The `concurrency-performance` branch implements a robust, highly concurrent elevator system with comprehensive performance optimizations and fault tolerance mechanisms. The architecture leverages Go's concurrency primitives effectively while implementing industry-standard patterns like circuit breakers for resilience.

Key achievements:

- ✅ **Zero Race Conditions**: All tests pass with race detection enabled
- ✅ **Efficient Resource Usage**: Optimized memory allocation and lock usage
- ✅ **Fault Tolerance**: Circuit breaker pattern protects against failures
- ✅ **Comprehensive Testing**: Extensive benchmark and concurrent testing suite
- ✅ **Production-Ready Monitoring**: Full metrics and health check implementation
- ✅ **Scalable Architecture**: Designed to handle high concurrent load efficiently

The system is designed for production use with enterprise-level reliability, observability, and performance characteristics. 