package main

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

var logger *zap.Logger

func initLogger() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync() // Flushes buffer, if any
}

const _directionUp = "up"
const _directionDown = "down"

// ***************************************************************************

type Destinations struct {
	up   map[int][]int
	down map[int][]int
	mu   sync.RWMutex
}

func NewDestinations() *Destinations {
	return &Destinations{
		up:   make(map[int][]int),
		down: make(map[int][]int),
	}
}

func (d *Destinations) Append(direction string, fromFloor, toFloor int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if direction == _directionUp {
		d.up[fromFloor] = append(d.up[fromFloor], toFloor)
		return
	}

	if direction == _directionDown {
		d.down[fromFloor] = append(d.down[fromFloor], toFloor)
	}
}

func (d *Destinations) Flush(direction string, fromFloor int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if direction == _directionUp {
		if len(d.up[fromFloor]) > 0 {
			for _, floor := range d.up[fromFloor] {
				if _, exists := d.up[floor]; !exists {
					d.up[floor] = make([]int, 0)
				}
			}

		}

		delete(d.up, fromFloor)
	}

}

func (d *Destinations) isUpExisting() bool {
	d.mu.RLock()
	existing := len(d.up) > 0
	d.mu.RUnlock()
	return existing
}

func (d *Destinations) isDownExisting() bool {
	d.mu.RLock()
	existing := len(d.down) > 0
	d.mu.RUnlock()
	return existing
}

//***************************************************************************

type Elevator struct {
	name         string
	maxFloor     int
	minFloor     int
	currentFloor int
	direction    string
	mu           sync.RWMutex
	destinations *Destinations
	switchOnChan chan byte // Channel for status updates
}

func NewElevator(name string) *Elevator {
	e := &Elevator{
		name:         name,
		maxFloor:     9,
		minFloor:     0,
		currentFloor: 0,
		destinations: NewDestinations(),
		switchOnChan: make(chan byte, 10),
	}

	go e.switchOn()
	return e
}

func (e *Elevator) switchOn() {
	for range e.switchOnChan {
		e.Run()
	}
}

func (e *Elevator) Run() {
	e.mu.Lock()
	defer e.mu.Unlock()

	fmt.Printf("The elevator %s is on the %d floor\n", e.name, e.currentFloor)

	if e.direction == _directionUp && e.destinations.isUpExisting() {
		if _, exists := e.destinations.up[e.currentFloor]; exists {
			e.openDoor()
			e.destinations.Flush(e.direction, e.currentFloor)
			e.closeDoor()

		}

		e.currentFloor++
		e.switchOnChan <- 1
	}

	if e.direction == _directionDown && e.destinations.isDownExisting() {
		if _, exists := e.destinations.down[e.currentFloor]; exists {
			e.openDoor()
			e.destinations.Flush(e.direction, e.currentFloor)
			e.closeDoor()
			e.currentFloor++
			e.switchOnChan <- 1
		}
	}
}

func (e *Elevator) openDoor() {
	fmt.Printf("Elevator %s opened the doors at floor %d\n", e.name, e.currentFloor)
}

func (e *Elevator) closeDoor() {
	fmt.Printf("Elevator %s closed the doors at floor %d\n", e.name, e.currentFloor)
}

// Append the request to the elevator regardless of the current direction if `must` is `true`.
// If the direction is empty, set the requested direction.
// If the current direction is opposite, reject the request.
func (e *Elevator) Request(direction string, fromFloor, toFloor int, must bool) bool {
	currentDirection := e.GetDirection()

	if !must && currentDirection != "" && currentDirection != direction {
		return false
	}

	if currentDirection == "" {
		e.SetDirection(direction)
	}

	e.destinations.Append(direction, fromFloor, toFloor)
	e.switchOnChan <- 1
	return true
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
	mu        sync.RWMutex
	elevators []*Elevator
}

func NewManager() *Manager {
	return &Manager{
		elevators: make([]*Elevator, 0),
	}
}

func (m *Manager) AddElevator(elevator *Elevator) {
	m.elevators = append(m.elevators, elevator)
}

func (m *Manager) RequestElevator(fromFloor, toFloor int) error {

	if toFloor == fromFloor {
		return fmt.Errorf("%d floor is equal to the same %d floor. The requested floor should be different from your floor", toFloor, fromFloor)
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

func main() {

	initLogger()

	manager := NewManager()

	elevator1 := NewElevator("E1")
	// elevator2 := NewElevator()

	manager.AddElevator(elevator1)
	// manager.AddElevator(elevator2)

	// Request an elevator going from floor 1 to floor 9
	if err := manager.RequestElevator(1, 9); err != nil {
		fmt.Println(err)
	}

	// Request an elevator going from floor 3 to floor 5
	if err := manager.RequestElevator(3, 5); err != nil {
		fmt.Println(err)
	}

	for {
		time.Sleep(time.Second)
	}
}
