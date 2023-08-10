package main

import (
	"fmt"
	"sync"
)

type Manager struct {
	mu        sync.RWMutex
	elevators []*Elevator
}

func NewManager() *Manager {
	return &Manager{
		elevators: make([]*Elevator, 0),
	}
}

func (m *Manager) AddElevator(elevator *Elevator) {
	m.mu.Lock()
	m.elevators = append(m.elevators, elevator)
	m.mu.Unlock()
}

func (m *Manager) RequestElevator(fromFloor, toFloor int) error {

	if toFloor == fromFloor {
		return fmt.Errorf("the requested floor (%d) should be different from your floor (%d)", toFloor, fromFloor)
	}

	direction := _directionUp
	if toFloor < fromFloor {
		direction = _directionDown
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	approved := false
	for _, elevator := range m.elevators {
		approved = elevator.Request(direction, fromFloor, toFloor, false)
	}

	if !approved {
		m.elevators[0].Request(direction, fromFloor, toFloor, true)
	}

	return nil

}
