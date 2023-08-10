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
		if e.directions.isUpExisting() || e.directions.isDownExisting() {
			e.Run()
		}
	}
}

func (e *Elevator) Run() {
	e.mu.Lock()
	defer e.mu.Unlock()

	logger.Debug("current floor", zap.String("elevator", e.name), zap.Int("floor", e.currentFloor))
	time.Sleep(time.Millisecond * 500)

	if e.direction == _directionUp && e.directions.isUpExisting() {
		if _, exists := e.directions.up[e.currentFloor]; exists {
			e.openDoor()
			e.directions.Flush(e.direction, e.currentFloor)
			e.closeDoor()
		}

		//if elevator arrived to the top
		if e.currentFloor == e.maxFloor {
			e.direction = _directionDown
			return
		}

		if e.shouldMoveUp() {
			e.currentFloor++
			e.switchOnChan <- 1
			return
		}

	}
	//direction down && requests are down
	if e.direction == _directionDown && e.directions.isDownExisting() {
		if _, exists := e.directions.down[e.currentFloor]; exists {
			e.openDoor()
			e.directions.Flush(e.direction, e.currentFloor)
			e.closeDoor()
		}

		//check if elevator arrived to the bottom
		if e.currentFloor == e.minFloor {
			e.direction = _directionUp
			return
		}

		if e.shouldMoveDown() {
			e.currentFloor--
			e.switchOnChan <- 1
			return
		}

	}

	//case of elevator moving down && no more requests to move down BUT there is a request to move up on the smallest floor
	//smallest floor of the UP direction which is smaller then current floor
	if e.direction == _directionDown && e.directions.isUpExisting() {
		smallest := findSmallestKey(e.directions.up)
		if smallest < e.currentFloor {
			e.currentFloor--
			e.switchOnChan <- 1
			return
		}

		if smallest == e.currentFloor {
			e.direction = _directionUp
			e.switchOnChan <- 1
			return
		}
	}

	//the edge case when elevator moving up && there is no more requests to move up BUT new requests are existing to movedown from the largest floor
	//largest floor of the DOWN direction which is greater then current floor
	if e.direction == _directionUp && e.directions.isDownExisting() {
		largest := findLargestKey(e.directions.down)
		if largest > e.currentFloor {
			e.currentFloor++
			e.switchOnChan <- 1
			return
		}

		if largest == e.currentFloor {
			e.direction = _directionDown
			e.switchOnChan <- 1
			return
		}
	}

}

// check if elevator has more requests in the up direction and should continue move up
func (e *Elevator) shouldMoveUp() bool {
	if e.directions.isUpExisting() {
		largest := findLargestKey(e.directions.up)
		return largest > e.currentFloor
	}

	return false
}

// check if elevator has more requests in the down direction and should continue move down
func (e *Elevator) shouldMoveDown() bool {
	if e.directions.isDownExisting() {
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
	currentDirection := e.GetDirection()
	if currentDirection == "" {
		e.SetDirection(direction)
	}

	e.directions.Append(direction, fromFloor, toFloor)
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

func (e *Elevator) GetCurrentFloor() int {
	e.mu.RLock()
	currentFloor := e.currentFloor
	e.mu.RUnlock()
	return currentFloor
}
