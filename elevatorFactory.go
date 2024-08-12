package main

import (
	"time"

	"go.uber.org/zap"
)

type ElevatorFactory interface {
	CreateElevator(name string,
		minFloor, maxFloor int,
		eachFloorDuration, openDoorDuration time.Duration,
		logger *zap.Logger) (*Elevator, error)
}

type StandardElevatorFactory struct{}

func (f *StandardElevatorFactory) CreateElevator(name string,
	minFloor, maxFloor int,
	eachFloorDuration, openDoorDuration time.Duration,
	logger *zap.Logger) (*Elevator, error) {

	return NewElevator(name,
		minFloor, maxFloor,
		eachFloorDuration, openDoorDuration,
		logger)
}
