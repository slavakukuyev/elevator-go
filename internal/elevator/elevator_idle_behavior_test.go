package elevator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/slavakukuyev/elevator-go/internal/domain"
)

// TestElevator_IdleStateTransitions tests that elevators properly transition to idle state
func TestElevator_IdleStateTransitions(t *testing.T) {
	t.Run("elevator becomes idle after single request completion", func(t *testing.T) {
		elevator, err := New("IdleTestElevator", 0, 10, time.Millisecond*10, time.Millisecond*10,
			30*time.Second, 5, 30*time.Second, 3, 12)
		require.NoError(t, err)
		defer elevator.Shutdown()

		// Initial state should be idle
		assert.Equal(t, domain.DirectionIdle, elevator.CurrentDirection())
		assert.Equal(t, 0, elevator.Directions().DirectionsLength())

		// Make a request
		elevator.Request(domain.DirectionUp, domain.NewFloor(0), domain.NewFloor(3))

		// Elevator should now be moving: map[0] = [3] → 1 floor + 1 request = 2
		assert.Equal(t, domain.DirectionUp, elevator.CurrentDirection())
		assert.Equal(t, 2, elevator.Directions().DirectionsLength())

		// Simulate elevator processing the request (picking up passenger at floor 0)
		elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(0))

		// After flushing: map[3] = [] → 1 floor + 0 requests = 1 (destination marker)
		assert.Equal(t, 1, elevator.Directions().DirectionsLength(), "One destination floor remaining")
		assert.False(t, elevator.Directions().IsIdle(), "Not idle with destination floor")

		// Elevator should still be moving up to reach the destination
		assert.Equal(t, domain.DirectionUp, elevator.CurrentDirection())

		// Simulate reaching destination floor
		elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(3))

		// Now elevator should have no pending requests - all work completed
		assert.Equal(t, 0, elevator.Directions().DirectionsLength())
		assert.True(t, elevator.Directions().IsIdle(), "Elevator should be idle after completing all work")
	})

	t.Run("elevator status reflects idle state correctly", func(t *testing.T) {
		elevator, err := New("StatusTestElevator", 0, 5, time.Millisecond*10, time.Millisecond*10,
			30*time.Second, 5, 30*time.Second, 3, 12)
		require.NoError(t, err)
		defer elevator.Shutdown()

		// Check initial status
		status := elevator.GetStatus()
		assert.Equal(t, "StatusTestElevator", status.Name)
		assert.Equal(t, domain.DirectionIdle, status.Direction)
		assert.Equal(t, 0, status.Requests)
		assert.True(t, status.IsIdle())
		assert.False(t, status.IsMoving())

		// Make a request
		elevator.Request(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(4))

		// Status should show busy: map[1] = [4] → 1 floor + 1 request = 2
		status = elevator.GetStatus()
		assert.Equal(t, domain.DirectionUp, status.Direction)
		assert.Equal(t, 2, status.Requests)
		assert.False(t, status.IsIdle())
		assert.True(t, status.IsMoving())

		// Process the pickup
		elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(1))

		// Status should show 1 request: map[4] = [] → 1 floor + 0 requests = 1 (destination marker)
		status = elevator.GetStatus()
		assert.Equal(t, 1, status.Requests, "Status should reflect 1 destination floor after flush")

		// Process the dropoff
		elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(4))

		// Still should show 0 requests
		status = elevator.GetStatus()
		assert.Equal(t, 0, status.Requests)
	})

	t.Run("health metrics reflect idle state correctly", func(t *testing.T) {
		elevator, err := New("HealthTestElevator", 0, 8, time.Millisecond*10, time.Millisecond*10,
			30*time.Second, 5, 30*time.Second, 3, 12)
		require.NoError(t, err)
		defer elevator.Shutdown()

		// Check initial health metrics
		health := elevator.GetHealthMetrics()
		assert.Equal(t, 0, health["pending_requests"])
		assert.Equal(t, string(domain.DirectionIdle), health["direction"])

		// Make a down request (elevator at floor 0, needs to go up to floor 5 first)
		elevator.Request(domain.DirectionDown, domain.NewFloor(5), domain.NewFloor(2))

		// Health should show busy state: map[5] = [2] → 1 floor + 1 request = 2
		// Note: elevator starts at floor 0, so it needs to go UP to reach floor 5 first
		health = elevator.GetHealthMetrics()
		assert.Equal(t, 2, health["pending_requests"])
		assert.Equal(t, string(domain.DirectionUp), health["direction"], "Elevator should go UP first to reach pickup floor")

		// Process the request (elevator reaches floor 5 to pick up passenger)
		elevator.Directions().Flush(domain.DirectionDown, domain.NewFloor(5))

		// Health should show 1 pending request: map[2] = [] → 1 floor + 0 requests = 1 (destination marker)
		health = elevator.GetHealthMetrics()
		assert.Equal(t, 1, health["pending_requests"], "Health metrics should show 1 destination floor")
	})
}

// TestElevator_ComplexIdleScenarios tests more complex scenarios that could prevent idle state
func TestElevator_ComplexIdleScenarios(t *testing.T) {
	t.Run("multiple passengers same origin floor", func(t *testing.T) {
		elevator, err := New("MultiPassengerElevator", 0, 15, time.Millisecond*10, time.Millisecond*10,
			30*time.Second, 5, 30*time.Second, 3, 12)
		require.NoError(t, err)
		defer elevator.Shutdown()

		// Multiple passengers board at floor 3, going to different floors
		elevator.Request(domain.DirectionUp, domain.NewFloor(3), domain.NewFloor(7))
		elevator.Request(domain.DirectionUp, domain.NewFloor(3), domain.NewFloor(10))
		elevator.Request(domain.DirectionUp, domain.NewFloor(3), domain.NewFloor(12))

		// map[3] = [7, 10, 12] → 1 floor + 3 requests = 4
		assert.Equal(t, 4, elevator.Directions().DirectionsLength(), "Should have 1 source floor with 3 destinations")

		// Elevator reaches floor 3 and picks up all passengers
		elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(3))

		// After flush: map[7]=[], map[10]=[], map[12]=[] → 3 floors + 0 requests = 3
		assert.Equal(t, 3, elevator.Directions().DirectionsLength(), "Should have 3 destination floors after pickup")

		// Elevator continues to destinations and drops off passengers
		elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(7))
		elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(10))
		elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(12))

		// Should remain at 0 requests
		assert.Equal(t, 0, elevator.Directions().DirectionsLength(), "Should remain idle after all dropoffs")
	})

	t.Run("mixed direction requests sequence", func(t *testing.T) {
		elevator, err := New("MixedDirectionElevator", 0, 20, time.Millisecond*10, time.Millisecond*10,
			30*time.Second, 5, 30*time.Second, 3, 12)
		require.NoError(t, err)
		defer elevator.Shutdown()

		// Up request: map[2] = [8] → 1 floor + 1 request = 2
		elevator.Request(domain.DirectionUp, domain.NewFloor(2), domain.NewFloor(8))
		assert.Equal(t, 2, elevator.Directions().DirectionsLength())

		// Down request: up[2]=[8] + down[15]=[5] → 2 + 2 = 4
		elevator.Request(domain.DirectionDown, domain.NewFloor(15), domain.NewFloor(5))
		assert.Equal(t, 4, elevator.Directions().DirectionsLength())

		// Process up request: up[8]=[] + down[15]=[5] → 1 + 2 = 3
		elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(2))
		assert.Equal(t, 3, elevator.Directions().DirectionsLength(), "Should have 3 remaining (1 up destination + down request)")

		// Process down request: up[8]=[] + down[5]=[] → 1 + 1 = 2
		elevator.Directions().Flush(domain.DirectionDown, domain.NewFloor(15))
		assert.Equal(t, 2, elevator.Directions().DirectionsLength(), "Should have 2 destination floors remaining")

		// Complete both destinations
		elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(8))
		elevator.Directions().Flush(domain.DirectionDown, domain.NewFloor(5))
		assert.Equal(t, 0, elevator.Directions().DirectionsLength(), "Should be idle after completing all destinations")
	})

	t.Run("rapid request sequence", func(t *testing.T) {
		elevator, err := New("RapidRequestElevator", 0, 10, time.Millisecond*10, time.Millisecond*10,
			30*time.Second, 5, 30*time.Second, 3, 12)
		require.NoError(t, err)
		defer elevator.Shutdown()

		// Rapid sequence of requests
		floors := []struct{ from, to int }{
			{1, 5}, {2, 6}, {3, 7}, {4, 8}, {5, 9},
		}

		for _, req := range floors {
			elevator.Request(domain.DirectionUp, domain.NewFloor(req.from), domain.NewFloor(req.to))
		}

		// 5 floors + 5 requests = 10 total
		assert.Equal(t, 10, elevator.Directions().DirectionsLength(), "Should have 5 source floors with requests")

		// Process each pickup
		for _, req := range floors {
			elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(req.from))
		}

		// After all pickups: 5 destination floors remaining
		assert.Equal(t, 4, elevator.Directions().DirectionsLength(), "Should have 4 destination floors after all pickups")

		// Process each dropoff
		for _, req := range floors {
			elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(req.to))
		}

		assert.Equal(t, 0, elevator.Directions().DirectionsLength(), "Should be idle after all dropoffs")
	})
}

// TestElevator_RegressionIdleBug tests the specific bug that was fixed
func TestElevator_RegressionIdleBug(t *testing.T) {
	t.Run("regression: elevator A stuck in moving status", func(t *testing.T) {
		// This test reproduces the exact scenario mentioned in the issue
		elevator, err := New("Elevator A", -2, 10, time.Millisecond*50, time.Millisecond*25,
			30*time.Second, 5, 30*time.Second, 3, 12)
		require.NoError(t, err)
		defer elevator.Shutdown()

		// Initial state: elevator should be idle
		assert.Equal(t, domain.DirectionIdle, elevator.CurrentDirection())
		assert.Equal(t, 0, elevator.Directions().DirectionsLength())

		// Make a simple request that previously caused the bug
		elevator.Request(domain.DirectionUp, domain.NewFloor(0), domain.NewFloor(5))

		// Elevator should be moving: map[0] = [5] → 1 floor + 1 request = 2
		assert.Equal(t, domain.DirectionUp, elevator.CurrentDirection())
		assert.Equal(t, 2, elevator.Directions().DirectionsLength())

		// Simulate the elevator reaching the pickup floor
		elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(0))

		// CRITICAL: After pickup, there's still a destination floor to visit
		// After flush: map[5] = [] → 1 floor + 0 requests = 1 (destination marker)
		// This is correct behavior - elevator still has work to do at floor 5
		assert.Equal(t, 1, elevator.Directions().DirectionsLength(),
			"REGRESSION TEST: Elevator should show 1 destination floor after pickup")
		assert.False(t, elevator.Directions().IsIdle(), "Elevator should not be idle with destination floor")

		// Complete the journey by visiting the destination
		elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(5))

		// Now elevator should be idle
		assert.Equal(t, 0, elevator.Directions().DirectionsLength(), "All work completed")
		assert.True(t, elevator.Directions().IsIdle(), "Elevator should be idle after completing all work")

		// The elevator state machine should eventually set direction to idle
		// when it detects no more requests

		// Simulate reaching destination
		elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(5))

		// Should still be 0 requests
		assert.Equal(t, 0, elevator.Directions().DirectionsLength(),
			"Should remain at 0 requests after destination flush")

		// Verify elevator can properly report idle status
		status := elevator.GetStatus()
		assert.Equal(t, 0, status.Requests, "Status should show 0 requests")
	})

	t.Run("regression: empty slices in health metrics", func(t *testing.T) {
		elevator, err := New("HealthMetricsElevator", 0, 5, time.Millisecond*10, time.Millisecond*10,
			30*time.Second, 5, 30*time.Second, 3, 12)
		require.NoError(t, err)
		defer elevator.Shutdown()

		// Make a request
		elevator.Request(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(4))

		// Initial health check: map[1] = [4] → 1 floor + 1 request = 2
		health := elevator.GetHealthMetrics()
		assert.Equal(t, 2, health["pending_requests"])

		// Process pickup (creates empty slice for floor 4)
		elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(1))

		// Health metrics should show 1 pending request: map[4] = [] → 1 floor + 0 requests = 1
		health = elevator.GetHealthMetrics()
		assert.Equal(t, 1, health["pending_requests"],
			"REGRESSION TEST: Health metrics should show 1 destination floor")

		// Process dropoff
		elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(4))

		// Should still show 0
		health = elevator.GetHealthMetrics()
		assert.Equal(t, 0, health["pending_requests"])
	})
}

// TestElevator_EdgeCaseIdleBehavior tests edge cases that could affect idle behavior
func TestElevator_EdgeCaseIdleBehavior(t *testing.T) {
	t.Run("same floor pickup and dropoff", func(t *testing.T) {
		elevator, err := New("SameFloorElevator", 0, 10, time.Millisecond*10, time.Millisecond*10,
			30*time.Second, 5, 30*time.Second, 3, 12)
		require.NoError(t, err)
		defer elevator.Shutdown()

		// Request from floor to same floor (edge case)
		elevator.Request(domain.DirectionUp, domain.NewFloor(3), domain.NewFloor(3))

		// This creates an unusual scenario where pickup and dropoff are the same
		elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(3))

		// Should be idle
		assert.Equal(t, 0, elevator.Directions().DirectionsLength())
	})

	t.Run("boundary floor requests", func(t *testing.T) {
		elevator, err := New("BoundaryElevator", -3, 7, time.Millisecond*10, time.Millisecond*10,
			30*time.Second, 5, 30*time.Second, 3, 12)
		require.NoError(t, err)
		defer elevator.Shutdown()

		// Request at minimum boundary
		elevator.Request(domain.DirectionUp, domain.NewFloor(-3), domain.NewFloor(0))
		elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(-3))
		assert.Equal(t, 1, elevator.Directions().DirectionsLength()) // Destination floor 0 remaining

		// Complete the destination
		elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(0))
		assert.Equal(t, 0, elevator.Directions().DirectionsLength())

		// Request at maximum boundary
		elevator.Request(domain.DirectionDown, domain.NewFloor(7), domain.NewFloor(3))
		elevator.Directions().Flush(domain.DirectionDown, domain.NewFloor(7))
		assert.Equal(t, 1, elevator.Directions().DirectionsLength()) // Destination floor 3 remaining

		// Complete the destination
		elevator.Directions().Flush(domain.DirectionDown, domain.NewFloor(3))
		assert.Equal(t, 0, elevator.Directions().DirectionsLength())
	})

	t.Run("zero floor handling", func(t *testing.T) {
		elevator, err := New("ZeroFloorElevator", -2, 3, time.Millisecond*10, time.Millisecond*10,
			30*time.Second, 5, 30*time.Second, 3, 12)
		require.NoError(t, err)
		defer elevator.Shutdown()

		// Requests involving floor 0: up[0]=[2] + down[0]=[-1] → 2 + 2 = 4
		elevator.Request(domain.DirectionUp, domain.NewFloor(0), domain.NewFloor(2))
		elevator.Request(domain.DirectionDown, domain.NewFloor(0), domain.NewFloor(-1))

		assert.Equal(t, 4, elevator.Directions().DirectionsLength())

		// Process requests
		elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(0))
		elevator.Directions().Flush(domain.DirectionDown, domain.NewFloor(0))

		// After both flushes: up[2]=[], down[-1]=[] → 1 + 1 = 2 destination floors
		assert.Equal(t, 2, elevator.Directions().DirectionsLength(), "Two destination floors remaining")

		// Complete both destinations
		elevator.Directions().Flush(domain.DirectionUp, domain.NewFloor(2))
		elevator.Directions().Flush(domain.DirectionDown, domain.NewFloor(-1))

		// Now should be idle
		assert.Equal(t, 0, elevator.Directions().DirectionsLength(), "Zero floor requests completed - elevator idle")
	})
}
