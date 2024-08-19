package factory

import (
	"time"

	"github.com/slavakukuyev/elevator-go/internal/elevator"
	"github.com/slavakukuyev/elevator-go/internal/infra/config"
)

type ElevatorFactory interface {
	CreateElevator(cfg *config.Config, name string,
		minFloor, maxFloor int,
		eachFloorDuration, openDoorDuration time.Duration) (*elevator.Elevator, error)
}

type StandardElevatorFactory struct{}

func (f StandardElevatorFactory) CreateElevator(cfg *config.Config, name string,
	minFloor, maxFloor int,
	eachFloorDuration, openDoorDuration time.Duration) (*elevator.Elevator, error) {

	return elevator.New(cfg, name,
		minFloor, maxFloor,
		eachFloorDuration, openDoorDuration)
}
