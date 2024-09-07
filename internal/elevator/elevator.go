package elevator

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/slavakukuyev/elevator-go/internal/directions"
	"github.com/slavakukuyev/elevator-go/internal/infra/config"
)

type T struct {
	name              string
	minFloor          int
	maxFloor          int
	currentFloor      int
	direction         string
	mu                sync.RWMutex
	selfDirections    *directions.T
	switchOnChan      chan byte // Channel for status updates
	eachFloorDuration time.Duration
	openDoorDuration  time.Duration
	_directionDown    string
	_directionUp      string
}

func New(cfg *config.Config, name string,
	minFloor, maxFloor int,
	eachFloorDuration, openDoorDuration time.Duration) (*T, error) {

	if name == "" {
		return nil, fmt.Errorf("name can't be empty")
	}

	if minFloor == maxFloor {
		return nil, fmt.Errorf("minFloor and maxFloor can't be equal")
	}

	e := &T{
		name:              name,
		minFloor:          minFloor,
		maxFloor:          maxFloor,
		currentFloor:      0,
		selfDirections:    directions.New(cfg),
		switchOnChan:      make(chan byte, 10),
		eachFloorDuration: eachFloorDuration,
		openDoorDuration:  openDoorDuration,
		_directionDown:    cfg.DirectionDownKey,
		_directionUp:      cfg.DirectionUpKey,
	}

	//start read events process
	go e.switchOn()
	return e, nil
}

func (e *T) Name() string {
	return e.name
}

func (e *T) SetName(name string) *T {
	e.name = name
	return e
}

func (e *T) switchOn() {
	for range e.switchOnChan {
		if e.selfDirections.UpDirectionLength() > 0 || e.selfDirections.DownDirectionLength() > 0 {
			e.Run()
		}
	}
}

func (e *T) Run() {
	currentFloor := e.CurrentFloor()
	direction := e.CurrentDirection()

	slog.Debug("current floor", slog.Int("floor", currentFloor))
	time.Sleep(e.eachFloorDuration)

	if direction == e._directionUp && e.selfDirections.UpDirectionLength() > 0 {
		if _, exists := e.selfDirections.Up()[currentFloor]; exists {
			e.openDoor()
			e.selfDirections.Flush(direction, currentFloor)
			e.closeDoor()
		}

		//if elevator arrived to the top
		if currentFloor == e.maxFloor {
			e.setDirection(e._directionDown)
			return
		}

		if e.shouldMoveUp() {
			currentFloor++
			e.setCurrentFloor(currentFloor)
			go e.push()
			return
		}

	}
	//direction down && requests are down
	if direction == e._directionDown && e.selfDirections.DownDirectionLength() > 0 {
		if _, exists := e.selfDirections.Down()[currentFloor]; exists {
			e.openDoor()
			e.selfDirections.Flush(direction, currentFloor)
			e.closeDoor()
		}

		//check if elevator arrived to the bottom
		if currentFloor == e.minFloor {
			e.setDirection(e._directionUp)
			return
		}

		if e.shouldMoveDown() {
			currentFloor--
			e.setCurrentFloor(currentFloor)
			go e.push()
			return
		}

	}

	//case of elevator moving down && no more requests to move down BUT there is a request to move up on the smallest floor
	//the smallest floor of the UP direction which is smaller than current floor
	if direction == e._directionDown && e.selfDirections.UpDirectionLength() > 0 {
		smallest := findSmallestKey(e.selfDirections.Up())
		if smallest < currentFloor {
			currentFloor--
			e.setCurrentFloor(currentFloor)
			go e.push()
			return
		}

		if smallest == currentFloor {
			e.setDirection(e._directionUp)
			go e.push()
			return
		}
	}

	// the edge case when elevator moving up && there is no more requests to move up BUT new requests are existing to move down from the largest floor
	// the largest floor of the DOWN direction which is greater than current floor
	if direction == e._directionUp && e.selfDirections.DownDirectionLength() > 0 {
		largest := findLargestKey(e.selfDirections.Down())
		if largest > currentFloor {
			currentFloor++
			e.setCurrentFloor(currentFloor)
			go e.push()
			return
		}

		if largest == currentFloor {
			e.setDirection(e._directionDown)
			go e.push()
			return
		}
	}

	// the edge case when elevator moving up &&
	//  there are no requests to move above the current floor  &&
	// there are no requests to move down &&
	// there is at least one request moving up , but the elevator already above the requested floor
	if direction == e._directionUp && e.selfDirections.UpDirectionLength() > 0 && findLargestKey(e.selfDirections.Up()) < currentFloor {
		e.setDirection(e._directionDown)
		go e.push()
		return
	}

	// the edge case when elevator moving down &&
	//  there are no requests to move below the current floor  &&
	// there are no requests to move up &&
	// there is at least one request moving down , but the elevator already below the requested floor
	if direction == e._directionDown && e.selfDirections.DownDirectionLength() > 0 && findSmallestKey(e.selfDirections.Down()) > currentFloor {
		e.setDirection(e._directionUp)
		go e.push()
		return
	}

	if e.selfDirections.UpDirectionLength() == 0 && e.selfDirections.DownDirectionLength() == 0 {
		e.setDirection("")
	}
}

// check if elevator has more requests in the up direction and should continue move up
func (e *T) shouldMoveUp() bool {
	if e.selfDirections.UpDirectionLength() > 0 {
		largest := findLargestKey(e.selfDirections.Up())
		return largest > e.CurrentFloor()
	}

	return false
}

// check if elevator has more requests in the down direction and should continue move down
func (e *T) shouldMoveDown() bool {
	if e.selfDirections.DownDirectionLength() > 0 {
		smallest := findSmallestKey(e.selfDirections.Down())
		return smallest < e.CurrentFloor()
	}

	return false
}

func (e *T) openDoor() {
	slog.Info("open doors", slog.Int("floor", e.CurrentFloor()))
	time.Sleep(e.openDoorDuration)
}

func (e *T) closeDoor() {
	slog.Info("close doors", slog.Int("floor", e.CurrentFloor()))
}

func (e *T) Request(direction string, fromFloor, toFloor int) {
	currentDirection := e.CurrentDirection()
	if currentDirection == "" {
		setDirection := direction
		currentFloor := e.CurrentFloor()
		if direction == e._directionDown && currentFloor < fromFloor {
			setDirection = e._directionUp
		} else if direction == e._directionUp && currentFloor > fromFloor {
			setDirection = e._directionDown
		}

		e.setDirection(setDirection)
	}

	e.selfDirections.Append(direction, fromFloor, toFloor)
	go e.push()
}

func (e *T) CurrentDirection() string {
	e.mu.RLock()
	direction := e.direction
	e.mu.RUnlock()
	return direction
}

func (e *T) setDirection(direction string) {
	e.mu.Lock()
	e.direction = direction
	e.mu.Unlock()
}

func (e *T) CurrentFloor() int {
	e.mu.RLock()
	currentFloor := e.currentFloor
	e.mu.RUnlock()
	return currentFloor
}

func (e *T) setCurrentFloor(floor int) {
	e.mu.Lock()
	e.currentFloor = floor
	e.mu.Unlock()
}

func (e *T) Directions() *directions.T {
	e.mu.RLock()
	d := e.selfDirections
	e.mu.RUnlock()
	return d
}

func (e *T) push() {
	e.switchOnChan <- 1
}

func (e *T) IsRequestInRange(fromFloor, toFloor int) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return fromFloor >= e.minFloor && fromFloor <= e.maxFloor && toFloor >= e.minFloor && toFloor <= e.maxFloor
}

func findLargestKey(m map[int][]int) int {
	largest := 0

	for key := range m {
		if key > largest {
			largest = key
		}
	}

	return largest
}

func findSmallestKey(m map[int][]int) int {
	smallest := 0
	first := true

	for key := range m {
		if first || key < smallest {
			smallest = key
			first = false
		}
	}

	return smallest
}

func (e *T) MinFloor() int {
	return e.minFloor
}

func (e *T) MaxFloor() int {
	return e.maxFloor
}
