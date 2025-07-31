package directions

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/slavakukuyev/elevator-go/internal/domain"
)

func TestDirections_Append(t *testing.T) {
	directions := New()

	// Append a request in the up direction
	directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(3))
	directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(5))

	// Check the updated directions map
	assert.Equal(t, []int{3, 5}, directions.up[1])

	// Append a request in the down direction
	directions.Append(domain.DirectionDown, domain.NewFloor(2), domain.NewFloor(4))
	directions.Append(domain.DirectionDown, domain.NewFloor(2), domain.NewFloor(6))

	// Check the updated directions map
	assert.Equal(t, []int{4, 6}, directions.down[2])
}

func TestDirections_Flush(t *testing.T) {
	directions := New()

	// Append requests in the up direction
	directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(3))
	directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(5))

	// Flush the requests from the current floor
	directions.Flush(domain.DirectionUp, domain.NewFloor(1))

	// Check the updated directions map
	assert.Empty(t, directions.up[1])
	assert.Empty(t, directions.up[3])
	assert.Empty(t, directions.up[5])

	// Append requests in the down direction
	directions.Append(domain.DirectionDown, domain.NewFloor(2), domain.NewFloor(4))
	directions.Append(domain.DirectionDown, domain.NewFloor(2), domain.NewFloor(6))

	// Flush the requests from the current floor
	directions.Flush(domain.DirectionDown, domain.NewFloor(2))

	// Check the updated directions map
	assert.Empty(t, directions.down[2])
	assert.Empty(t, directions.down[4])
	assert.Empty(t, directions.down[6])
}

func TestDirections_UpDirectionLength(t *testing.T) {
	directions := New()

	// Append requests in the up direction
	directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(3))
	directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(5))

	// Check the length: map[1] = [3, 5] → 1 floor + 2 requests = 3
	assert.Equal(t, 3, directions.UpDirectionLength())

	// Flush the requests from the current floor
	directions.Flush(domain.DirectionUp, domain.NewFloor(1))

	// Check the updated length: map[3] = [], map[5] = [] → 2 floors + 0 requests = 2
	assert.Equal(t, 2, directions.UpDirectionLength())
}

func TestDirections_DownDirectionLength(t *testing.T) {
	directions := New()

	// Append requests in the down direction
	directions.Append(domain.DirectionDown, domain.NewFloor(2), domain.NewFloor(4))
	directions.Append(domain.DirectionDown, domain.NewFloor(2), domain.NewFloor(6))

	// Check the length: map[2] = [4, 6] → 1 floor + 2 requests = 3
	assert.Equal(t, 3, directions.DownDirectionLength())

	// Flush the requests from the current floor
	directions.Flush(domain.DirectionDown, domain.NewFloor(2))
	directions.Flush(domain.DirectionDown, domain.NewFloor(4))
	directions.Flush(domain.DirectionDown, domain.NewFloor(6))

	// Check the updated length after all destinations completed
	assert.Equal(t, 0, directions.DownDirectionLength())
}

func TestDirections_DirectionsLength(t *testing.T) {
	directions := New()

	// Append requests in both directions
	directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(3))
	directions.Append(domain.DirectionDown, domain.NewFloor(2), domain.NewFloor(4))

	// Check the total length: up[1]=[3] + down[2]=[4] → 2 + 2 = 4
	assert.Equal(t, 4, directions.DirectionsLength())

	// Flush the requests from the current floors
	directions.Flush(domain.DirectionUp, domain.NewFloor(1))
	directions.Flush(domain.DirectionDown, domain.NewFloor(2))

	// Check the updated total length: up[3]=[] + down[4]=[] → 1 + 1 = 2
	assert.Equal(t, 2, directions.DirectionsLength())

	// Complete all destinations
	directions.Flush(domain.DirectionUp, domain.NewFloor(3))
	directions.Flush(domain.DirectionDown, domain.NewFloor(4))

	// Now should be 0
	assert.Equal(t, 0, directions.DirectionsLength())
}

func TestDirections_IsExisting(t *testing.T) {
	directions := New()

	// Append requests in both directions
	directions.Append(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(3))
	directions.Append(domain.DirectionDown, domain.NewFloor(2), domain.NewFloor(4))

	// Check if the requests exist
	assert.True(t, directions.IsRequestExisting(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(3)))
	assert.True(t, directions.IsRequestExisting(domain.DirectionDown, domain.NewFloor(2), domain.NewFloor(4)))

	// Check if non-existing requests return false
	assert.False(t, directions.IsRequestExisting(domain.DirectionUp, domain.NewFloor(1), domain.NewFloor(4)))
	assert.False(t, directions.IsRequestExisting(domain.DirectionDown, domain.NewFloor(2), domain.NewFloor(3)))
}

func TestValidateIntInMapSlice(t *testing.T) {
	m := map[int][]int{
		1: {3, 5},
		2: {4, 6},
	}

	// Check if the values exist in the map slice
	assert.True(t, isValueInMapSlice(m, 1, 3))
	assert.True(t, isValueInMapSlice(m, 1, 5))
	assert.True(t, isValueInMapSlice(m, 2, 4))
	assert.True(t, isValueInMapSlice(m, 2, 6))

	// Check if non-existing values return false
	assert.False(t, isValueInMapSlice(m, 1, 4))
	assert.False(t, isValueInMapSlice(m, 2, 3))
}
