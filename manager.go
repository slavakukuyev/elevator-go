package main

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
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
	logger.Info("Request has been approved", zap.String("elevator", elevator.name), zap.Int("fromFloor", fromFloor), zap.Int("toFloor", toFloor))
	return nil

}

func (m *Manager) chooseElevator(elevators []*Elevator, requestedDirection string, fromFloor int) *Elevator {
	directions := make(map[*Elevator]string)

	//case when elevator is waiting to start
	for _, e := range elevators {
		d := e.CurrentDirection()
		if d == "" {
			return e
		}
		directions[e] = d
	}

	/******** SAME DIRECTION ********/

	filteredElevators := elevatorsMatchingDirections(directions, requestedDirection)

	//case when single elevator with the same direction
	//should validate if the elevator still on his way to the floor
	if len(filteredElevators) == 1 {
		e := filteredElevators[0]
		currentFloor := e.CurrentFloor()

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
			currentFloor := e.CurrentFloor()

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

		//all the elevators in the same direction already passed the requested floor
		//find the one with less requests in both directions for now
		e := elevatorWithMinRequestsByDirection(elevators, "")
		if e != nil {
			return e
		}

	}
	/******** OPPOSITE DIRECTION ********/

	filteredElevators = elevatorsOppositeDirections(directions, requestedDirection)

	//if only one found, then the previous conditions didn't work
	//then return this single filtered elevator, because:
	// * the other elevators already passed the floors
	// * this one will finish its opposite direction first and then will switch to required one
	if len(filteredElevators) == 1 {
		return filteredElevators[0]
	}

	if len(filteredElevators) > 1 {
		var e *Elevator
		if requestedDirection == _directionUp {
			e = elevatorWithMinRequestsByDirection(elevators, _directionDown)
		} else if requestedDirection == _directionDown {
			e = elevatorWithMinRequestsByDirection(elevators, _directionUp)
		}

		if e != nil {
			return e
		}

	}

	//default response will not stuck elevators -> at least one will work
	return elevators[0]

}

func elevatorsMatchingDirections(directions map[*Elevator]string, requestedDirection string) []*Elevator {
	elevators := make([]*Elevator, 0, len(directions))
	for e, sourceDirection := range directions {
		if sourceDirection == requestedDirection {
			elevators = append(elevators, e)
		}
	}
	return elevators
}

func elevatorsOppositeDirections(directions map[*Elevator]string, requestedDirection string) []*Elevator {
	elevators := make([]*Elevator, 0, len(directions))
	for e, sourceDirection := range directions {
		if sourceDirection != requestedDirection {
			elevators = append(elevators, e)
		}
	}
	return elevators
}

func elevatorWithMinRequestsByDirection(elevators []*Elevator, direction string) *Elevator {
	var elevator *Elevator
	var smallest int
	var first bool = true

	for _, e := range elevators {
		directions := e.Directions()
		l := 0
		switch direction {
		case _directionUp:
			l = directions.UpDirectionLength()
		case _directionDown:
			l = directions.DownDirectionLength()
		default:
			l = directions.UpDirectionLength() + directions.DownDirectionLength()
		}

		if first || (smallest < l) {
			smallest = l
			elevator = e
			first = false
		}
	}

	return elevator
}
