package elevator

import (
	"sync"

	"github.com/slavakukuyev/elevator-go/internal/domain"
)

// State manages the internal state of an elevator
type State struct {
	mu           sync.RWMutex
	name         string
	currentFloor domain.Floor
	direction    domain.Direction
	minFloor     domain.Floor
	maxFloor     domain.Floor
}

// NewState creates a new elevator state
func NewState(name string, minFloor, maxFloor domain.Floor) *State {
	return &State{
		name:         name,
		currentFloor: minFloor, // Start at minimum floor
		direction:    domain.DirectionIdle,
		minFloor:     minFloor,
		maxFloor:     maxFloor,
	}
}

// Name returns the name of the elevator
func (s *State) Name() string {
	return s.name
}

// SetName sets the name of the elevator
func (s *State) SetName(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.name = name
}

// CurrentFloor returns the current floor
func (s *State) CurrentFloor() domain.Floor {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentFloor
}

// SetCurrentFloor sets the current floor
func (s *State) SetCurrentFloor(floor domain.Floor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentFloor = floor
}

// Direction returns the current direction
func (s *State) Direction() domain.Direction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.direction
}

// SetDirection sets the current direction
func (s *State) SetDirection(direction domain.Direction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.direction = direction
}

// MinFloor returns the minimum floor
func (s *State) MinFloor() domain.Floor {
	return s.minFloor
}

// MaxFloor returns the maximum floor
func (s *State) MaxFloor() domain.Floor {
	return s.maxFloor
}

// IsAtTopFloor checks if elevator is at the top floor
func (s *State) IsAtTopFloor() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentFloor.IsEqual(s.maxFloor)
}

// IsAtBottomFloor checks if elevator is at the bottom floor
func (s *State) IsAtBottomFloor() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentFloor.IsEqual(s.minFloor)
}

// IsFloorInRange checks if a floor is within the elevator's range
func (s *State) IsFloorInRange(floor domain.Floor) bool {
	return floor.IsValid(s.minFloor, s.maxFloor)
}

// GetStatus returns the current status of the elevator
func (s *State) GetStatus(requestCount int) domain.ElevatorStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return domain.NewElevatorStatus(
		s.name,
		s.currentFloor,
		s.direction,
		requestCount,
		s.minFloor,
		s.maxFloor,
	)
}
