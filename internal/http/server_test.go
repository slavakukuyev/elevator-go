package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestElevatorHandler(t *testing.T) {
	factory := &StandardElevatorFactory{}
	// Create a new Manager and Server instance
	manager := NewManager(factory, zap.NewNop())
	server := NewServer(8080, manager, zap.NewNop())

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
	factory := &StandardElevatorFactory{}
	// Create a new Manager and Server instance
	manager := NewManager(factory, zap.NewNop())
	err := manager.AddElevator("Elevator1", 0, 9, 1*time.Second, 1*time.Second, zap.NewNop())
	assert.Nil(t, err)

	server := NewServer(8080, manager, zap.NewNop())

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
