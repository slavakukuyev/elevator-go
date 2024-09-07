package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/slavakukuyev/elevator-go/internal/factory"
	"github.com/slavakukuyev/elevator-go/internal/infra/config"
	"github.com/slavakukuyev/elevator-go/internal/manager"
	"github.com/stretchr/testify/assert"
)

func buildServerTestConfig() *config.Config {
	return &config.Config{DirectionUpKey: "up", DirectionDownKey: "down"}
}

func TestElevatorHandler(t *testing.T) {
	factory := &factory.StandardElevatorFactory{}
	// Create a new Manager and Server instance
	manager := manager.New(buildServerTestConfig(), factory)
	server := NewServer(buildServerTestConfig(), 8080, manager)

	// Create a new HTTP request
	requestBody := ElevatorRequestBody{
		Name:     "Elevator1",
		MinFloor: 0,
		MaxFloor: 9,
	}
	requestBodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/elevator", bytes.NewBuffer(requestBodyBytes))

	// Create a new HTTP response recorder
	rr := httptest.NewRecorder()

	// Call the elevatorHandler method
	handler := http.HandlerFunc(server.elevatorHandler)
	handler.ServeHTTP(rr, req)

	// Check the response status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response body
	expectedResponse := "elevator Elevator1 has been created successfully"
	assert.Equal(t, expectedResponse, rr.Body.String())
}

func TestFloorHandler(t *testing.T) {
	factory := &factory.StandardElevatorFactory{}
	// Create a new Manager and Server instance
	manager := manager.New(buildServerTestConfig(), factory)
	err := manager.AddElevator(buildServerTestConfig(), "Elevator1", 0, 9, 1*time.Second, 1*time.Second)
	assert.Nil(t, err)

	server := NewServer(buildServerTestConfig(), 8080, manager)

	// Create a new HTTP request
	requestBody := FloorRequestBody{
		From: 0,
		To:   9,
	}
	requestBodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/floor", bytes.NewBuffer(requestBodyBytes))

	// Create a new HTTP response recorder
	rr := httptest.NewRecorder()

	// Call the floorHandler method
	handler := http.HandlerFunc(server.floorHandler)
	handler.ServeHTTP(rr, req)

	// Check the response status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response body
	expectedResponse := "elevator Elevator1 received request: from 0 to 9"
	assert.Equal(t, expectedResponse, rr.Body.String())
}
