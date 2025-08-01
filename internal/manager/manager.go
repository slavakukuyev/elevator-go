package manager

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/slavakukuyev/elevator-go/internal/constants"
	"github.com/slavakukuyev/elevator-go/internal/domain"
	"github.com/slavakukuyev/elevator-go/internal/elevator"
	"github.com/slavakukuyev/elevator-go/internal/factory"
	"github.com/slavakukuyev/elevator-go/internal/infra/config"
	"github.com/slavakukuyev/elevator-go/metrics"
)

type Manager struct {
	mu        sync.RWMutex
	elevators []*elevator.Elevator
	factory   factory.ElevatorFactory
	logger    *slog.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	cfg       *config.Config
}

func New(cfg *config.Config, factory factory.ElevatorFactory) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		elevators: make([]*elevator.Elevator, 0),
		factory:   factory,
		logger:    slog.With(slog.String("component", constants.ComponentManager)),
		ctx:       ctx,
		cancel:    cancel,
		cfg:       cfg,
	}
}

func (m *Manager) AddElevator(ctx context.Context, cfg *config.Config, name string,
	minFloor, maxFloor int,
	eachFloorDuration, openDoorDuration time.Duration, overloadThreshold int) error {

	// Create a timeout context for elevator creation using configuration
	createCtx, cancel := context.WithTimeout(ctx, m.cfg.CreateElevatorTimeout)
	defer cancel()

	// Check if elevator with the same name already exists - optimized lock scope
	if m.elevatorExists(name) {
		err := domain.NewValidationError("elevator with this name already exists", nil).
			WithContext("name", name)
		m.logger.ErrorContext(createCtx, "failed to add elevator",
			slog.String("name", name),
			slog.String("error", err.Error()))
		return err
	}

	e, err := m.factory.CreateElevator(cfg, name,
		minFloor, maxFloor,
		eachFloorDuration, openDoorDuration, overloadThreshold)
	if err != nil {
		m.logger.ErrorContext(createCtx, "failed to initialize new elevator",
			slog.String("name", name),
			slog.Int("minFloor", minFloor),
			slog.Int("maxFloor", maxFloor),
			slog.String("error", err.Error()))

		// Preserve validation errors, wrap others as internal errors
		if domainErr, ok := err.(*domain.DomainError); ok && domainErr.Type == domain.ErrTypeValidation {
			return err
		}

		return domain.NewInternalError("failed to initialize new elevator", err).
			WithContext("name", name).
			WithContext("minFloor", minFloor).
			WithContext("maxFloor", maxFloor)
	}

	// Add to the collection with minimal lock time
	m.mu.Lock()
	m.elevators = append(m.elevators, e)
	m.mu.Unlock()

	m.logger.InfoContext(createCtx, "new elevator added to the management pool",
		slog.String("elevator", e.Name()),
		slog.Int("minFloor", e.MinFloor().Value()),
		slog.Int("maxFloor", e.MaxFloor().Value()),
	)
	return nil
}

// elevatorExists checks if an elevator with the given name already exists
func (m *Manager) elevatorExists(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, e := range m.elevators {
		if e.Name() == name {
			return true
		}
	}
	return false
}

func (m *Manager) GetElevator(name string) *elevator.Elevator {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, e := range m.elevators {
		if e.Name() == name {
			return e
		}
	}
	return nil
}

func (m *Manager) GetElevators() []*elevator.Elevator {
	m.mu.RLock()
	defer m.mu.RUnlock()
	elevators := make([]*elevator.Elevator, len(m.elevators))
	copy(elevators, m.elevators)
	return elevators
}

func (m *Manager) RequestElevator(ctx context.Context, fromFloor, toFloor int) (*elevator.Elevator, error) {
	start := time.Now()

	// Create a timeout context for elevator request processing using configuration
	requestCtx, cancel := context.WithTimeout(ctx, m.cfg.RequestTimeout)
	defer cancel()

	if toFloor == fromFloor {
		err := domain.NewValidationError("requested floor must be different from current floor", nil).
			WithContext("fromFloor", fromFloor).
			WithContext("toFloor", toFloor)
		m.logger.ErrorContext(requestCtx, "invalid floor request",
			slog.Int("fromFloor", fromFloor),
			slog.Int("toFloor", toFloor),
			slog.String("error", err.Error()))

		// Record validation error
		metrics.IncError("validation_error", "manager")
		return nil, err
	}

	direction := domain.DirectionUp
	if toFloor < fromFloor {
		direction = domain.DirectionDown
	}

	fromFloorDomain := domain.NewFloor(fromFloor)
	toFloorDomain := domain.NewFloor(toFloor)

	// Get a snapshot of elevators to reduce lock time
	elevators := m.GetElevators()

	if len(elevators) == 0 {
		err := domain.NewInternalError("no elevators created yet", nil)
		m.logger.ErrorContext(requestCtx, "no elevators available",
			slog.Int("fromFloor", fromFloor),
			slog.Int("toFloor", toFloor),
			slog.String("error", err.Error()))
		return nil, err
	}

	var el *elevator.Elevator

	if len(elevators) == 1 {
		el = elevators[0]
		if !el.IsRequestInRange(fromFloorDomain, toFloorDomain) {
			return nil, domain.NewValidationError("requested floors out of range for the elevator", nil).
				WithContext("fromFloor", fromFloor).
				WithContext("toFloor", toFloor)
		}
	}

	if el == nil {
		// validate existing requests
		if el = requestedElevator(elevators, direction, fromFloorDomain, toFloorDomain); el != nil {
			m.logger.InfoContext(requestCtx, "found existing elevator request",
				slog.String("elevator", el.Name()),
				slog.Int("fromFloor", fromFloor),
				slog.Int("toFloor", toFloor))

			// Record existing request metrics
			duration := time.Since(start)
			metrics.RecordRequestDuration(el.Name(), "existing", duration.Seconds())
			return el, nil
		}
	}

	if el == nil {
		var err error
		el, err = m.chooseElevatorWithTimeout(requestCtx, elevators, direction, fromFloorDomain, toFloorDomain)
		if err != nil {
			m.logger.ErrorContext(requestCtx, "failed to choose elevator",
				slog.Int("fromFloor", fromFloor),
				slog.Int("toFloor", toFloor),
				slog.String("error", err.Error()))

			// Record elevator selection failure
			metrics.IncError("elevator_selection_failed", "manager")
			return nil, domain.NewNotFoundError("no suitable elevator found", err).
				WithContext("fromFloor", fromFloor).
				WithContext("toFloor", toFloor)
		}
	}

	// Safety check: ensure elevator is not nil
	if el == nil {
		err := domain.NewInternalError("elevator selection returned nil without error", nil)
		m.logger.ErrorContext(requestCtx, "internal error: nil elevator returned",
			slog.Int("fromFloor", fromFloor),
			slog.Int("toFloor", toFloor))

		metrics.IncError("nil_elevator_selection", "manager")
		return nil, err
	}

	el.Request(direction, fromFloorDomain, toFloorDomain)

	// Record successful request metrics
	duration := time.Since(start)
	directionStr := string(direction)

	metrics.RecordRequestDuration(el.Name(), "success", duration.Seconds())
	metrics.IncRequestsTotal(el.Name(), directionStr, "success")

	// Calculate estimated wait time (simplified estimation)
	currentFloor := el.CurrentFloor().Value()
	waitTimeEstimate := float64(abs(fromFloor-currentFloor)) * 2.0 // 2 seconds per floor estimate
	metrics.RecordWaitTime(el.Name(), waitTimeEstimate)

	// Calculate travel time estimate
	travelDistance := abs(toFloor - fromFloor)
	travelTimeEstimate := float64(travelDistance) * 2.0 // 2 seconds per floor
	metrics.RecordTravelTime(el.Name(), fmt.Sprintf("%d", travelDistance), travelTimeEstimate)

	m.logger.InfoContext(requestCtx, "request has been approved",
		slog.String("elevator", el.Name()),
		slog.Int("fromFloor", fromFloor),
		slog.Int("toFloor", toFloor),
		slog.Float64("processing_time_seconds", duration.Seconds()),
		slog.Float64("estimated_wait_time", waitTimeEstimate))
	return el, nil
}

func requestedElevator(elevators []*elevator.Elevator, direction domain.Direction, fromFloor, toFloor domain.Floor) *elevator.Elevator {
	for _, e := range elevators {
		if e.Directions().IsRequestExisting(direction, fromFloor, toFloor) {
			return e
		}
	}

	return nil
}

// chooseElevatorWithTimeout wraps chooseElevator with timeout support
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
		m.logger.Error("elevator selection timed out!!!",
			slog.String("error", ctx.Err().Error()),
			slog.Int("fromFloor", fromFloor.Value()),
			slog.Int("toFloor", toFloor.Value()))
		return nil, domain.NewInternalError("elevator selection timed out", ctx.Err())
	case res := <-resultCh:
		return res.elevator, res.err
	}
}

func (m *Manager) chooseElevator(elevators []*elevator.Elevator, requestedDirection domain.Direction, fromFloor, toFloor domain.Floor) (*elevator.Elevator, error) {
	elevatorsWaiting := make(map[*elevator.Elevator]domain.Floor)
	elevatorsByDirection := make(map[*elevator.Elevator]domain.Direction)

	// case when elevator is waiting to start
	for _, e := range elevators {
		if !e.IsRequestInRange(fromFloor, toFloor) {
			continue
		}

		d := e.CurrentDirection()
		if d == domain.DirectionIdle {
			elevatorsWaiting[e] = e.CurrentFloor()
		} else {
			elevatorsByDirection[e] = d
		}
	}

	if len(elevatorsWaiting) > 0 {
		if e := findNearestElevator(elevatorsWaiting, fromFloor); e != nil {
			m.logger.Debug("found nearest elevator in waiting state!!!",
				slog.String("elevator", e.Name()),
				slog.Int("fromFloor", fromFloor.Value()),
				slog.Int("toFloor", toFloor.Value()))
			return e, nil
		}
	}

	if len(elevatorsByDirection) == 0 {
		m.logger.Debug("no elevators in the same direction!!!",
			slog.Int("fromFloor", fromFloor.Value()),
			slog.Int("toFloor", toFloor.Value()))
		return nil, domain.NewValidationError("requested floors out of range for all elevators", nil).
			WithContext("fromFloor", fromFloor.Value()).
			WithContext("toFloor", toFloor.Value())
	}

	/******** SAME DIRECTION ********/

	filteredElevators := elevatorsMatchingDirections(elevatorsByDirection, requestedDirection)

	// case when single elevator with the same direction
	// should validate if the elevator still on his way to the floor and not overloaded
	if len(filteredElevators) == 1 {
		e := filteredElevators[0]
		currentFloor := e.CurrentFloor()

		if ((requestedDirection == domain.DirectionUp && (currentFloor.IsBelow(fromFloor) || currentFloor.IsEqual(fromFloor))) ||
			(requestedDirection == domain.DirectionDown && (currentFloor.IsAbove(fromFloor) || currentFloor.IsEqual(fromFloor)))) &&
			!isElevatorOverloaded(e) {
			return e, nil
		}
		// If single elevator doesn't meet criteria, continue to opposite direction check
	}

	// case when more than one elevator with the same direction
	// should check the smallest number between current floor and requested floor, avoiding overloaded elevators
	if len(filteredElevators) > 1 {
		var first = true
		var smallest int
		var nearestE *elevator.Elevator

		for _, e := range filteredElevators {
			// Skip overloaded elevators
			if isElevatorOverloaded(e) {
				continue
			}

			currentFloor := e.CurrentFloor()

			if requestedDirection == domain.DirectionUp && (currentFloor.IsBelow(fromFloor) || currentFloor.IsEqual(fromFloor)) {
				diff := fromFloor.Distance(currentFloor)
				if first || (smallest > diff) {
					smallest = diff
					nearestE = e
					first = false
				}
			}

			if requestedDirection == domain.DirectionDown && (currentFloor.IsAbove(fromFloor) || currentFloor.IsEqual(fromFloor)) {
				diff := currentFloor.Distance(fromFloor)
				if first || (smallest > diff) {
					smallest = diff
					nearestE = e
					first = false
				}
			}
		}

		if nearestE != nil {
			return nearestE, nil
		}
		// If no suitable elevator in same direction, continue to opposite direction check
	}

	/******** OPPOSITE DIRECTION ********/

	filteredElevators = elevatorsOppositeDirections(elevatorsByDirection, requestedDirection)

	if len(filteredElevators) == 1 {
		e := filteredElevators[0]
		// Only accept opposite direction elevator if not overloaded
		if !isElevatorOverloaded(e) {
			return e, nil
		}
		// If single opposite direction elevator is overloaded, continue to multi-elevator check
	}

	if len(filteredElevators) > 1 {
		e := elevatorWithMinRequestsByDirection(filteredElevators, requestedDirection)
		if e != nil {
			return e, nil
		}
		// If all elevators in opposite direction are overloaded, fall through to error
	}

	return nil, domain.NewValidationError("no elevators available for this request", nil).
		WithContext("direction", string(requestedDirection)).
		WithContext("fromFloor", fromFloor.Value()).
		WithContext("toFloor", toFloor.Value())
}

// isElevatorOverloaded checks if an elevator has too many requests to serve efficiently
// Uses the elevator's configured overload threshold (defaults to 12 if not specified)
func isElevatorOverloaded(e *elevator.Elevator) bool {
	directions := e.Directions()
	totalRequests := directions.DirectionsLength()
	return totalRequests > e.OverloadThreshold()
}

func elevatorsMatchingDirections(elevatorsByDirection map[*elevator.Elevator]domain.Direction, requestedDirection domain.Direction) []*elevator.Elevator {
	filteredElevators := make([]*elevator.Elevator, 0)
	for e, d := range elevatorsByDirection {
		if d == requestedDirection {
			filteredElevators = append(filteredElevators, e)
		}
	}
	return filteredElevators
}

func elevatorsOppositeDirections(elevatorsByDirection map[*elevator.Elevator]domain.Direction, requestedDirection domain.Direction) []*elevator.Elevator {
	filteredElevators := make([]*elevator.Elevator, 0)
	for e, d := range elevatorsByDirection {
		if d != requestedDirection {
			filteredElevators = append(filteredElevators, e)
		}
	}
	return filteredElevators
}

func floorsDiff(floor, requestedFloor domain.Floor) int {
	return floor.Distance(requestedFloor)
}

func findNearestElevator(elevatorsWaiting map[*elevator.Elevator]domain.Floor, requestedFloor domain.Floor) *elevator.Elevator {
	var first = true
	var smallest int
	var nearestE *elevator.Elevator

	for e, floor := range elevatorsWaiting {
		// Skip overloaded elevators
		if isElevatorOverloaded(e) {
			continue
		}

		diff := floorsDiff(floor, requestedFloor)
		if first || (smallest > diff) {
			smallest = diff
			nearestE = e
			first = false
		}
	}

	return nearestE
}

func elevatorWithMinRequestsByDirection(elevators []*elevator.Elevator, direction domain.Direction) *elevator.Elevator {
	var el *elevator.Elevator
	var smallest int
	var first = true

	for _, e := range elevators {
		// Skip overloaded elevators
		if isElevatorOverloaded(e) {
			continue
		}

		directions := e.Directions()
		l := 0
		switch direction {
		case domain.DirectionUp:
			l = directions.UpDirectionLength()
		case domain.DirectionDown:
			l = directions.DownDirectionLength()
		default:
			l = directions.UpDirectionLength() + directions.DownDirectionLength()
		}

		if first || (smallest > l) {
			smallest = l
			el = e
			first = false
		}
	}

	return el
}

func (m *Manager) GetStatus() (map[string]interface{}, error) {
	// Use a timeout for status collection to prevent hanging
	ctx, cancel := context.WithTimeout(m.ctx, m.cfg.HealthCheckTimeout)
	defer cancel()

	type statusResult struct {
		status map[string]interface{}
		err    error
	}

	resultCh := make(chan statusResult, 1)

	go func() {
		m.mu.RLock()
		status := make(map[string]interface{}, len(m.elevators))
		for _, e := range m.elevators {
			status[e.Name()] = e.GetStatus()
		}
		m.mu.RUnlock()
		resultCh <- statusResult{status: status, err: nil}
	}()

	select {
	case <-ctx.Done():
		return nil, domain.NewInternalError("status collection timed out", ctx.Err())
	case result := <-resultCh:
		return result.status, result.err
	}
}

// GetHealthStatus returns health status for all elevators including circuit breaker metrics
func (m *Manager) GetHealthStatus() (map[string]interface{}, error) {
	// Use a timeout for health status collection
	ctx, cancel := context.WithTimeout(m.ctx, m.cfg.HealthCheckTimeout)
	defer cancel()

	type healthResult struct {
		health map[string]interface{}
		err    error
	}

	resultCh := make(chan healthResult, 1)

	go func() {
		m.mu.RLock()
		health := make(map[string]interface{})

		elevatorHealth := make(map[string]interface{}, len(m.elevators))
		totalElevators := len(m.elevators)
		healthyElevators := 0
		activeRequests := 0

		for _, e := range m.elevators {
			metrics := e.GetHealthMetrics()
			elevatorHealth[e.Name()] = metrics

			if isHealthy, ok := metrics["is_healthy"].(bool); ok && isHealthy {
				healthyElevators++
			}

			// Count active requests (pending requests) for OpenAPI compliance
			if pendingRequests, ok := metrics["pending_requests"].(int); ok {
				activeRequests += pendingRequests
			}
		}

		health["elevators"] = elevatorHealth
		health["total_elevators"] = totalElevators
		health["elevators_count"] = totalElevators // OpenAPI spec compatibility
		health["healthy_elevators"] = healthyElevators
		health["active_requests"] = activeRequests // OpenAPI spec field
		// System is healthy if there are no elevators (initial state) or if at least one elevator is working
		health["system_healthy"] = totalElevators == 0 || healthyElevators > 0
		health["timestamp"] = time.Now().Format(time.RFC3339) // OpenAPI spec expects date-time format

		m.mu.RUnlock()
		resultCh <- healthResult{health: health, err: nil}
	}()

	select {
	case <-ctx.Done():
		return nil, domain.NewInternalError("health status collection timed out", ctx.Err())
	case result := <-resultCh:
		return result.health, result.err
	}
}

// GetMetrics returns operational metrics for monitoring
func (m *Manager) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalRequests := 0
	totalUpRequests := 0
	totalDownRequests := 0
	healthyElevators := 0
	totalSuccessfulRequests := 0
	totalFailedRequests := 0

	for _, e := range m.elevators {
		directions := e.Directions()
		upRequests := directions.UpDirectionLength()
		downRequests := directions.DownDirectionLength()

		totalUpRequests += upRequests
		totalDownRequests += downRequests

		// Update Prometheus metrics for each elevator
		metrics.SetCurrentFloor(e.Name(), float64(e.CurrentFloor().Value()))
		metrics.SetPendingRequests(e.Name(), "up", float64(upRequests))
		metrics.SetPendingRequests(e.Name(), "down", float64(downRequests))

		// Check elevator health
		healthMetrics := e.GetHealthMetrics()
		if isHealthy, ok := healthMetrics["is_healthy"].(bool); ok && isHealthy {
			healthyElevators++
		}

		// Update circuit breaker metrics
		if cbState, ok := healthMetrics["circuit_breaker_state"].(string); ok {
			stateValue := 0.0 // closed
			switch cbState {
			case "half-open":
				stateValue = 1.0
			case "open":
				stateValue = 2.0
			default:
				stateValue = 0.0
			}
			metrics.SetCircuitBreakerState(e.Name(), stateValue)
		}
	}
	totalRequests = totalUpRequests + totalDownRequests

	// Calculate efficiency metrics
	if totalRequests > 0 {
		efficiency := float64(totalSuccessfulRequests) / float64(totalRequests+totalSuccessfulRequests+totalFailedRequests)
		for _, e := range m.elevators {
			metrics.SetElevatorEfficiency(e.Name(), efficiency)
		}
	}

	// Update system health metrics
	// System is healthy if there are no elevators (initial state) or if at least one elevator is working
	systemHealthy := len(m.elevators) == 0 || healthyElevators > 0
	metrics.SetSystemHealth("elevators", systemHealthy)
	metrics.SetSystemHealth("manager", true) // Manager is healthy if it's responding

	avgLoad := float64(totalRequests) / float64(max(len(m.elevators), 1))

	return map[string]interface{}{
		"total_elevators":     len(m.elevators),
		"healthy_elevators":   healthyElevators,
		"total_requests":      totalRequests,
		"total_up_requests":   totalUpRequests,
		"total_down_requests": totalDownRequests,
		"average_load":        avgLoad,
		"system_efficiency":   float64(totalSuccessfulRequests) / float64(max(totalRequests+totalSuccessfulRequests+totalFailedRequests, 1)),
		"performance_score":   m.calculatePerformanceScore(avgLoad, float64(healthyElevators)/float64(max(len(m.elevators), 1))),
		"timestamp":           time.Now().Format(time.RFC3339), // OpenAPI spec expects date-time format
	}
}

// calculatePerformanceScore calculates a performance score based on load and health
func (m *Manager) calculatePerformanceScore(avgLoad, healthRatio float64) float64 {
	// Performance score considers both load efficiency and system health
	// Score ranges from 0.0 (poor) to 1.0 (excellent)

	// Load efficiency: lower load is better (up to a reasonable point)
	loadScore := 1.0
	if avgLoad > 2.0 {
		loadScore = 2.0 / avgLoad // Penalize high load
	} else if avgLoad < 0.5 {
		loadScore = avgLoad / 0.5 // Reward moderate utilization
	}

	// Health score is directly proportional to healthy elevators ratio
	healthScore := healthRatio

	// Combined score: 60% health, 40% load efficiency
	return (healthScore * 0.6) + (loadScore * 0.4)
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Shutdown gracefully shuts down all elevators
func (m *Manager) Shutdown() {
	m.logger.Info("shutting down elevator manager")

	// Shutdown all elevators
	elevators := m.GetElevators()
	for _, e := range elevators {
		e.Shutdown()
	}

	// Cancel manager context
	if m.cancel != nil {
		m.cancel()
	}

	m.logger.Info("elevator manager shutdown completed")
}
