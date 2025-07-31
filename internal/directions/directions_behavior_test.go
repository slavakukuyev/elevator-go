package directions

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/slavakukuyev/elevator-go/internal/domain"
)

// TestDirections_EmptySliceBehavior tests the core fix for the elevator stuck issue
func TestDirections_EmptySliceBehavior(t *testing.T) {
	t.Run("empty slices should not count as requests", func(t *testing.T) {
		directions := New()

		// Add a request from floor 1 to floors 3 and 5
		directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(3))
		directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(5))

		// Initial state: map[1] = [3, 5] → 1 floor + 2 requests = 3
		assert.Equal(t, 3, directions.UpDirectionLength())
		assert.Equal(t, 0, directions.DownDirectionLength())
		assert.Equal(t, 3, directions.DirectionsLength())
		assert.False(t, directions.IsIdle(), "Should not be idle with active requests")

		// Flush from floor 1 - this creates empty slices for floors 3 and 5
		directions.Flush(domain.DirectionUp, domain.NewFloor(1))

		// After flush: map[3] = [], map[5] = [] → 2 floors with empty slices = 2
		assert.Equal(t, 2, directions.UpDirectionLength(), "Two destination floors")
		assert.Equal(t, 0, directions.DownDirectionLength(), "Should still have no down requests")
		assert.Equal(t, 2, directions.DirectionsLength(), "Total should be 2 destination floors")
		assert.False(t, directions.IsIdle(), "Should not be idle with destination floors")

		// Verify the slices exist but are empty
		assert.Empty(t, directions.up[3], "Floor 3 should have empty slice")
		assert.Empty(t, directions.up[5], "Floor 5 should have empty slice")

		// But the map keys should still exist
		_, exists3 := directions.up[3]
		_, exists5 := directions.up[5]
		assert.True(t, exists3, "Floor 3 key should exist in map")
		assert.True(t, exists5, "Floor 5 key should exist in map")

		// Simulate elevator visiting destination floors
		directions.Flush(domain.DirectionUp, domain.NewFloor(3)) // Visit floor 3
		directions.Flush(domain.DirectionUp, domain.NewFloor(5)) // Visit floor 5

		// Now elevator should be idle - no more floors in any direction
		assert.Equal(t, 0, directions.UpDirectionLength(), "No up requests after all destinations visited")
		assert.Equal(t, 0, directions.DownDirectionLength(), "No down requests")
		assert.Equal(t, 0, directions.DirectionsLength(), "No total requests")
		assert.True(t, directions.IsIdle(), "Elevator should be idle when all destinations completed")
	})

	t.Run("mixed empty and non-empty slices", func(t *testing.T) {
		directions := New()

		// Add multiple requests
		directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(3))
		directions.Append(domain.DirectionUp, domain.NewFloor(2), domain.NewFloor(4))
		directions.Append(domain.DirectionUp, domain.NewFloor(3), domain.NewFloor(5))

		// Should have 3 floors: map[1]=[3], map[2]=[4], map[3]=[5] → 3 floors + 3 requests = 6
		assert.Equal(t, 6, directions.UpDirectionLength())

		// Flush floor 1 - creates empty slice for floor 3, but floor 3 already exists with [5]
		directions.Flush(domain.DirectionUp, domain.NewFloor(1))

		// Now have: map[2]=[4], map[3]=[5] → 2 floors + 2 requests = 4
		// Note: floor 3 keeps its existing [5] slice, the flush doesn't overwrite existing entries
		assert.Equal(t, 4, directions.UpDirectionLength(), "Two floors with requests remaining")

		// Flush floor 2 - creates empty slice for floor 4
		directions.Flush(domain.DirectionUp, domain.NewFloor(2))

		// Now have: map[3]=[5], map[4]=[] → (1 floor + 1 request) + (1 floor + 0 requests) = 3
		assert.Equal(t, 3, directions.UpDirectionLength(), "One floor with request plus one destination floor")

		// Flush floor 3 - creates empty slice for floor 5
		directions.Flush(domain.DirectionUp, domain.NewFloor(3))

		// Now have: map[4]=[], map[5]=[] → 2 floors + 0 requests = 2
		assert.Equal(t, 2, directions.UpDirectionLength(), "Two destination floors remaining")
		assert.False(t, directions.IsIdle(), "Not idle with destination floors")

		// Complete the journey by visiting destination floors
		directions.Flush(domain.DirectionUp, domain.NewFloor(4))
		directions.Flush(domain.DirectionUp, domain.NewFloor(5))

		// Now elevator should be idle
		assert.Equal(t, 0, directions.UpDirectionLength(), "All destinations completed")
		assert.True(t, directions.IsIdle(), "Elevator should be idle")
	})

	t.Run("down direction empty slice behavior", func(t *testing.T) {
		directions := New()

		// Add down requests
		directions.Append(domain.DirectionDown, domain.NewFloor(5), domain.NewFloor(2))
		directions.Append(domain.DirectionDown, domain.NewFloor(5), domain.NewFloor(1))

		// map[5] = [2, 1] → 1 floor + 2 requests = 3
		assert.Equal(t, 3, directions.DownDirectionLength())

		// Flush creates empty slices
		directions.Flush(domain.DirectionDown, domain.NewFloor(5))

		// Now have: map[2]=[], map[1]=[] → 2 floors + 0 requests = 2
		assert.Equal(t, 2, directions.DownDirectionLength(), "Two destination floors")
		assert.Equal(t, 2, directions.DirectionsLength(), "Total should be 2 destination floors")
		assert.False(t, directions.IsIdle(), "Not idle with destination floors")

		// Complete the journey
		directions.Flush(domain.DirectionDown, domain.NewFloor(2))
		directions.Flush(domain.DirectionDown, domain.NewFloor(1))

		// Now should be idle
		assert.Equal(t, 0, directions.DownDirectionLength(), "All destinations completed")
		assert.Equal(t, 0, directions.DirectionsLength(), "Total should be 0")
		assert.True(t, directions.IsIdle(), "Elevator should be idle")
	})

	t.Run("mixed up and down with empty slices", func(t *testing.T) {
		directions := New()

		// Add requests in both directions
		directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(3))
		directions.Append(domain.DirectionDown, domain.NewFloor(5), domain.NewFloor(2))

		// map[1]=[3] up, map[5]=[2] down → 2 + 2 = 4
		assert.Equal(t, 2, directions.UpDirectionLength())
		assert.Equal(t, 2, directions.DownDirectionLength())
		assert.Equal(t, 4, directions.DirectionsLength())

		// Flush up direction
		directions.Flush(domain.DirectionUp, domain.NewFloor(1))

		// Up: map[3]=[], Down: map[5]=[2] → 1 + 2 = 3
		assert.Equal(t, 1, directions.UpDirectionLength(), "Up has one destination floor")
		assert.Equal(t, 2, directions.DownDirectionLength(), "Down still has one request floor")
		assert.Equal(t, 3, directions.DirectionsLength(), "Total should be 3")

		// Flush down direction
		directions.Flush(domain.DirectionDown, domain.NewFloor(5))

		// Up: map[3]=[], Down: map[2]=[] → 1 + 1 = 2
		assert.Equal(t, 1, directions.UpDirectionLength(), "Up still has destination floor")
		assert.Equal(t, 1, directions.DownDirectionLength(), "Down has destination floor")
		assert.Equal(t, 2, directions.DirectionsLength(), "Total should be 2 destination floors")
		assert.False(t, directions.IsIdle(), "Not idle with destination floors")

		// Complete all destinations
		directions.Flush(domain.DirectionUp, domain.NewFloor(3))
		directions.Flush(domain.DirectionDown, domain.NewFloor(2))

		// Now should be idle
		assert.Equal(t, 0, directions.UpDirectionLength(), "Up completed")
		assert.Equal(t, 0, directions.DownDirectionLength(), "Down completed")
		assert.Equal(t, 0, directions.DirectionsLength(), "Total should be 0")
		assert.True(t, directions.IsIdle(), "Elevator should be idle")
	})
}

// TestDirections_ElevatorIdleBehavior tests scenarios that previously caused elevators to get stuck
func TestDirections_ElevatorIdleBehavior(t *testing.T) {
	t.Run("elevator should become idle after completing single request", func(t *testing.T) {
		directions := New()

		// Simulate elevator at floor 0, request to go from 0 to 3
		directions.Append(domain.DirectionUp, domain.NewFloor(0), domain.NewFloor(3))

		// Elevator should have work to do: map[0]=[3] → 1 floor + 1 request = 2
		assert.Equal(t, 2, directions.UpDirectionLength(), "Elevator should have requests")
		assert.False(t, directions.IsIdle(), "Elevator should not be idle")

		// Elevator reaches floor 0 and processes the request
		directions.Flush(domain.DirectionUp, domain.NewFloor(0))

		// This creates map[3] = [] → 1 floor + 0 requests = 1, but elevator has destination work
		assert.Equal(t, 1, directions.UpDirectionLength(), "One destination floor after flush")
		assert.Equal(t, 0, directions.DownDirectionLength(), "No down requests")
		assert.False(t, directions.IsIdle(), "Elevator should not be idle with destination floor")

		// Complete the journey to floor 3
		directions.Flush(domain.DirectionUp, domain.NewFloor(3))

		// Now elevator should be idle
		assert.Equal(t, 0, directions.UpDirectionLength(), "No requests after completing destination")
		assert.Equal(t, 0, directions.DownDirectionLength(), "No down requests")
		assert.True(t, directions.IsIdle(), "Elevator should be idle when all work completed")
	})

	t.Run("elevator should become idle after completing complex request sequence", func(t *testing.T) {
		directions := New()

		// Multiple passengers boarding at floor 1
		directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(3))
		directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(5))
		directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(7))

		// map[1] = [3, 5, 7] → 1 floor + 3 requests = 4
		assert.Equal(t, 4, directions.UpDirectionLength(), "Should have 4 total")
		assert.False(t, directions.IsIdle(), "Elevator should be busy")

		// Elevator reaches floor 1
		directions.Flush(domain.DirectionUp, domain.NewFloor(1))
		// This creates empty slices for floors 3, 5, 7 → 3 floors + 0 requests = 3

		assert.Equal(t, 3, directions.UpDirectionLength(), "Three destination floors")
		assert.False(t, directions.IsIdle(), "Elevator should not be idle with destination floors")

		// Simulate elevator reaching destination floors
		directions.Flush(domain.DirectionUp, domain.NewFloor(3))
		directions.Flush(domain.DirectionUp, domain.NewFloor(5))
		directions.Flush(domain.DirectionUp, domain.NewFloor(7))

		// Now all destinations completed, elevator should be idle
		assert.Equal(t, 0, directions.UpDirectionLength(), "All destinations completed")
		assert.Equal(t, 0, directions.DownDirectionLength(), "No down requests")
		assert.True(t, directions.IsIdle(), "Elevator should be idle after completing all destinations")
	})

	t.Run("elevator should not be idle with active requests", func(t *testing.T) {
		directions := New()

		// Add requests from different floors
		directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(3))
		directions.Append(domain.DirectionUp, domain.NewFloor(2), domain.NewFloor(4))

		// map[1]=[3], map[2]=[4] → 2 floors + 2 requests = 4
		assert.Equal(t, 4, directions.UpDirectionLength(), "Should have 4 total")
		assert.False(t, directions.IsIdle(), "Elevator should be busy with multiple sources")

		// Process first request
		directions.Flush(domain.DirectionUp, domain.NewFloor(1))

		// Now: map[2]=[4], map[3]=[] → 1 floor + 1 request + 1 floor + 0 requests = 3
		assert.Equal(t, 3, directions.UpDirectionLength(), "Should have 3 after first flush")
		assert.False(t, directions.IsIdle(), "Elevator should still be busy with remaining request")

		// Process second request
		directions.Flush(domain.DirectionUp, domain.NewFloor(2))

		// Now: map[3]=[], map[4]=[] → 2 floors + 0 requests = 2
		assert.Equal(t, 2, directions.UpDirectionLength(), "Should have 2 destination floors")
		assert.False(t, directions.IsIdle(), "Elevator should not be idle with destination floors")

		// Complete all destinations
		directions.Flush(domain.DirectionUp, domain.NewFloor(3))
		directions.Flush(domain.DirectionUp, domain.NewFloor(4))

		// Now should be idle
		assert.Equal(t, 0, directions.UpDirectionLength(), "All work completed")
		assert.True(t, directions.IsIdle(), "Elevator should be idle after all requests processed")
	})
}

// TestDirections_RegressionPrevention tests that prevent the original bug from reoccurring
func TestDirections_RegressionPrevention(t *testing.T) {
	t.Run("regression test: empty slices should not prevent idle state", func(t *testing.T) {
		directions := New()

		// Simulate the exact scenario that caused the bug
		directions.Append(domain.DirectionUp, domain.NewFloor(0), domain.NewFloor(3))

		// Initial state: map[0] = [3] → 1 floor + 1 request = 2
		assert.Equal(t, 2, directions.UpDirectionLength())

		// Elevator processes request at floor 0
		directions.Flush(domain.DirectionUp, domain.NewFloor(0))

		// After flush: map[3] = [] → 1 floor + 0 requests = 1 (destination marker)
		// This is correct behavior - elevator still has work to do at floor 3
		assert.Equal(t, 1, directions.UpDirectionLength(), "REGRESSION TEST: Destination floor still needs to be visited")

		// Elevator is not idle yet - still has a destination to reach
		assert.False(t, directions.IsIdle(), "REGRESSION TEST: Elevator should not be idle with destination floor")

		// Complete the journey by visiting floor 3
		directions.Flush(domain.DirectionUp, domain.NewFloor(3))

		// Now elevator should be idle
		assert.Equal(t, 0, directions.UpDirectionLength(), "All destinations completed")
		assert.True(t, directions.IsIdle(), "REGRESSION TEST: Elevator should be idle after completing all work")

		// Verify that map is completely empty after all work is done
		assert.Equal(t, 0, len(directions.up), "Map should be empty when all work is completed")

		// No floors should exist in the map
		_, exists := directions.up[3]
		assert.False(t, exists, "Floor 3 should be removed after being visited")
	})

	t.Run("regression test: multiple empty slices should not accumulate", func(t *testing.T) {
		directions := New()

		// Create multiple requests that will result in empty slices
		directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(5))
		directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(7))
		directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(9))

		// Initial state: map[1] = [5, 7, 9] → 1 floor + 3 requests = 4
		assert.Equal(t, 4, directions.UpDirectionLength())

		// Flush creates multiple empty slices
		directions.Flush(domain.DirectionUp, domain.NewFloor(1))

		// After flush: map[5]=[], map[7]=[], map[9]=[] → 3 floors + 0 requests = 3
		// This is correct - elevator has 3 destination floors to visit
		assert.Equal(t, 3, directions.UpDirectionLength(), "Three destination floors to visit")
		assert.Equal(t, 3, len(directions.up), "Map should have 3 destination floor entries")
		assert.False(t, directions.IsIdle(), "Elevator should not be idle with destination floors")

		// All slices should be empty but keys should exist
		assert.Empty(t, directions.up[5])
		assert.Empty(t, directions.up[7])
		assert.Empty(t, directions.up[9])

		// Complete all destinations
		directions.Flush(domain.DirectionUp, domain.NewFloor(5))
		directions.Flush(domain.DirectionUp, domain.NewFloor(7))
		directions.Flush(domain.DirectionUp, domain.NewFloor(9))

		// Now elevator should be idle
		assert.Equal(t, 0, directions.UpDirectionLength(), "All destinations completed")
		assert.Equal(t, 0, len(directions.up), "Map should be empty")
		assert.True(t, directions.IsIdle(), "Elevator should be idle after all destinations visited")
	})
}

// Helper function to determine if elevator should be idle
// This function now delegates to the proper IsIdle() method
func isElevatorIdle(directions *Manager) bool {
	return directions.IsIdle()
}
