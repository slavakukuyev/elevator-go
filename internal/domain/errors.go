package domain

import (
	"fmt"
)

// ErrType represents different categories of errors in the system
type ErrType string

const (
	// ErrTypeValidation represents validation errors
	ErrTypeValidation ErrType = "validation"
	// ErrTypeNotFound represents resource not found errors
	ErrTypeNotFound ErrType = "not_found"
	// ErrTypeConflict represents conflict errors
	ErrTypeConflict ErrType = "conflict"
	// ErrTypeInternal represents internal system errors
	ErrTypeInternal ErrType = "internal"
	// ErrTypeExternal represents external service errors
	ErrTypeExternal ErrType = "external"
)

// DomainError represents a structured error with type and context
type DomainError struct {
	Type    ErrType
	Message string
	Err     error
	Context map[string]interface{}
}

// Error implements the error interface
func (de *DomainError) Error() string {
	if de.Err != nil {
		return fmt.Sprintf("%s: %s: %v", de.Type, de.Message, de.Err)
	}
	return fmt.Sprintf("%s: %s", de.Type, de.Message)
}

// Unwrap returns the wrapped error
func (de *DomainError) Unwrap() error {
	return de.Err
}

// NewValidationError creates a new validation error
func NewValidationError(message string, err error) *DomainError {
	return &DomainError{
		Type:    ErrTypeValidation,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string, err error) *DomainError {
	return &DomainError{
		Type:    ErrTypeNotFound,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// NewConflictError creates a new conflict error
func NewConflictError(message string, err error) *DomainError {
	return &DomainError{
		Type:    ErrTypeConflict,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string, err error) *DomainError {
	return &DomainError{
		Type:    ErrTypeInternal,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// NewExternalError creates a new external error
func NewExternalError(message string, err error) *DomainError {
	return &DomainError{
		Type:    ErrTypeExternal,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// WithContext adds context to the error
func (de *DomainError) WithContext(key string, value interface{}) *DomainError {
	de.Context[key] = value
	return de
}

// Common elevator domain errors
var (
	ErrElevatorNameEmpty  = NewValidationError("elevator name cannot be empty", nil)
	ErrElevatorFloorsSame = NewValidationError("minFloor and maxFloor cannot be equal", nil)
	ErrFloorsSame         = NewValidationError("requested floor should be different from current floor", nil)
	ErrFloorsOutOfRange   = NewValidationError("requested floors should be in range of existing elevators", nil)
	ErrNoElevatorFound    = NewNotFoundError("no suitable elevator found for request", nil)
	ErrElevatorCreation   = NewInternalError("failed to create elevator", nil)
	ErrElevatorNotInRange = NewValidationError("elevator does not serve requested floor range", nil)
)
