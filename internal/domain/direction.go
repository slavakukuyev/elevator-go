package domain

// Direction represents the movement direction of an elevator
type Direction string

const (
	DirectionUp       Direction = "up"
	DirectionDown     Direction = "down"
	DirectionIdle     Direction = ""
	DirectionDeleting Direction = "deleting"
)

// String returns the string representation of the direction
func (d Direction) String() string {
	return string(d)
}

// IsValid checks if the direction is valid
func (d Direction) IsValid() bool {
	return d == DirectionUp || d == DirectionDown || d == DirectionIdle || d == DirectionDeleting
}

// Opposite returns the opposite direction
func (d Direction) Opposite() Direction {
	switch d {
	case DirectionUp:
		return DirectionDown
	case DirectionDown:
		return DirectionUp
	default:
		return DirectionIdle
	}
}

// IsOperational returns true if the elevator can accept new requests
func (d Direction) IsOperational() bool {
	return d != DirectionDeleting
}
