package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestElevator_Run(t *testing.T) {
	logger := zap.NewNop()

	// Create a new elevator
	elevator, err := NewElevator("TestElevator", 0, 10, time.Millisecond*100, time.Millisecond*100, logger)
	assert.Nil(t, err)

	// Add some requests to the elevator
	elevator.Request(_directionUp, 2, 5)
	elevator.Request(_directionDown, 8, 3)

	// Wait for the elevator to finish running
	time.Sleep(time.Millisecond * 800)

	// Check the current floor and direction of the elevator
	assert.Equal(t, 5, elevator.CurrentFloor())
	assert.Equal(t, _directionUp, elevator.CurrentDirection())

	time.Sleep(time.Millisecond * 1500)

	assert.Equal(t, 3, elevator.CurrentFloor())
	assert.Equal(t, "", elevator.CurrentDirection())
}

func TestElevator_IsRequestInRange(t *testing.T) {
	logger := zap.NewNop()

	// Create a new elevator
	elevator, err := NewElevator("TestElevator", 0, 5, time.Millisecond*500, time.Second*2, logger)
	assert.Nil(t, err)

	// Check if the request is in range
	assert.True(t, elevator.IsRequestInRange(0, 5))
	assert.False(t, elevator.IsRequestInRange(-1, 5))
	assert.False(t, elevator.IsRequestInRange(0, 6))
	assert.False(t, elevator.IsRequestInRange(-1, 6))
}

func TestElevator_CurrentDirection(t *testing.T) {
	logger := zap.NewNop()

	// Create a new elevator
	elevator, err := NewElevator("TestElevator", 0, 10, time.Millisecond*500, time.Second*2, logger)
	assert.Nil(t, err)

	// Check the initial current direction of the elevator
	assert.Equal(t, "", elevator.CurrentDirection())

	// Set the current direction of the elevator
	elevator.setDirection(_directionUp)

	// Check the updated current direction of the elevator
	assert.Equal(t, _directionUp, elevator.CurrentDirection())
}

func TestElevator_CurrentFloor(t *testing.T) {
	logger := zap.NewNop()

	// Create a new elevator
	elevator, err := NewElevator("TestElevator", 0, 10, time.Millisecond*500, time.Second*2, logger)
	assert.Nil(t, err)

	// Check the initial current floor of the elevator
	assert.Equal(t, 0, elevator.CurrentFloor())

	// Set the current floor of the elevator
	elevator.setCurrentFloor(5)

	// Check the updated current floor of the elevator
	assert.Equal(t, 5, elevator.CurrentFloor())
}

func TestElevator_Directions(t *testing.T) {
	logger := zap.NewNop()

	// Create a new elevator
	elevator, err := NewElevator("TestElevator", 0, 10, time.Millisecond*500, time.Second*2, logger)
	assert.Nil(t, err)

	// Check the initial directions of the elevator
	assert.NotNil(t, elevator.Directions())
	assert.Empty(t, elevator.Directions().up)
	assert.Empty(t, elevator.Directions().down)

	// Add some requests to the elevator
	elevator.Request(_directionUp, 2, 5)
	elevator.Request(_directionDown, 8, 3)

	// Check the updated directions of the elevator
	assert.NotEmpty(t, elevator.Directions().up)
	assert.NotEmpty(t, elevator.Directions().down)
}
