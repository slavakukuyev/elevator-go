package domain

// ElevatorStatus represents the current status of an elevator
type ElevatorStatus struct {
	Name         string    `json:"name"`
	CurrentFloor Floor     `json:"current_floor"`
	Direction    Direction `json:"direction"`
	Requests     int       `json:"requests"`
	MinFloor     Floor     `json:"min_floor"`
	MaxFloor     Floor     `json:"max_floor"`
	IsDeleting   bool      `json:"is_deleting"`
}

// NewElevatorStatus creates a new elevator status
func NewElevatorStatus(name string, currentFloor Floor, direction Direction, requests int, minFloor, maxFloor Floor) ElevatorStatus {
	return ElevatorStatus{
		Name:         name,
		CurrentFloor: currentFloor,
		Direction:    direction,
		Requests:     requests,
		MinFloor:     minFloor,
		MaxFloor:     maxFloor,
		IsDeleting:   direction == DirectionDeleting,
	}
}

// IsIdle returns true if the elevator is idle (no direction)
func (es ElevatorStatus) IsIdle() bool {
	return es.Direction == DirectionIdle
}

// IsMoving returns true if the elevator is moving
func (es ElevatorStatus) IsMoving() bool {
	return !es.IsIdle()
}

// IsAtTopFloor returns true if elevator is at the top floor
func (es ElevatorStatus) IsAtTopFloor() bool {
	return es.CurrentFloor.IsEqual(es.MaxFloor)
}

// IsAtBottomFloor returns true if elevator is at the bottom floor
func (es ElevatorStatus) IsAtBottomFloor() bool {
	return es.CurrentFloor.IsEqual(es.MinFloor)
}

// CanServeFloor returns true if the elevator can serve the given floor
func (es ElevatorStatus) CanServeFloor(floor Floor) bool {
	return floor.IsValid(es.MinFloor, es.MaxFloor)
}

// CanAcceptNewRequests returns true if the elevator can accept new requests
func (es ElevatorStatus) CanAcceptNewRequests() bool {
	return !es.IsDeleting && es.Direction.IsOperational()
}

// IsDeleting returns true if the elevator is being deleted
func (es ElevatorStatus) IsBeingDeleted() bool {
	return es.IsDeleting
}
