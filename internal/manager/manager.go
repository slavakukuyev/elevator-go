package manager

import (
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/slavakukuyev/elevator-go/internal/elevator"
	"github.com/slavakukuyev/elevator-go/internal/infra/config"
)

const _directionUp = "up"
const _directionDown = "down"

type Manager struct {
	mu        sync.RWMutex
	elevators []*elevator.Elevator
	factory   elevator.ElevatorFactory
}

func NewManager(cfg *config.Config, factory elevator.ElevatorFactory) *Manager {
	return &Manager{
		elevators: []*elevator.Elevator{},
		factory:   factory,
	}
}

func (m *Manager) AddElevator(cfg *config.Config, name string,
	minFloor, maxFloor int,
	eachFloorDuration, openDoorDuration time.Duration) error {
	elevator, err := m.factory.CreateElevator(cfg, name,
		minFloor, maxFloor,
		eachFloorDuration, openDoorDuration)
	if err != nil {
		return fmt.Errorf("error on initialization new elevator: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.elevators = append(m.elevators, elevator)
	slog.Info("new elevator added to the managment pool", slog.String("elevator", elevator.Name()))
	return nil
}

func (m *Manager) RequestElevator(fromFloor, toFloor int) (*elevator.Elevator, error) {

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

	var elevator *elevator.Elevator

	// validate existing requests
	if elevator = requestedElevator(elevators, direction, fromFloor, toFloor); elevator != nil {
		return elevator, nil
	}

	elevator, err := m.chooseElevator(elevators, direction, fromFloor, toFloor)
	if err != nil {
		return nil, err
	}

	elevator.Request(direction, fromFloor, toFloor)
	slog.Info("request has been approved", slog.String("elevator", elevator.Name()), slog.Int("fromFloor", fromFloor), slog.Int("toFloor", toFloor))
	return elevator, nil

}

func requestedElevator(elevators []*elevator.Elevator, direction string, fromFloor, toFloor int) *elevator.Elevator {
	for _, e := range elevators {
		if e.Directions().IsExisting(direction, fromFloor, toFloor) {
			return e
		}
	}

	return nil
}

func (m *Manager) chooseElevator(elevators []*elevator.Elevator, requestedDirection string, fromFloor, toFloor int) (*elevator.Elevator, error) {
	elevatorsWaiting := make(map[*elevator.Elevator]int)
	elevatorsByDirection := make(map[*elevator.Elevator]string)

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
		var nearestE *elevator.Elevator

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
		var e *elevator.Elevator
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

func elevatorsMatchingDirections(elevatorsByDirection map[*elevator.Elevator]string, requestedDirection string) []*elevator.Elevator {
	elevators := make([]*elevator.Elevator, 0, len(elevatorsByDirection))
	for e, sourceDirection := range elevatorsByDirection {
		if sourceDirection == requestedDirection {
			elevators = append(elevators, e)
		}
	}
	return elevators
}

func elevatorsOppositeDirections(elevatorsByDirection map[*elevator.Elevator]string, requestedDirection string) []*elevator.Elevator {
	elevators := make([]*elevator.Elevator, 0, len(elevatorsByDirection))
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

func findNearestElevator(elevatorsWaiting map[*elevator.Elevator]int, requestedFloor int) *elevator.Elevator {
	elevatorsLength := len(elevatorsWaiting)
	if elevatorsLength == 0 {
		return nil
	}

	if elevatorsLength == 1 {
		for elevator := range elevatorsWaiting {
			return elevator
		}
	}
	var minDistanceElevators []*elevator.Elevator
	minDistance := -1

	for el, floor := range elevatorsWaiting {
		distance := floorsDiff(floor, requestedFloor)

		// If it's the first key or has the same minimum distance, add it to the list.
		if minDistance == -1 || distance == minDistance {
			minDistanceElevators = append(minDistanceElevators, el)
			minDistance = distance
		} else if distance < minDistance {
			// If it's closer than the previous ones, reset the list.
			minDistanceElevators = []*elevator.Elevator{el}
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
func elevatorWithMinRequestsByDirection(elevators []*elevator.Elevator, direction string) *elevator.Elevator {
	var elevator *elevator.Elevator
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
