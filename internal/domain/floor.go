package domain

import (
	"fmt"

	"github.com/slavakukuyev/elevator-go/internal/constants"
)

// Floor represents a floor number in a building
type Floor int

// NewFloor creates a new Floor with basic validation
func NewFloor(value int) Floor {
	return Floor(value)
}

// NewFloorWithValidation creates a new Floor with strict validation for client input
func NewFloorWithValidation(value int) (Floor, error) {
	if value < constants.MinAllowedFloor || value > constants.MaxAllowedFloor {
		return Floor(0), NewValidationError(
			fmt.Sprintf("floor value %d is outside allowed range [%d, %d]",
				value, constants.MinAllowedFloor, constants.MaxAllowedFloor), nil).
			WithContext("floor", value).
			WithContext("min_allowed", constants.MinAllowedFloor).
			WithContext("max_allowed", constants.MaxAllowedFloor)
	}
	return Floor(value), nil
}

// Value returns the integer value of the floor
func (f Floor) Value() int {
	return int(f)
}

// IsValid checks if the floor is within the given range
func (f Floor) IsValid(minFloor, maxFloor Floor) bool {
	return f >= minFloor && f <= maxFloor
}

// IsValidAbsolute checks if the floor is within absolute system limits
func (f Floor) IsValidAbsolute() bool {
	return int(f) >= constants.MinAllowedFloor && int(f) <= constants.MaxAllowedFloor
}

// Distance calculates the distance between two floors
func (f Floor) Distance(other Floor) int {
	diff := int(f) - int(other)
	if diff < 0 {
		return -diff
	}
	return diff
}

// String returns string representation of the floor
func (f Floor) String() string {
	return fmt.Sprintf("%d", int(f))
}

// IsAbove checks if this floor is above another floor
func (f Floor) IsAbove(other Floor) bool {
	return f > other
}

// IsBelow checks if this floor is below another floor
func (f Floor) IsBelow(other Floor) bool {
	return f < other
}

// IsEqual checks if this floor equals another floor
func (f Floor) IsEqual(other Floor) bool {
	return f == other
}

// ValidateFloorRange validates that from and to floors make sense
func ValidateFloorRange(from, to Floor) error {
	if from == to {
		return NewValidationError("from and to floors cannot be the same", nil).
			WithContext("from_floor", from.Value()).
			WithContext("to_floor", to.Value())
	}

	if !from.IsValidAbsolute() {
		return NewValidationError("from floor is outside valid range", nil).
			WithContext("from_floor", from.Value()).
			WithContext("min_allowed", constants.MinAllowedFloor).
			WithContext("max_allowed", constants.MaxAllowedFloor)
	}

	if !to.IsValidAbsolute() {
		return NewValidationError("to floor is outside valid range", nil).
			WithContext("to_floor", to.Value()).
			WithContext("min_allowed", constants.MinAllowedFloor).
			WithContext("max_allowed", constants.MaxAllowedFloor)
	}

	return nil
}
