package factory

import (
	"time"

	"github.com/slavakukuyev/elevator-go/internal/elevator"
	"github.com/slavakukuyev/elevator-go/internal/infra/config"
)

type ElevatorFactory interface {
	CreateElevator(cfg *config.Config, name string,
		minFloor, maxFloor int,
		eachFloorDuration, openDoorDuration time.Duration, overloadThreshold int) (*elevator.Elevator, error)
}

type StandardElevatorFactory struct{}

func (f StandardElevatorFactory) CreateElevator(cfg *config.Config, name string,
	minFloor, maxFloor int,
	eachFloorDuration, openDoorDuration time.Duration, overloadThreshold int) (*elevator.Elevator, error) {

	return elevator.New(name,
		minFloor, maxFloor,
		eachFloorDuration, openDoorDuration, cfg.OperationTimeout,
		cfg.CircuitBreakerMaxFailures, cfg.CircuitBreakerResetTimeout, cfg.CircuitBreakerHalfOpenLimit, overloadThreshold)
}
