package directions

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/slavakukuyev/elevator-go/internal/domain"
)

// TestDirections_WebSocketBehaviorFix tests the complete fix for frontend synchronization
func TestDirections_WebSocketBehaviorFix(t *testing.T) {
	t.Run("WebSocket format compatibility - empty slices should report 0", func(t *testing.T) {
		directions := New()

		// Simulate elevator request cycle
		directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(5))

		// Initially: map[1] = [5] → 1 floor + 1 request = 2
		assert.Equal(t, 2, directions.DirectionsLength())

		// Simulate elevator reaching pickup floor 1 (creates empty slice at floor 5)
		directions.Flush(domain.DirectionUp, domain.NewFloor(1))

		// After flush: map[5] = [] → 1 floor + 0 requests = 1 (destination marker)
		// For WebSocket idle status, use IsIdle() instead of DirectionsLength()
		assert.Equal(t, 1, directions.DirectionsLength(), "One destination floor remaining")
		assert.False(t, directions.IsIdle(), "Not idle - elevator must visit floor 5 for WebSocket status")

		// Simulate elevator reaching destination floor 5
		directions.Flush(domain.DirectionUp, domain.NewFloor(5))

		// Now elevator should be idle - all work completed
		assert.Equal(t, 0, directions.DirectionsLength(), "All work completed")
		assert.True(t, directions.IsIdle(), "WebSocket should report idle status")
	})

	t.Run("Frontend synchronization scenario - multiple elevators", func(t *testing.T) {
		// Simulate the exact scenario that was causing frontend confusion
		elevatorADirections := New()
		elevatorCDirections := New()
		parkingElevatorDirections := New()

		// All elevators receive requests
		elevatorADirections.Append(domain.DirectionUp, domain.NewFloor(-2), domain.NewFloor(5))
		elevatorCDirections.Append(domain.DirectionDown, domain.NewFloor(12), domain.NewFloor(3))
		parkingElevatorDirections.Append(domain.DirectionUp, domain.NewFloor(-5), domain.NewFloor(0))

		// All have 2 initially (1 floor + 1 request each)
		assert.Equal(t, 2, elevatorADirections.DirectionsLength())
		assert.Equal(t, 2, elevatorCDirections.DirectionsLength())
		assert.Equal(t, 2, parkingElevatorDirections.DirectionsLength())

		// Elevators complete their requests (pickup phase)
		elevatorADirections.Flush(domain.DirectionUp, domain.NewFloor(-2))
		elevatorCDirections.Flush(domain.DirectionDown, domain.NewFloor(12))
		parkingElevatorDirections.Flush(domain.DirectionUp, domain.NewFloor(-5))

		// After pickup: each has 1 destination floor remaining
		assert.Equal(t, 1, elevatorADirections.DirectionsLength(), "Elevator A has destination floor 5")
		assert.Equal(t, 1, elevatorCDirections.DirectionsLength(), "Elevator C has destination floor 3")
		assert.Equal(t, 1, parkingElevatorDirections.DirectionsLength(), "Parking Elevator has destination floor 0")

		// None should be idle yet - they have destinations to visit
		assert.False(t, elevatorADirections.IsIdle(), "Elevator A not idle")
		assert.False(t, elevatorCDirections.IsIdle(), "Elevator C not idle")
		assert.False(t, parkingElevatorDirections.IsIdle(), "Parking Elevator not idle")

		// Elevators complete destinations (dropoff phase)
		elevatorADirections.Flush(domain.DirectionUp, domain.NewFloor(5))
		elevatorCDirections.Flush(domain.DirectionDown, domain.NewFloor(3))
		parkingElevatorDirections.Flush(domain.DirectionUp, domain.NewFloor(0))

		// Now all elevators should be idle - all work completed
		assert.Equal(t, 0, elevatorADirections.DirectionsLength(), "Elevator A completed all work")
		assert.Equal(t, 0, elevatorCDirections.DirectionsLength(), "Elevator C completed all work")
		assert.Equal(t, 0, parkingElevatorDirections.DirectionsLength(), "Parking Elevator completed all work")

		// WebSocket idle status should be true
		assert.True(t, elevatorADirections.IsIdle(), "Elevator A should be idle for WebSocket")
		assert.True(t, elevatorCDirections.IsIdle(), "Elevator C should be idle for WebSocket")
		assert.True(t, parkingElevatorDirections.IsIdle(), "Parking Elevator should be idle for WebSocket")
	})

	t.Run("Rapid request processing - WebSocket consistency", func(t *testing.T) {
		directions := New()

		// Rapid sequence of requests and flushes (simulating busy elevator)
		floors := []struct{ from, to int }{
			{0, 3}, {1, 4}, {2, 5}, {6, 9}, {7, 10},
		}

		// Add all requests
		for _, req := range floors {
			directions.Append(domain.DirectionUp, domain.NewFloor(req.from), domain.NewFloor(req.to))
		}

		// Should have 5 floors + 5 requests = 10 total
		assert.Equal(t, 10, directions.DirectionsLength(), "Should have 5 source floors with requests")

		// Process all pickups (creates empty destination slices)
		for _, req := range floors {
			directions.Flush(domain.DirectionUp, domain.NewFloor(req.from))
		}

		// After pickups: 5 destination floors remaining
		assert.Equal(t, 5, directions.DirectionsLength(),
			"Should have 5 destination floors after pickups")
		assert.False(t, directions.IsIdle(), "Not idle with destination floors")

		// Process all dropoffs
		for _, req := range floors {
			directions.Flush(domain.DirectionUp, domain.NewFloor(req.to))
		}

		// Now all work is completed
		assert.Equal(t, 0, directions.DirectionsLength(), "All dropoffs completed")
		assert.True(t, directions.IsIdle(), "Elevator should be idle for WebSocket after all work done")
	})

	t.Run("Mixed direction WebSocket behavior", func(t *testing.T) {
		directions := New()

		// Mixed up and down requests
		directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(5))
		directions.Append(domain.DirectionDown, domain.NewFloor(8), domain.NewFloor(3))

		// Initial: up[1]=[5] + down[8]=[3] → 2 + 2 = 4
		assert.Equal(t, 4, directions.DirectionsLength())
		assert.Equal(t, 2, directions.UpDirectionLength())
		assert.Equal(t, 2, directions.DownDirectionLength())

		// Process up direction pickup
		directions.Flush(domain.DirectionUp, domain.NewFloor(1))

		// After up flush: up[5]=[] + down[8]=[3] → 1 + 2 = 3
		assert.Equal(t, 3, directions.DirectionsLength())
		assert.Equal(t, 1, directions.UpDirectionLength())
		assert.Equal(t, 2, directions.DownDirectionLength())

		// Process down direction pickup
		directions.Flush(domain.DirectionDown, domain.NewFloor(8))

		// After both pickups: up[5]=[] + down[3]=[] → 1 + 1 = 2
		assert.Equal(t, 2, directions.DirectionsLength(),
			"Should have 2 destination floors remaining")
		assert.Equal(t, 1, directions.UpDirectionLength())
		assert.Equal(t, 1, directions.DownDirectionLength())
		assert.False(t, directions.IsIdle(), "Not idle with destination floors")

		// Process destinations
		directions.Flush(domain.DirectionUp, domain.NewFloor(5))
		directions.Flush(domain.DirectionDown, domain.NewFloor(3))

		// Now all work is completed
		assert.Equal(t, 0, directions.DirectionsLength())
		assert.Equal(t, 0, directions.UpDirectionLength())
		assert.Equal(t, 0, directions.DownDirectionLength())
		assert.True(t, directions.IsIdle(), "Elevator should be idle after all destinations completed")
	})

	t.Run("Edge case - same floor pickup and destination", func(t *testing.T) {
		directions := New()

		// Edge case that might cause confusion
		directions.Append(domain.DirectionUp, domain.NewFloor(3), domain.NewFloor(3))

		// Initial: map[3] = [3] → 1 floor + 1 request = 2
		assert.Equal(t, 2, directions.DirectionsLength())

		// Flush the floor (both pickup and destination)
		directions.Flush(domain.DirectionUp, domain.NewFloor(3))

		// After flush: since pickup and destination are same floor, it should become empty
		assert.Equal(t, 0, directions.DirectionsLength(),
			"Same floor pickup/destination should result in immediate completion")
		assert.True(t, directions.IsIdle(), "Should be idle immediately for same floor requests")
	})

	t.Run("Regression test - exact stuck elevator scenario", func(t *testing.T) {
		// Test the exact scenario that was reported as "stuck"
		directions := New()

		// Elevator A scenario: from floor -2 to floor 10 range, request from 0 to 5
		directions.Append(domain.DirectionUp, domain.NewFloor(0), domain.NewFloor(5))

		// Elevator starts processing: map[0] = [5] → 1 floor + 1 request = 2
		assert.Equal(t, 2, directions.DirectionsLength(), "Initial request should count")

		// Elevator reaches pickup floor 0
		directions.Flush(domain.DirectionUp, domain.NewFloor(0))

		// After pickup: map[5] = [] → 1 floor + 0 requests = 1 (destination marker)
		// This is correct behavior - elevator still has work to do at floor 5
		assert.Equal(t, 1, directions.DirectionsLength(),
			"REGRESSION FIX: Destination floor still needs to be visited")
		assert.False(t, directions.IsIdle(), "Elevator should not be idle with destination floor")

		// Elevator reaches destination floor 5
		directions.Flush(domain.DirectionUp, domain.NewFloor(5))

		// Now all work is completed - elevator should be idle
		assert.Equal(t, 0, directions.DirectionsLength(),
			"Should be idle after completing journey")

		// Verify individual direction counts are also 0
		assert.Equal(t, 0, directions.UpDirectionLength(), "Up direction should be 0")
		assert.Equal(t, 0, directions.DownDirectionLength(), "Down direction should be 0")
		assert.True(t, directions.IsIdle(), "Elevator should be idle for WebSocket after all work completed")
	})
}
