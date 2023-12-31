package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Manager struct {
	mu        sync.RWMutex
	elevators []*Elevator
	logger    *zap.Logger
}

func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		elevators: make([]*Elevator, 0),
		logger:    logger,
	}
}

func (m *Manager) AddElevator(elevator *Elevator) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.elevators = append(m.elevators, elevator)
	m.logger.Info("new elevator added to the managment pool", zap.String("elevator", elevator.name))
}

func (m *Manager) RequestElevator(fromFloor, toFloor int) (*Elevator, error) {

	if toFloor == fromFloor {
		return nil, fmt.Errorf("the requested floor (%d) should be different from your floor (%d)", toFloor, fromFloor)
	}

	direction := _directionUp
	if toFloor < fromFloor {
		direction = _directionDown
	}

	m.mu.RLock()
	elevators := m.elevators
	m.mu.RUnlock()

	var elevator *Elevator

	// validate existing requests
	if elevator = requestedElevator(elevators, direction, fromFloor, toFloor); elevator != nil {
		return elevator, nil
	}

	elevator, err := m.chooseElevator(elevators, direction, fromFloor, toFloor)
	if err != nil {
		return nil, err
	}

	elevator.Request(direction, fromFloor, toFloor)
	m.logger.Info("request has been approved", zap.String("elevator", elevator.name), zap.Int("fromFloor", fromFloor), zap.Int("toFloor", toFloor))
	return elevator, nil

}

func requestedElevator(elevators []*Elevator, direction string, fromFloor, toFloor int) *Elevator {
	for _, e := range elevators {
		if e.directions.IsExisting(direction, fromFloor, toFloor) {
			return e
		}
	}

	return nil
}

func (m *Manager) chooseElevator(elevators []*Elevator, requestedDirection string, fromFloor, toFloor int) (*Elevator, error) {
	elevatorsWaiting := make(map[*Elevator]int)
	elevatorsByDirection := make(map[*Elevator]string)

	//case when elevator is waiting to start
	for _, e := range elevators {
		if !e.IsRequestInRange(fromFloor, toFloor) {
			continue
		}

		d := e.CurrentDirection()
		if d == "" {
			elevatorsWaiting[e] = e.CurrentFloor()
		} else {
			elevatorsByDirection[e] = d
		}
	}

	if len(elevatorsWaiting) > 0 {
		if e := findNearestElevator(elevatorsWaiting, fromFloor); e != nil {
			return e, nil
		}
	}

	if len(elevatorsByDirection) == 0 {
		return nil, fmt.Errorf("the requested floors (%d, %d) should be in range of existing elevators", fromFloor, toFloor)
	}

	/******** SAME DIRECTION ********/

	filteredElevators := elevatorsMatchingDirections(elevatorsByDirection, requestedDirection)

	//case when single elevator with the same direction
	//should validate if the elevator still on his way to the floor
	if len(filteredElevators) == 1 {
		e := filteredElevators[0]
		currentFloor := e.CurrentFloor()

		if (requestedDirection == _directionUp && currentFloor < fromFloor) ||
			(requestedDirection == _directionDown && currentFloor > fromFloor) {
			return e, nil
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
			return nearestE, nil
		}

		//all the elevators in the same direction already passed the requested floor
		//find the one with less requests in both directions for now
		e := elevatorWithMinRequestsByDirection(elevators, "")
		if e != nil {
			return e, nil
		}

	}
	/******** OPPOSITE DIRECTION ********/

	filteredElevators = elevatorsOppositeDirections(elevatorsByDirection, requestedDirection)

	//if only one found, then the previous conditions didn't work
	//then return this single filtered elevator, because:
	// * the other elevators already passed the floors
	// * this one will finish its opposite direction first and then will switch to required one
	filteredElevatorsLength := len(filteredElevators)
	if filteredElevatorsLength == 1 {
		return filteredElevators[0], nil
	}

	if filteredElevatorsLength > 1 {
		var e *Elevator
		if requestedDirection == _directionUp {
			e = elevatorWithMinRequestsByDirection(elevators, _directionDown)
		} else if requestedDirection == _directionDown {
			e = elevatorWithMinRequestsByDirection(elevators, _directionUp)
		}

		if e != nil {
			return e, nil
		}
	}

	//default response will not stuck elevators -> at least one will work
	for e := range elevatorsByDirection {
		return e, nil
	}

	return nil, fmt.Errorf("no elevator found for reqeusted floors: fromFloor(%d) toFloor(%d) [WTF: One more case]", fromFloor, toFloor)
}

func elevatorsMatchingDirections(elevatorsByDirection map[*Elevator]string, requestedDirection string) []*Elevator {
	elevators := make([]*Elevator, 0, len(elevatorsByDirection))
	for e, sourceDirection := range elevatorsByDirection {
		if sourceDirection == requestedDirection {
			elevators = append(elevators, e)
		}
	}
	return elevators
}

func elevatorsOppositeDirections(elevatorsByDirection map[*Elevator]string, requestedDirection string) []*Elevator {
	elevators := make([]*Elevator, 0, len(elevatorsByDirection))
	for e, sourceDirection := range elevatorsByDirection {
		if sourceDirection != requestedDirection {
			elevators = append(elevators, e)
		}
	}
	return elevators
}

func floorsDiff(floor, requestedFloor int) int {
	if floor < requestedFloor {
		return requestedFloor - floor
	}

	if floor > requestedFloor {
		return floor - requestedFloor
	}

	return 0
}

func findNearestElevator(elevatorsWaiting map[*Elevator]int, requestedFloor int) *Elevator {
	elevatorsLength := len(elevatorsWaiting)
	if elevatorsLength == 0 {
		return nil
	}

	if elevatorsLength == 1 {
		for elevator := range elevatorsWaiting {
			return elevator
		}
	}
	var minDistanceElevators []*Elevator
	minDistance := -1

	for elevator, floor := range elevatorsWaiting {
		distance := floorsDiff(floor, requestedFloor)

		// If it's the first key or has the same minimum distance, add it to the list.
		if minDistance == -1 || distance == minDistance {
			minDistanceElevators = append(minDistanceElevators, elevator)
			minDistance = distance
		} else if distance < minDistance {
			// If it's closer than the previous ones, reset the list.
			minDistanceElevators = []*Elevator{elevator}
			minDistance = distance
		}
	}

	// Randomly choose one of the keys with the same minimum distance.
	if len(minDistanceElevators) > 0 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomIndex := r.Intn(len(minDistanceElevators))
		return minDistanceElevators[randomIndex]
	}

	return nil
}

// elevatorWithMinRequestsByDirection selects an elevator with the minimum number of pending requests
// in the specified direction from the given slice of elevators.
// If the direction is empty, it selects the elevator with the overall minimum number of requests.
// Parameters:
// - elevators: A slice of elevators to choose from.
// - direction: The requested direction ("up", "down", or empty for any direction).
// Returns:
// - An Elevator pointer representing the selected elevator.
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
