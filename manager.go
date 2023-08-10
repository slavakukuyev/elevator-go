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

// Algorithm for a  few elevators:
// Lets assume the max and min flloors are the same
//
//  Unique direction (==1)  and floor is appropriate
// Same direction: the shortest way to requested floor
// No direction
// Opposite directions: the less requests in the requested direction

func (m *Manager) RequestElevator(fromFloor, toFloor int) error {

	if toFloor == fromFloor {
		return fmt.Errorf("the requested floor (%d) should be different from your floor (%d)", toFloor, fromFloor)
	}

	direction := _directionUp
	if toFloor < fromFloor {
		direction = _directionDown
	}

	m.mu.RLock()
	elevators := m.elevators
	m.mu.RUnlock()

	elevator := m.chooseElevator(elevators, direction, fromFloor)
	elevator.Request(direction, fromFloor, toFloor)

	return nil

}

func (m *Manager) chooseElevator(elevators []*Elevator, requestedDirection string, fromFloor int) *Elevator {
	directions := make(map[*Elevator]string)

	//case when elevator is waiting to start
	for _, e := range elevators {
		d := e.GetDirection()
		if d == "" {
			return e
		}
		directions[e] = d
	}

	filteredElevators := countMatchingDirections(directions, requestedDirection)

	//case when single elevator with the same direction
	//should validate if the elevator still on his way to the floor
	if len(filteredElevators) == 1 {
		e := filteredElevators[0]
		currentFloor := e.GetCurrentFloor()

		if (requestedDirection == _directionUp && currentFloor < fromFloor) ||
			(requestedDirection == _directionDown && currentFloor > fromFloor) {
			return e
		}
	}

	//case when more then one elevator with the same direction
	// should check smallest number between currentfloor and requested floor
	if len(filteredElevators) > 1 {
		var first bool = true
		var smallest int
		var nearestE *Elevator

		for _, e := range filteredElevators {
			currentFloor := e.GetCurrentFloor()

			if requestedDirection == _directionUp && currentFloor < fromFloor {
				diff := fromFloor - currentFloor
				if first || (smallest > diff) {
					nearestE = e
					smallest = diff
					first = false
				}
			} else if requestedDirection == _directionDown && currentFloor > fromFloor {
				diff := currentFloor - fromFloor
				if first || (smallest > diff) {
					nearestE = e
					smallest = diff
					first = false
				}
			}
		}

		if nearestE != nil {
			return nearestE
		}

	}

	//default response will not stuck elevators -> at least one will work
	return elevators[0]

}

func countMatchingDirections(directions map[*Elevator]string, requestedDirection string) []*Elevator {
	elevators := make([]*Elevator, 0, len(directions))
	for e, sourceDirection := range directions {
		if sourceDirection == requestedDirection {
			elevators = append(elevators, e)
		}
	}
	return elevators
}
