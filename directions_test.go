package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirections_Append(t *testing.T) {
	directions := NewDirections()

	// Append a request in the up direction
	directions.Append(_directionUp, 1, 3)
	directions.Append(_directionUp, 1, 5)

	// Check the updated directions map
	assert.Equal(t, []int{3, 5}, directions.up[1])

	// Append a request in the down direction
	directions.Append(_directionDown, 2, 4)
	directions.Append(_directionDown, 2, 6)

	// Check the updated directions map
	assert.Equal(t, []int{4, 6}, directions.down[2])
}

func TestDirections_Flush(t *testing.T) {
	directions := NewDirections()

	// Append requests in the up direction
	directions.Append(_directionUp, 1, 3)
	directions.Append(_directionUp, 1, 5)

	// Flush the requests from the current floor
	directions.Flush(_directionUp, 1)

	// Check the updated directions map
	assert.Empty(t, directions.up[1])
	assert.Empty(t, directions.up[3])
	assert.Empty(t, directions.up[5])

	// Append requests in the down direction
	directions.Append(_directionDown, 2, 4)
	directions.Append(_directionDown, 2, 6)

	// Flush the requests from the current floor
	directions.Flush(_directionDown, 2)

	// Check the updated directions map
	assert.Empty(t, directions.down[2])
	assert.Empty(t, directions.down[4])
	assert.Empty(t, directions.down[6])
}

func TestDirections_UpDirectionLength(t *testing.T) {
	directions := NewDirections()

	// Append requests in the up direction
	directions.Append(_directionUp, 1, 3)
	directions.Append(_directionUp, 1, 5)

	// Check the length of the up direction
	assert.Equal(t, 1, directions.UpDirectionLength())

	// Flush the requests from the current floor
	directions.Flush(_directionUp, 1)

	// Check the updated length of the up direction, because floors 3 and 5 have been created as map keys
	assert.Equal(t, 2, directions.UpDirectionLength())
}

func TestDirections_DownDirectionLength(t *testing.T) {
	directions := NewDirections()

	// Append requests in the down direction
	directions.Append(_directionDown, 2, 4)
	directions.Append(_directionDown, 2, 6)

	// Check the length of the down direction
	assert.Equal(t, 1, directions.DownDirectionLength())

	// Flush the requests from the current floor
	directions.Flush(_directionDown, 2)
	directions.Flush(_directionDown, 4)
	directions.Flush(_directionDown, 6)

	// Check the updated length of the down direction
	assert.Equal(t, 0, directions.DownDirectionLength())
}

func TestDirections_DirectionsLength(t *testing.T) {
	directions := NewDirections()

	// Append requests in both directions
	directions.Append(_directionUp, 1, 3)
	directions.Append(_directionDown, 2, 4)

	// Check the total length of the directions
	assert.Equal(t, 2, directions.DirectionsLength())

	// Flush the requests from the current floors
	directions.Flush(_directionUp, 1)
	directions.Flush(_directionDown, 2)

	// Check the updated total length of the directions
	assert.Equal(t, 2, directions.DirectionsLength())
}

func TestDirections_IsExisting(t *testing.T) {
	directions := NewDirections()

	// Append requests in both directions
	directions.Append(_directionUp, 1, 3)
	directions.Append(_directionDown, 2, 4)

	// Check if the requests exist
	assert.True(t, directions.IsExisting(_directionUp, 1, 3))
	assert.True(t, directions.IsExisting(_directionDown, 2, 4))

	// Check if non-existing requests return false
	assert.False(t, directions.IsExisting(_directionUp, 1, 4))
	assert.False(t, directions.IsExisting(_directionDown, 2, 3))
}

func TestValidateIntInMapSlice(t *testing.T) {
	m := map[int][]int{
		1: {3, 5},
		2: {4, 6},
	}

	// Check if the values exist in the map slice
	assert.True(t, validateIntInMapSlice(m, 1, 3))
	assert.True(t, validateIntInMapSlice(m, 1, 5))
	assert.True(t, validateIntInMapSlice(m, 2, 4))
	assert.True(t, validateIntInMapSlice(m, 2, 6))

	// Check if non-existing values return false
	assert.False(t, validateIntInMapSlice(m, 1, 4))
	assert.False(t, validateIntInMapSlice(m, 2, 3))
}
