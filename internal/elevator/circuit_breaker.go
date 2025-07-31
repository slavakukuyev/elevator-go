package elevator

// circuit_breaker.go implements the Circuit Breaker pattern for elevator operations.
//
// PURPOSE:
// The circuit breaker protects the elevator system from cascading failures by monitoring
// operation success/failure rates and temporarily blocking requests when the failure rate
// becomes too high. This prevents overwhelming a failing subsystem and allows it time to recover.
//
// FUNCTIONALITY:
// The circuit breaker operates in three states:
//
// 1. CLOSED (Normal Operation):
//    - All requests are allowed through
//    - Failures are counted, successes reset the failure counter
//    - Transitions to OPEN when failure threshold is exceeded
//
// 2. OPEN (Failure Protection):
//    - All requests are immediately rejected without execution
//    - Prevents further load on the failing system
//    - After a timeout period, transitions to HALF-OPEN to test recovery
//
// 3. HALF-OPEN (Recovery Testing):
//    - Limited number of requests are allowed through to test system recovery
//    - If requests succeed, transitions back to CLOSED
//    - If any request fails, immediately returns to OPEN state
//
// ELEVATOR CONTEXT:
// In the elevator system, this protects against scenarios like:
// - Hardware failures (motor, sensors, doors)
// - Communication timeouts with external systems
// - Overload conditions that could cause system instability
//
// The circuit breaker ensures that when elevator operations start failing,
// the system gracefully degrades rather than continuing to attempt operations
// that are likely to fail, potentially causing more damage or instability.

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	// StateClosed means the circuit breaker is closed and allowing requests
	StateClosed CircuitBreakerState = iota
	// StateOpen means the circuit breaker is open and rejecting requests
	StateOpen
	// StateHalfOpen means the circuit breaker is allowing limited requests to test recovery
	StateHalfOpen
)

// CircuitBreaker implements a circuit breaker pattern for elevator operations
type CircuitBreaker struct {
	mu           sync.RWMutex
	state        CircuitBreakerState
	failureCount int
	successCount int
	lastFailTime time.Time
	nextRetry    time.Time

	// Configuration
	maxFailures   int
	resetTimeout  time.Duration
	halfOpenLimit int
}

// NewCircuitBreaker creates a new circuit breaker with configurable settings
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration, halfOpenLimit int) *CircuitBreaker {
	return &CircuitBreaker{
		state:         StateClosed,
		maxFailures:   maxFailures,
		resetTimeout:  resetTimeout,
		halfOpenLimit: halfOpenLimit,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, operation func() error) error {
	if !cb.allowRequest() {
		return fmt.Errorf("circuit breaker is open - request rejected")
	}

	// Execute the operation
	err := operation()

	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// allowRequest determines if a request should be allowed based on circuit breaker state
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Now().After(cb.nextRetry) {
			cb.state = StateHalfOpen
			cb.successCount = 0
			return true
		}
		return false
	case StateHalfOpen:
		return cb.successCount < cb.halfOpenLimit
	default:
		return false
	}
}

// recordSuccess records a successful operation
func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0

	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.halfOpenLimit {
			// Enough successful requests, close the circuit
			cb.state = StateClosed
		}
	}
}

// recordFailure records a failed operation
func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailTime = time.Now()

	if cb.state == StateHalfOpen {
		// Failure in half-open state, go back to open
		cb.state = StateOpen
		cb.nextRetry = time.Now().Add(cb.resetTimeout)
	} else if cb.failureCount >= cb.maxFailures {
		// Too many failures, open the circuit
		cb.state = StateOpen
		cb.nextRetry = time.Now().Add(cb.resetTimeout)
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetMetrics returns current metrics of the circuit breaker
func (cb *CircuitBreaker) GetMetrics() (state CircuitBreakerState, failures int, successes int) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state, cb.failureCount, cb.successCount
}
