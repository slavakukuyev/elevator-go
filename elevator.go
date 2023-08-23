package main

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type Elevator struct {
	name         string
	maxFloor     int
	minFloor     int
	currentFloor int
	direction    string
	mu           sync.RWMutex
	directions   *Directions
	switchOnChan chan byte // Channel for status updates
}

func NewElevator(name string) *Elevator {
	e := &Elevator{
		name:         name,
		maxFloor:     9,
		minFloor:     0,
		currentFloor: 0,
		directions:   NewDirections(),
		switchOnChan: make(chan byte, 10),
	}

	go e.switchOn()
	return e
}

func (e *Elevator) switchOn() {
	for range e.switchOnChan {
		if e.directions.UpDirectionLength() > 0 || e.directions.DownDirectionLength() > 0 {
			e.Run()
		}
	}
}

func (e *Elevator) Run() {
	currentFloor := e.CurrentFloor()
	direction := e.CurrentDirection()

	logger.Debug("current floor", zap.String("elevator", e.name), zap.Int("floor", currentFloor))
	time.Sleep(time.Millisecond * 500)

	if direction == _directionUp && e.directions.UpDirectionLength() > 0 {
		if _, exists := e.directions.up[currentFloor]; exists {
			e.openDoor()
			e.directions.Flush(direction, currentFloor)
			e.closeDoor()
		}

		//if elevator arrived to the top
		if currentFloor == e.maxFloor {
			e.setDirection(_directionDown)
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
	if direction == _directionDown && e.directions.DownDirectionLength() > 0 {
		if _, exists := e.directions.down[currentFloor]; exists {
			e.openDoor()
			e.directions.Flush(direction, currentFloor)
			e.closeDoor()
		}

		//check if elevator arrived to the bottom
		if currentFloor == e.minFloor {
			e.setDirection(_directionUp)
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
	//smallest floor of the UP direction which is smaller then current floor
	if direction == _directionDown && e.directions.UpDirectionLength() > 0 {
		smallest := findSmallestKey(e.directions.up)
		if smallest < currentFloor {
			currentFloor--
			e.setCurrentFloor(currentFloor)
			go e.push()
			return
		}

		if smallest == currentFloor {
			e.setDirection(_directionUp)
			go e.push()
			return
		}
	}

	//the edge case when elevator moving up && there is no more requests to move up BUT new requests are existing to movedown from the largest floor
	//largest floor of the DOWN direction which is greater then current floor
	if direction == _directionUp && e.directions.DownDirectionLength() > 0 {
		largest := findLargestKey(e.directions.down)
		if largest > currentFloor {
			currentFloor++
			e.setCurrentFloor(currentFloor)
			go e.push()
			return
		}

		if largest == currentFloor {
			e.setDirection(_directionDown)
			go e.push()
			return
		}
	}

	// the edge case when elevator moving up &&
	//  thre are no requests to move above the current floor  &&
	// there are no requests to move down &&
	// there is at least one request moving up , but the elavator already above the requiested floor
	if direction == _directionUp && e.directions.UpDirectionLength() > 0 && findLargestKey(e.directions.up) < currentFloor {
		e.setDirection(_directionDown)
		go e.push()
		return
	}

	// the edge case when elevator moving down &&
	//  thre are no requests to move below the current floor  &&
	// there are no requests to move up &&
	// there is at least one request moving down , but the elavator already below the requested floor
	if direction == _directionDown && e.directions.DownDirectionLength() > 0 && findSmallestKey(e.directions.down) > currentFloor {
		e.setDirection(_directionUp)
		go e.push()
		return
	}

	if e.directions.UpDirectionLength() == 0 && e.directions.DownDirectionLength() == 0 {
		e.setDirection("")
	}
}

// check if elevator has more requests in the up direction and should continue move up
func (e *Elevator) shouldMoveUp() bool {
	if e.directions.UpDirectionLength() > 0 {
		largest := findLargestKey(e.directions.up)
		return largest > e.currentFloor
	}

	return false
}

// check if elevator has more requests in the down direction and should continue move down
func (e *Elevator) shouldMoveDown() bool {
	if e.directions.DownDirectionLength() > 0 {
		smallest := findSmallestKey(e.directions.down)
		return smallest < e.currentFloor
	}

	return false
}

func (e *Elevator) openDoor() {
	logger.Info("open doors", zap.String("elevator", e.name), zap.Int("floor", e.currentFloor))
	time.Sleep(time.Second * 2)
}

func (e *Elevator) closeDoor() {
	logger.Info("close doors", zap.String("elevator", e.name), zap.Int("floor", e.currentFloor))
}

func (e *Elevator) Request(direction string, fromFloor, toFloor int) bool {
	currentDirection := e.CurrentDirection()
	if currentDirection == "" {
		setDirection := direction
		currentFloor := e.CurrentFloor()
		if direction == _directionDown && currentFloor < fromFloor {
			setDirection = _directionUp
		} else if direction == _directionUp && currentFloor > fromFloor {
			setDirection = _directionDown
		}

		e.setDirection(setDirection)
	}

	e.directions.Append(direction, fromFloor, toFloor)
	go e.push()
	return true
}

func (e *Elevator) CurrentDirection() string {
	e.mu.RLock()
	direction := e.direction
	e.mu.RUnlock()
	return direction
}

func (e *Elevator) setDirection(direction string) {
	e.mu.Lock()
	e.direction = direction
	e.mu.Unlock()
}

func (e *Elevator) CurrentFloor() int {
	e.mu.RLock()
	currentFloor := e.currentFloor
	e.mu.RUnlock()
	return currentFloor
}

func (e *Elevator) setCurrentFloor(floor int) {
	e.mu.Lock()
	e.currentFloor = floor
	e.mu.Unlock()
}

func (e *Elevator) Directions() *Directions {
	e.mu.RLock()
	d := e.directions
	e.mu.RUnlock()
	return d
}

func (e *Elevator) push() {
	e.switchOnChan <- 1
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
