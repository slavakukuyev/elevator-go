package elevator

import (
	"context"
	"log/slog"
	"time"

	"github.com/slavakukuyev/elevator-go/internal/constants"
	"github.com/slavakukuyev/elevator-go/internal/directions"
	"github.com/slavakukuyev/elevator-go/internal/domain"
)

// Elevator represents an elevator with improved architecture and concurrency support
type Elevator struct {
	state             *State
	directionsManager *directions.Manager
	ctx               context.Context
	cancel            context.CancelFunc
	switchOnChan      chan struct{} // Channel for status updates - using struct{} for zero memory
	eachFloorDuration time.Duration
	openDoorDuration  time.Duration
	circuitBreaker    *CircuitBreaker // Circuit breaker for fault tolerance
	logger            *slog.Logger
	operationTimeout  time.Duration // Timeout for elevator operations
	overloadThreshold int           // Maximum number of requests before considering elevator overloaded
}

// New creates a new elevator instance with context support
func New(name string,
	minFloor, maxFloor int,
	eachFloorDuration, openDoorDuration, operationTimeout time.Duration,
	circuitBreakerMaxFailures int, circuitBreakerResetTimeout time.Duration, circuitBreakerHalfOpenLimit, overloadThreshold int) (*Elevator, error) {

	if name == "" {
		return nil, domain.NewValidationError("elevator name cannot be empty", nil)
	}

	if minFloor == maxFloor {
		return nil, domain.NewValidationError("minFloor and maxFloor cannot be equal", nil).
			WithContext("min_floor", minFloor).
			WithContext("max_floor", maxFloor)
	}

	minFloorDomain := domain.NewFloor(minFloor)
	maxFloorDomain := domain.NewFloor(maxFloor)

	logger := slog.With(
		slog.String("component", constants.ComponentElevator),
		slog.String("elevator_name", name),
	)

	ctx, cancel := context.WithCancel(context.Background())

	e := &Elevator{
		state:             NewState(name, minFloorDomain, maxFloorDomain),
		directionsManager: directions.New(),
		ctx:               ctx,
		cancel:            cancel,
		switchOnChan:      make(chan struct{}, 10), // Buffered channel using struct{} for zero memory
		eachFloorDuration: eachFloorDuration,
		openDoorDuration:  openDoorDuration,
		circuitBreaker:    NewCircuitBreaker(circuitBreakerMaxFailures, circuitBreakerResetTimeout, circuitBreakerHalfOpenLimit), // Initialize circuit breaker
		logger:            logger,
		operationTimeout:  operationTimeout,
		overloadThreshold: overloadThreshold,
	}

	// start read events process with context
	go e.switchOn()
	e.logger.Info("elevator created",
		slog.Int("min_floor", minFloor),
		slog.Int("max_floor", maxFloor),
		slog.Duration("floor_duration", eachFloorDuration),
		slog.Duration("door_duration", openDoorDuration))
	return e, nil
}

// Name returns the elevator name
func (e *Elevator) Name() string {
	return e.state.Name()
}

// SetName sets the elevator name
func (e *Elevator) SetName(name string) *Elevator {
	e.state.SetName(name)
	return e
}

// switchOn processes elevator events with context support
func (e *Elevator) switchOn() {
	for {
		select {
		case <-e.ctx.Done():
			e.logger.Info("elevator stopped due to context cancellation")
			return
		case <-e.switchOnChan:
			if e.directionsManager.HasUpRequests() || e.directionsManager.HasDownRequests() {
				e.runWithTimeout()
			}
		}
	}
}

// runWithTimeout executes the elevator movement logic with timeout
func (e *Elevator) runWithTimeout() {
	// Create a timeout context for this operation using configured timeout
	ctx, cancel := context.WithTimeout(e.ctx, e.operationTimeout)
	defer cancel()

	done := make(chan struct{})
	var operationErr error

	go func() {
		defer close(done)
		// Wrap Run operation with circuit breaker protection
		operationErr = e.circuitBreaker.Execute(ctx, func() error {
			e.Run()
			return nil // Run doesn't return error, but we can add error handling in future
		})
	}()

	select {
	case <-ctx.Done():
		e.logger.Warn("elevator operation timed out",
			slog.Duration("timeout", e.operationTimeout),
			slog.Int("current_floor", e.state.CurrentFloor().Value()))
	case <-done:
		if operationErr != nil {
			state, failures, _ := e.circuitBreaker.GetMetrics()
			e.logger.Warn("elevator operation failed via circuit breaker",
				slog.String("circuit_breaker_state", e.getCircuitBreakerStateName(state)),
				slog.Int("failure_count", failures),
				slog.String("error", operationErr.Error()))
		}
		// Operation completed successfully or failed gracefully
	}
}

// getCircuitBreakerStateName returns the string representation of circuit breaker state
func (e *Elevator) getCircuitBreakerStateName(state CircuitBreakerState) string {
	switch state {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Run executes the main elevator movement algorithm using SCAN/LOOK optimization.
// This function implements intelligent direction switching and request processing
// to optimize passenger service in real-world scenarios including high-rise buildings.
//
// Algorithm Flow:
// 1. Get current position and direction
// 2. Wait for floor transition duration (simulates real elevator movement)
// 3. Process requests in current direction with smart boundary handling
// 4. Handle direction changes when needed (LOOK algorithm optimization)
// 5. Manage idle state when no requests exist (energy efficient)
//
// The algorithm handles 7 main scenarios:
// - Moving up with up requests (normal upward traffic)
// - Moving down with down requests (normal downward traffic)
// - Direction change: down to up transition (efficient pickup)
// - Direction change: up to down transition (efficient pickup)
// - Overshot recovery: up direction (prevents stranded passengers)
// - Overshot recovery: down direction (prevents stranded passengers)
// - Idle state management (energy efficient when no work)
func (e *Elevator) Run() {
	currentFloor := e.state.CurrentFloor()
	direction := e.state.Direction()

	e.logger.Debug("processing elevator movement",
		slog.Int("current_floor", currentFloor.Value()),
		slog.String("direction", string(direction)))

	// Simulate real elevator movement time between floors
	// This prevents the algorithm from running too fast and allows for
	// realistic timing in the simulation
	select {
	case <-e.ctx.Done():
		return
	case <-time.After(e.eachFloorDuration):
		// Continue with normal operation
	}

	// SCENARIO 1: Boundary handling - Top floor with up direction but no up requests
	// This prevents the elevator from getting stuck at the top floor
	// when there are no more upward requests but downward requests exist
	if direction == domain.DirectionUp && e.state.IsAtTopFloor() && !e.directionsManager.HasUpRequests() {
		if e.directionsManager.HasDownRequests() {
			e.state.SetDirection(domain.DirectionDown)
			e.pushWithContext()
			return
		}
		// No requests in either direction, will transition to idle at the end
	}

	// SCENARIO 2: Moving up with up requests (normal upward traffic)
	// This handles the primary case of serving passengers going up
	// including high-rise building scenarios and normal upward traffic
	if direction == domain.DirectionUp && e.directionsManager.HasUpRequests() {
		// Service current floor if passengers requested it
		// This handles both pickup and dropoff at the current floor
		if e.directionsManager.HasUpFloor(currentFloor.Value()) {
			e.openDoor()
			e.directionsManager.Flush(direction, currentFloor)
			e.closeDoor()
		}

		// Boundary handling: Check if elevator reached the top floor
		// This implements smart direction switching at building boundaries
		if e.state.IsAtTopFloor() {
			// Priority 1: Check for down requests first (LOOK algorithm)
			if e.directionsManager.HasDownRequests() {
				e.state.SetDirection(domain.DirectionDown)
				e.pushWithContext()
				return
			}
			// Priority 2: Check for up requests below current floor that need pickup
			// This handles cases where passengers are waiting on lower floors
			if e.directionsManager.HasUpRequests() {
				smallest, hasKey := e.directionsManager.GetSmallestUpKey()
				if hasKey {
					smallestFloor := domain.NewFloor(smallest)
					if smallestFloor.IsBelow(currentFloor) {
						e.state.SetDirection(domain.DirectionDown)
						e.pushWithContext()
						return
					}
				}
			}
			// If no requests requiring downward movement, transition to idle and continue to idle check
		}

		// Continue moving up if there are more up requests above current floor
		// This implements the core SCAN algorithm - continue in one direction until all requests served
		if e.shouldMoveUp() {
			newFloor := domain.NewFloor(currentFloor.Value() + 1)
			e.state.SetCurrentFloor(newFloor)
			e.pushWithContext()
			return
		}
	}

	// SCENARIO 3: Boundary handling - Bottom floor with down direction but no down requests
	// This prevents the elevator from getting stuck at the bottom floor
	// when there are no more downward requests but upward requests exist
	// Important for underground parking scenarios
	if direction == domain.DirectionDown && e.state.IsAtBottomFloor() && !e.directionsManager.HasDownRequests() {
		if e.directionsManager.HasUpRequests() {
			e.state.SetDirection(domain.DirectionUp)
			e.pushWithContext()
			return
		}
		// No requests in either direction, will transition to idle at the end
	}

	// SCENARIO 4: Moving down with down requests (normal downward traffic)
	// This handles the primary case of serving passengers going down
	// including underground parking scenarios and normal downward traffic
	if direction == domain.DirectionDown && e.directionsManager.HasDownRequests() {
		// Service current floor if passengers requested it
		// This handles both pickup and dropoff at the current floor
		if e.directionsManager.HasDownFloor(currentFloor.Value()) {
			e.openDoor()
			e.directionsManager.Flush(direction, currentFloor)
			e.closeDoor()
		}

		// Boundary handling: Check if elevator reached the bottom floor
		// This implements smart direction switching at building boundaries
		if e.state.IsAtBottomFloor() {
			// Priority 1: Check for up requests first (LOOK algorithm)
			if e.directionsManager.HasUpRequests() {
				e.state.SetDirection(domain.DirectionUp)
				e.pushWithContext()
				return
			}
			// Priority 2: Check for down requests above current floor that need pickup
			// This handles cases where passengers are waiting on upper floors
			if e.directionsManager.HasDownRequests() {
				largest, hasKey := e.directionsManager.GetLargestDownKey()
				if hasKey {
					largestFloor := domain.NewFloor(largest)
					if largestFloor.IsAbove(currentFloor) {
						e.state.SetDirection(domain.DirectionUp)
						e.pushWithContext()
						return
					}
				}
			}
			// If no requests requiring upward movement, transition to idle and continue to idle check
		}

		// Continue moving down if there are more down requests below current floor
		// This implements the core SCAN algorithm - continue in one direction until all requests served
		if e.shouldMoveDown() {
			newFloor := domain.NewFloor(currentFloor.Value() - 1)
			e.state.SetCurrentFloor(newFloor)
			e.pushWithContext()
			return
		}
	}

	// SCENARIO 5: Direction change - Down to Up transition (LOOK algorithm optimization)
	// This handles the case where elevator is moving down but has no more down requests
	// but there are up requests that need to be picked up
	// This implements efficient direction switching to minimize passenger wait time
	if direction == domain.DirectionDown && e.directionsManager.HasUpRequests() {
		smallest, hasKey := e.directionsManager.GetSmallestUpKey()
		if hasKey {
			smallestFloor := domain.NewFloor(smallest)
			// Continue moving down if the smallest up request is below current floor
			// This ensures we don't change direction prematurely
			if smallestFloor.IsBelow(currentFloor) {
				newFloor := domain.NewFloor(currentFloor.Value() - 1)
				e.state.SetCurrentFloor(newFloor)
				e.pushWithContext()
				return
			}

			// Change direction to up if we're already at the pickup floor
			// This handles immediate direction change when elevator reaches pickup point
			if smallestFloor.IsEqual(currentFloor) {
				e.state.SetDirection(domain.DirectionUp)
				e.pushWithContext()
				return
			}

			// If the smallest up request is above current floor, change direction to up
			// This prevents the elevator from moving past pickup floors
			if smallestFloor.IsAbove(currentFloor) {
				e.state.SetDirection(domain.DirectionUp)
				e.pushWithContext()
				return
			}
		}
	}

	// SCENARIO 6: Direction change - Up to Down transition (LOOK algorithm optimization)
	// This handles the case where elevator is moving up but has no more up requests
	// but there are down requests that need to be picked up
	// This implements efficient direction switching to minimize passenger wait time
	if direction == domain.DirectionUp && e.directionsManager.HasDownRequests() {
		largest, hasKey := e.directionsManager.GetLargestDownKey()
		if hasKey {
			largestFloor := domain.NewFloor(largest)
			// Continue moving up if the largest down request is above current floor
			// This ensures we don't change direction prematurely
			if largestFloor.IsAbove(currentFloor) {
				newFloor := domain.NewFloor(currentFloor.Value() + 1)
				e.state.SetCurrentFloor(newFloor)
				e.pushWithContext()
				return
			}

			// Change direction to down if we're already at the pickup floor
			// This handles immediate direction change when elevator reaches pickup point
			if largestFloor.IsEqual(currentFloor) {
				e.state.SetDirection(domain.DirectionDown)
				e.pushWithContext()
				return
			}

			// If the largest down request is below current floor, change direction to down
			// This prevents the elevator from moving past pickup floors
			if largestFloor.IsBelow(currentFloor) {
				e.state.SetDirection(domain.DirectionDown)
				e.pushWithContext()
				return
			}
		}
	}

	// SCENARIO 7: Overshot recovery - Up direction (prevents stranded passengers)
	// This handles the edge case where elevator is moving up but has overshot all up requests
	// This prevents passengers from being stranded when elevator moves past their floor
	if direction == domain.DirectionUp && e.directionsManager.HasUpRequests() {
		largest, hasKey := e.directionsManager.GetLargestUpKey()
		if hasKey {
			largestFloor := domain.NewFloor(largest)
			// If all up requests are below current floor, we've overshot
			// Change direction to down to return and serve missed requests
			if largestFloor.IsBelow(currentFloor) {
				e.state.SetDirection(domain.DirectionDown)
				e.pushWithContext()
				return
			}
		}
	}

	// SCENARIO 8: Overshot recovery - Down direction (prevents stranded passengers)
	// This handles the edge case where elevator is moving down but has overshot all down requests
	// This prevents passengers from being stranded when elevator moves past their floor
	if direction == domain.DirectionDown && e.directionsManager.HasDownRequests() {
		smallest, hasKey := e.directionsManager.GetSmallestDownKey()
		if hasKey {
			smallestFloor := domain.NewFloor(smallest)
			// If all down requests are above current floor, we've overshot
			// Change direction to up to return and serve missed requests
			if smallestFloor.IsAbove(currentFloor) {
				e.state.SetDirection(domain.DirectionUp)
				e.pushWithContext()
				return
			}
		}
	}

	// SCENARIO 9: Idle state management (energy efficient when no work)
	// This is the final check - if no requests exist in either direction
	// the elevator enters an idle state to save energy
	// This is normal behavior and not a bug - elevators should be idle when no work exists
	if e.directionsManager.IsIdle() {
		e.state.SetDirection(domain.DirectionIdle)
		e.logger.Debug("elevator stopped and has empty requests for both directions", slog.Int("floor", e.state.CurrentFloor().Value()))
	}
}

// shouldMoveUp determines if the elevator should continue moving up in the current direction.
// This implements the core SCAN algorithm principle: continue in one direction until all requests served.
//
// Returns true if:
// - There are up requests pending AND
// - The largest up request floor is above the current floor
//
// This prevents unnecessary direction changes and ensures efficient upward movement
// by continuing to the highest requested floor before considering direction changes.
func (e *Elevator) shouldMoveUp() bool {
	if e.directionsManager.HasUpRequests() {
		largest, hasKey := e.directionsManager.GetLargestUpKey()
		if hasKey {
			largestFloor := domain.NewFloor(largest)
			return largestFloor.IsAbove(e.state.CurrentFloor())
		}
	}
	return false
}

// shouldMoveDown determines if the elevator should continue moving down in the current direction.
// This implements the core SCAN algorithm principle: continue in one direction until all requests served.
//
// Returns true if:
// - There are down requests pending AND
// - The smallest down request floor is below the current floor
//
// This prevents unnecessary direction changes and ensures efficient downward movement
// by continuing to the lowest requested floor before considering direction changes.
// Important for underground parking scenarios where elevators serve basement levels.
func (e *Elevator) shouldMoveDown() bool {
	if e.directionsManager.HasDownRequests() {
		smallest, hasKey := e.directionsManager.GetSmallestDownKey()
		if hasKey {
			smallestFloor := domain.NewFloor(smallest)
			return smallestFloor.IsBelow(e.state.CurrentFloor())
		}
	}
	return false
}

func (e *Elevator) openDoor() {
	e.logger.Info("elevator doors operation",
		slog.String("action", "open"),
		slog.Int("floor", e.state.CurrentFloor().Value()))

	// Use context-aware sleep for door operations
	select {
	case <-e.ctx.Done():
		return
	case <-time.After(e.openDoorDuration):
		// Continue with normal operation
	}
}

func (e *Elevator) closeDoor() {
	e.logger.Info("elevator doors operation",
		slog.String("action", "close"),
		slog.Int("floor", e.state.CurrentFloor().Value()))
}

// Request adds a new elevator request
func (e *Elevator) Request(direction domain.Direction, fromFloor, toFloor domain.Floor) {
	currentDirection := e.state.Direction()
	if currentDirection == domain.DirectionIdle {
		setDirection := direction
		currentFloor := e.state.CurrentFloor()
		if direction == domain.DirectionDown && currentFloor.IsBelow(fromFloor) {
			setDirection = domain.DirectionUp
		} else if direction == domain.DirectionUp && currentFloor.IsAbove(fromFloor) {
			setDirection = domain.DirectionDown
		}

		e.state.SetDirection(setDirection)
	}

	e.directionsManager.Append(direction, fromFloor, toFloor)
	e.logger.Info("new elevator request received",
		slog.String("direction", string(direction)),
		slog.Int("from_floor", fromFloor.Value()),
		slog.Int("to_floor", toFloor.Value()),
		slog.String("current_direction", string(currentDirection)))
	e.pushWithContext()
}

// CurrentDirection returns the current direction
func (e *Elevator) CurrentDirection() domain.Direction {
	return e.state.Direction()
}

// CurrentFloor returns the current floor
func (e *Elevator) CurrentFloor() domain.Floor {
	return e.state.CurrentFloor()
}

// Directions returns the directions manager with optimized locking
func (e *Elevator) Directions() *directions.Manager {
	// Reduced mutex scope - only protecting the return of the manager reference
	return e.directionsManager
}

// pushWithContext triggers the elevator to process requests with context awareness
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

// Shutdown gracefully shuts down the elevator
func (e *Elevator) Shutdown() {
	e.logger.Info("shutting down elevator")
	if e.cancel != nil {
		e.cancel()
	}
}

// IsRequestInRange checks if the request is within the elevator's range
func (e *Elevator) IsRequestInRange(fromFloor, toFloor domain.Floor) bool {
	isInRange := e.state.IsFloorInRange(fromFloor) && e.state.IsFloorInRange(toFloor)
	if !isInRange {
		e.logger.Warn("request floor out of range",
			slog.Int("from_floor", fromFloor.Value()),
			slog.Int("to_floor", toFloor.Value()),
			slog.Int("min_floor", e.state.MinFloor().Value()),
			slog.Int("max_floor", e.state.MaxFloor().Value()))
	}
	return isInRange
}

func (e *Elevator) MinFloor() domain.Floor {
	return e.state.MinFloor()
}

func (e *Elevator) MaxFloor() domain.Floor {
	return e.state.MaxFloor()
}

// OverloadThreshold returns the elevator's overload threshold
func (e *Elevator) OverloadThreshold() int {
	return e.overloadThreshold
}

func (e *Elevator) GetStatus() domain.ElevatorStatus {
	requestCount := e.directionsManager.DirectionsLength()
	return e.state.GetStatus(requestCount)
}

// GetHealthMetrics returns health metrics including circuit breaker status
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
