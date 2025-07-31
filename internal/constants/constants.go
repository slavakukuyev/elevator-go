package constants

import "time"

// Application constants centralized in one location to improve type safety
// and eliminate magic strings throughout the codebase

// Default Configuration Values
const (
	// Server defaults
	DefaultPort     = 6660
	DefaultLogLevel = "INFO"
	DefaultMinFloor = 0
	DefaultMaxFloor = 9

	// Timing defaults
	DefaultEachFloorDuration = 500 * time.Millisecond
	DefaultOpenDoorDuration  = 2 * time.Second

	// WebSocket update interval
	StatusUpdateInterval = 1 * time.Second
)

// HTTP Content Types
const (
	ContentTypeJSON      = "application/json"
	ContentTypeTextPlain = "text/plain"
)

// HTTP Methods
const (
	MethodGET  = "GET"
	MethodPOST = "POST"
)

// Component Names for Logging
const (
	ComponentHTTPServer  = "http-server"
	ComponentHTTPHandler = "http_handler"
	ComponentElevator    = "elevator"
	ComponentManager     = "manager"
	ComponentDirections  = "directions"
)

// Floor Validation Limits
const (
	MinAllowedFloor = -100 // Reasonable minimum for basements
	MaxAllowedFloor = 200  // Reasonable maximum for skyscrapers
)

// Metrics
const (
	MetricsNamespace  = "elevator"
	ElevatorNameLabel = "elevator"
)

// Default Elevator Names
const (
	DefaultElevatorPrefix = "Elevator"
)
