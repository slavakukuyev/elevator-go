package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSetDirection(t *testing.T) {
	elevator := NewElevator("TestElevator", 0, 10)

	// Test setting direction
	direction := _directionUp
	elevator.setDirection(direction)
	if elevator.CurrentDirection() != direction {
		t.Errorf("Expected direction %s, got %s", direction, elevator.CurrentDirection())
	}

	// Test setting direction again
	direction = _directionDown
	elevator.setDirection(direction)
	if elevator.CurrentDirection() != direction {
		t.Errorf("Expected direction %s, got %s", direction, elevator.CurrentDirection())
	}
}

func TestElevator_Run(t *testing.T) {
	// Create a new elevator
	elevator := NewElevator("TestElevator", 0, 10)

	// Add some requests to the elevator
	elevator.Request(_directionUp, 2, 5)
	elevator.Request(_directionDown, 8, 3)

	// Run the elevator
	go elevator.Run()

	// Wait for the elevator to finish running
	time.Sleep(time.Second * 5)

	// Check the current floor and direction of the elevator
	assert.Equal(t, 3, elevator.CurrentFloor())
	assert.Equal(t, _directionDown, elevator.CurrentDirection())
}

func TestElevator_Request(t *testing.T) {

	// Create a new elevator
	elevator := NewElevator("TestElevator", 0, 10, logger)

	// Add a request to the elevator
	elevator.Request(_directionUp, 2, 5)

	// Check if the request is in range
	assert.True(t, elevator.IsRequestInRange(2, 5))
	assert.False(t, elevator.IsRequestInRange(1, 6))
}

func TestElevator_CurrentDirection(t *testing.T) {
	// Create a new elevator and set the current direction
	elevator := NewElevator("TestElevator", 0, 10)

	// Check the initial current direction of the elevator
	assert.Equal(t, "", elevator.CurrentDirection())

	// Set the current direction of the elevator
	elevator.setDirection(_directionUp)

	// Check the updated current direction of the elevator
	assert.Equal(t, _directionUp, elevator.CurrentDirection())
}

func TestElevator_CurrentFloor(t *testing.T) {
	// Create a new elevator
	elevator := NewElevator("TestElevator", 0, 10)

	// Check the initial current floor of the elevator
	assert.Equal(t, 0, elevator.CurrentFloor())

	// Set the current floor of the elevator
	elevator.setCurrentFloor(5)

	// Check the updated current floor of the elevator
	assert.Equal(t, 5, elevator.CurrentFloor())
}

func TestElevator_Directions(t *testing.T) {
	// Create a new elevator
	elevator := NewElevator("TestElevator", 0, 10)

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

func TestElevator_IsRequestInRange(t *testing.T) {
	// Create a new elevator
	elevator := NewElevator("TestElevator", 0, 10)

	// Check if a request is in range
	assert.True(t, elevator.IsRequestInRange(2, 5))
	assert.False(t, elevator.IsRequestInRange(1, 6))
}
