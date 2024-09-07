package elevator

import (
	"testing"
	"time"

	"github.com/slavakukuyev/elevator-go/internal/infra/config"
	"github.com/stretchr/testify/assert"
)

const _directionDown = "down"
const _directionUp = "up"

func buildElevatorTestConfig() *config.Config {
	return &config.Config{DirectionUpKey: "up", DirectionDownKey: "down"}
}

func TestElevator_Run(t *testing.T) {
	// Create a new elevator
	elevator, err := New(buildElevatorTestConfig(), "TestElevator", 0, 10, time.Millisecond*100, time.Millisecond*100)
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

func TestElevator_CurrentDirection(t *testing.T) {
	// Create a new elevator
	elevator, err := New(buildElevatorTestConfig(), "TestElevator", 0, 10, time.Millisecond*500, time.Second*2)
	assert.Nil(t, err)

	// Check the initial current direction of the elevator
	assert.Equal(t, "", elevator.CurrentDirection())

	// Set the current direction of the elevator
	elevator.setDirection(_directionUp)

	// Check the updated current direction of the elevator
	assert.Equal(t, _directionUp, elevator.CurrentDirection())
}

func TestElevator_CurrentFloor(t *testing.T) {
	// Create a new elevator
	elevator, err := New(buildElevatorTestConfig(), "TestElevator", 0, 10, time.Millisecond*500, time.Second*2)
	assert.Nil(t, err)

	// Check the initial current floor of the elevator
	assert.Equal(t, 0, elevator.CurrentFloor())

	// Set the current floor of the elevator
	elevator.setCurrentFloor(5)

	// Check the updated current floor of the elevator
	assert.Equal(t, 5, elevator.CurrentFloor())
}

func TestElevatorDirections(t *testing.T) {
	// Create a new elevator
	elevator, err := New(buildElevatorTestConfig(), "TestElevator", 0, 10, time.Millisecond*500, time.Second*2)
	assert.Nil(t, err)

	// Check the initial directions of the elevator
	assert.NotNil(t, elevator.Directions())
	assert.Empty(t, elevator.Directions().Up())
	assert.Empty(t, elevator.Directions().Down())

	// Add some requests to the elevator
	elevator.Request(_directionUp, 2, 5)
	elevator.Request(_directionDown, 8, 3)

	// Check the updated directions of the elevator
	assert.NotEmpty(t, elevator.Directions().Up())
	assert.NotEmpty(t, elevator.Directions().Down())
}
