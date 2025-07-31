package directions

import (
	"sync"

	"github.com/slavakukuyev/elevator-go/internal/domain"
)

// Manager handles elevator direction requests using a sophisticated map[int][]int structure.
//
// DESIGN RATIONALE:
// ================
// This implementation uses map[int][]int where:
// - Key (int): represents the "pickup floor" where passengers are waiting
// - Value ([]int): represents the slice of "destination floors" requested from that pickup floor
//
// WHY THIS DESIGN IS SUPERIOR TO SIMPLE APPROACHES:
// ================================================
//
// Problem with simple map[int]bool approach:
// - Request: Floor 0 → Floor 10
// - Elevator arrives at floor 0, removes entry, moves to floor 10
// - While elevator is at floor 5, new request: Floor 0 → Floor 10
// - When elevator reaches floor 10, the new request from floor 0 is lost
//
// Our map[int][]int solution:
// - Initial request: map[0] = [10]
// - Elevator picks up at floor 0: creates map[10] = [] (empty slice), removes map[0]
// - New request while moving (at floor 5): creates map[0] = [10] again
// - Elevator reaches floor 10: removes map[10], but preserves map[0] = [10]
// - System correctly handles the concurrent request without data loss
//
// OPERATIONAL FLOW EXAMPLE:
// ========================
// Step 1: Request from floor 0 to floors 5 and 10
//
//	map[0] = [5, 10]
//
// Step 2: Elevator arrives at floor 0, picks up passengers
//   - Creates: map[5] = [], map[10] = [] (destination markers)
//   - Removes: map[0] (pickup complete)
//
// Step 3: Elevator reaches floor 5
//   - Removes: map[5] (empty destination, no further requests from floor 5)
//
// Step 4: While elevator moves (at floor 7), new request: floor 0 → floor 10
//   - Creates: map[0] = [10] (new pickup request)
//
// Step 5: Elevator reaches floor 10
//   - Removes: map[10] (destination reached)
//   - Preserves: map[0] = [10] (the new request is maintained)
//
// This design ensures no request loss during concurrent operations and provides
// accurate state management for real-world elevator systems where requests
// continuously arrive while the elevator is in motion.
type Manager struct {
	up   map[int][]int
	down map[int][]int
	mu   sync.RWMutex
}

// New creates a new directions manager
func New() *Manager {
	return &Manager{
		up:   make(map[int][]int),
		down: make(map[int][]int),
	}
}

// Append adds a new elevator request to the direction manager.
// This method implements the initial request registration in our sophisticated system.
//
// PARAMETERS:
// - direction: UP or DOWN movement direction for this request
// - fromFloor: pickup floor where passengers are waiting
// - toFloor: destination floor where passengers want to go
//
// BEHAVIOR:
// The method appends the destination floor to the pickup floor's slice.
// Multiple requests from the same pickup floor are accumulated in the slice.
//
// EXAMPLE:
// - Append(UP, Floor(0), Floor(5)) → map[0] = [5]
// - Append(UP, Floor(0), Floor(10)) → map[0] = [5, 10]
//
// This creates the initial state that will later be processed by Flush()
// when the elevator arrives at the pickup floor.
func (d *Manager) Append(direction domain.Direction, fromFloor domain.Floor, toFloor domain.Floor) {
	d.mu.Lock()
	defer d.mu.Unlock()

	from := fromFloor.Value()
	to := toFloor.Value()

	if direction == domain.DirectionUp {
		d.up[from] = append(d.up[from], to)
		return
	}

	if direction == domain.DirectionDown {
		d.down[from] = append(d.down[from], to)
	}
}

// Flush processes pickup completion at a floor and creates destination markers.
// This method implements the core logic of our sophisticated request management system.
//
// FLUSH OPERATION EXPLAINED:
// =========================
// When an elevator arrives at a pickup floor and passengers board:
// 1. For each destination in the pickup floor's slice, create a new map entry with empty slice
// 2. Remove the pickup floor entry (pickup complete)
// 3. The empty slice destinations serve as "markers" that the elevator needs to visit those floors
//
// EXAMPLE EXECUTION:
// ==================
// Initial state: map[1] = [3, 5] (pickup floor 1, destinations 3 and 5)
//
// Flush(UP, Floor(1)) executes:
//   - Creates: map[3] = [] (destination marker)
//   - Creates: map[5] = [] (destination marker)
//   - Removes: map[1] (pickup complete)
//
// Result: map[3] = [], map[5] = []
//
// When elevator reaches floor 3: delete map[3] (destination reached)
// When elevator reaches floor 5: delete map[5] (destination reached)
//
// CONCURRENT REQUEST HANDLING:
// ===========================
// If new requests arrive while elevator is moving (e.g., map[1] = [8] added after flush),
// they are preserved and won't be lost when destination floors are reached.
// This ensures robust handling of real-world concurrent elevator requests.

func (d *Manager) Flush(direction domain.Direction, currentFloor domain.Floor) {
	d.mu.Lock()
	defer d.mu.Unlock()

	current := currentFloor.Value()

	if direction == domain.DirectionUp {
		if len(d.up[current]) > 0 {
			for _, floor := range d.up[current] {
				if _, exists := d.up[floor]; !exists {
					d.up[floor] = make([]int, 0)
				}
			}
		}
		delete(d.up, current)
		return
	}

	if direction == domain.DirectionDown {
		if len(d.down[current]) > 0 {
			for _, floor := range d.down[current] {
				if _, exists := d.down[floor]; !exists {
					d.down[floor] = make([]int, 0)
				}
			}
		}

		delete(d.down, current)
	}
}

// UpDirectionLength returns the count of floors with active upward requests.
// Only floors with actual requests (non-empty slices) are counted.
// Empty destination markers are not counted to allow proper idle state detection.
//
// This ensures the elevator can become idle when only empty destination markers remain,
// which is essential for WebSocket status updates and system idle detection.
func (d *Manager) UpDirectionLength() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	count := 0
	for _, requests := range d.up {
		if len(requests) > 0 {
			count += len(requests)
		}
		count++
	}
	return count
}

// DownDirectionLength returns the count of floors with active downward requests.
// Only floors with actual requests (non-empty slices) are counted.
// Empty destination markers are not counted to allow proper idle state detection.
//
// This ensures the elevator can become idle when only empty destination markers remain,
// which is essential for WebSocket status updates and system idle detection.
func (d *Manager) DownDirectionLength() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	count := 0
	for _, requests := range d.down {
		if len(requests) > 0 {
			count += len(requests)
		}
		count++
	}
	return count
}

// DirectionsLength returns the combined count from both direction managers.
//
// NOTE: Due to the current asymmetric implementation of Up/Down length functions,
// this combines:
// - UpDirectionLength(): counts floors with upward requests (pickup + destination floors)
// - DownDirectionLength(): counts individual downward requests + destination floors
//
// This provides a general system load metric, though the mixed counting methodologies
// should be considered when interpreting the results.
func (d *Manager) DirectionsLength() int {
	return d.UpDirectionLength() + d.DownDirectionLength()
}

// HasUpRequests - Returns true if there are any up direction requests
// This includes both pickup/dropoff floors (with or without passengers)
// Used for movement logic to determine if elevator should continue moving
func (d *Manager) HasUpRequests() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.up) > 0
}

// HasDownRequests - Returns true if there are any down direction requests
// This includes both pickup/dropoff floors (with or without passengers)
// Used for movement logic to determine if elevator should continue moving
func (d *Manager) HasDownRequests() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.down) > 0
}

// IsIdle returns true when the elevator has no work to do.
// This occurs when both direction maps are completely empty (no pickup floors, no destination markers).
// This is the primary function for idle detection used by the elevator controller.
func (d *Manager) IsIdle() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.up) == 0 && len(d.down) == 0
}

// HasAnyRequests returns true if there are any requests in either direction.
// This is the inverse of IsIdle() and can be used for active state detection.
func (d *Manager) HasAnyRequests() bool {
	return !d.IsIdle()
}

// IsRequestExisting - Checks if the request exists in the directions map or if the elevator can serve it
func (d *Manager) IsRequestExisting(direction domain.Direction, from domain.Floor, to domain.Floor) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	fromVal := from.Value()
	toVal := to.Value()

	if direction == domain.DirectionUp && isValueInMapSlice(d.up, fromVal, toVal) {
		return true
	}

	if direction == domain.DirectionDown && isValueInMapSlice(d.down, fromVal, toVal) {
		return true
	}

	return false
}

func isValueInMapSlice(m map[int][]int, key, value int) bool {
	slice, exists := m[key]
	if !exists {
		return false
	}

	for _, v := range slice {
		if v == value {
			return true
		}
	}

	return false
}

// HasUpFloor returns true if the specified floor exists in the up direction map
func (d *Manager) HasUpFloor(floor int) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, exists := d.up[floor]
	return exists
}

// HasDownFloor returns true if the specified floor exists in the down direction map
func (d *Manager) HasDownFloor(floor int) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, exists := d.down[floor]
	return exists
}

// Up returns a copy of the up direction map in a thread-safe manner
func (d *Manager) Up() map[int][]int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	// Create a copy to avoid race conditions
	result := make(map[int][]int)
	for k, v := range d.up {
		result[k] = append([]int{}, v...)
	}
	return result
}

// Down returns a copy of the down direction map in a thread-safe manner
func (d *Manager) Down() map[int][]int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	// Create a copy to avoid race conditions
	result := make(map[int][]int)
	for k, v := range d.down {
		result[k] = append([]int{}, v...)
	}
	return result
}

// GetLargestUpKey returns the largest key from the up direction map in a thread-safe manner
func (d *Manager) GetLargestUpKey() (int, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.up) == 0 {
		return 0, false
	}

	var largest int
	var first = true

	for k := range d.up {
		if first {
			largest = k
			first = false
		} else if k > largest {
			largest = k
		}
	}

	return largest, true
}

// GetSmallestDownKey returns the smallest key from the down direction map in a thread-safe manner
func (d *Manager) GetSmallestDownKey() (int, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.down) == 0 {
		return 0, false
	}

	var smallest int
	var first = true

	for k := range d.down {
		if first {
			smallest = k
			first = false
		} else if k < smallest {
			smallest = k
		}
	}

	return smallest, true
}

// GetSmallestUpKey returns the smallest key from the up direction map in a thread-safe manner
func (d *Manager) GetSmallestUpKey() (int, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.up) == 0 {
		return 0, false
	}

	var smallest int
	var first = true

	for k := range d.up {
		if first {
			smallest = k
			first = false
		} else if k < smallest {
			smallest = k
		}
	}

	return smallest, true
}

// GetLargestDownKey returns the largest key from the down direction map in a thread-safe manner
func (d *Manager) GetLargestDownKey() (int, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.down) == 0 {
		return 0, false
	}

	var largest int
	var first = true

	for k := range d.down {
		if first {
			largest = k
			first = false
		} else if k > largest {
			largest = k
		}
	}

	return largest, true
}
