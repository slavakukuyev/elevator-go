package main

import (
	"fmt"
	"sync"
)

const _directionUp = "up"
const _directionDown = "down"

// ***************************************************************************

type Destinations struct {
	up   map[int]int
	down map[int]int
	mu   sync.RWMutex
}

func NewDestinations() *Destinations {
	return &Destinations{
		up:   make(map[int]int),
		down: make(map[int]int),
	}
}

func (d *Destinations) isUpEmpty() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	for _, val := range d.up {
		if val == 1 {
			return false
		}
	}
	return true
}

func (d *Destinations) isDownEmpty() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	for _, val := range d.down {
		if val == 1 {
			return false
		}
	}
	return true
}

//***************************************************************************

type Elevator struct {
	maxFloor     int
	minFloor     int
	currentFloor int
	direction    string
	mu           sync.RWMutex
	moving       bool
	destinations *Destinations
}

func NewElevator(destinations *Destinations) *Elevator {
	return &Elevator{
		maxFloor:     9,
		minFloor:     0,
		currentFloor: 0,
		destinations: destinations,
	}
}

func (e *Elevator) Move() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.moving {
		e.moving = true
		e.Move()
	}
}

func (e *Elevator) isMoving() bool {
	e.mu.RLock()
	moving := e.moving
	e.mu.RUnlock()
	return moving
}

// if current direction is empty
func (e *Elevator) Request() {
	if e.isMoving() {
		return
	}

	currentDirection := e.GetDirection()

	if currentDirection != "" {
		return
	}

	if !e.destinations.isUpEmpty() {
		e.SetDirection(_directionUp)
	} else if !e.destinations.isDownEmpty() {
		e.SetDirection(_directionDown)
	}

	e.Move()

}

func (e *Elevator) GetDirection() string {
	e.mu.RLock()
	direction := e.direction
	e.mu.RUnlock()
	return direction
}

func (e *Elevator) SetDirection(direction string) {
	e.mu.Lock()
	e.direction = direction
	e.mu.Unlock()
}

// ***********************************************************************************************************
type Manager struct {
	destinations *Destinations
	elevators    []*Elevator
}

func NewManager(destinations *Destinations) *Manager {
	return &Manager{
		destinations: destinations,
		elevators:    make([]*Elevator, 0),
	}
}

func (m *Manager) AddElevator(elevator *Elevator) {
	m.elevators = append(m.elevators, elevator)
}

func (m *Manager) RequestElevator(fromFloor, toFloor int) {

	m.destinations.mu.Lock()
	if toFloor > fromFloor {
		m.destinations.up[fromFloor] = 1
		m.destinations.up[toFloor] = 1
	} else if toFloor < fromFloor {
		m.destinations.down[fromFloor] = 1
		m.destinations.down[toFloor] = 1
	}
	m.destinations.mu.Unlock()
	m.Ping()
}

func (m *Manager) Ping() {
	for _, e := range m.elevators {
		go e.Request()
	}
}

func main() {
	destinations := NewDestinations()
	manager := NewManager(destinations)

	elevator1 := NewElevator(destinations)
	// elevator2 := NewElevator()

	manager.AddElevator(elevator1)
	// manager.AddElevator(elevator2)

	// Request an elevator going from floor 1 to floor 9
	manager.RequestElevator(1, 9)

	// Request an elevator going from floor 3 to floor 5
	manager.RequestElevator(3, 5)

	fmt.Println("Elevators:", manager.elevators)
	fmt.Println("Destinations (Up):", manager.destinations.up)
	fmt.Println("Destinations (Down):", manager.destinations.down)
}
