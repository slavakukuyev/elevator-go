package domain

import (
	"errors"
	"testing"
)

func TestDomainError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *DomainError
		expected string
	}{
		{
			name: "validation error without wrapped error",
			err: &DomainError{
				Type:    ErrTypeValidation,
				Message: "invalid input",
			},
			expected: "validation: invalid input",
		},
		{
			name: "validation error with wrapped error",
			err: &DomainError{
				Type:    ErrTypeValidation,
				Message: "invalid input",
				Err:     errors.New("underlying error"),
			},
			expected: "validation: invalid input: underlying error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("DomainError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDomainError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := &DomainError{
		Type:    ErrTypeInternal,
		Message: "internal error",
		Err:     underlying,
	}

	if got := err.Unwrap(); got != underlying {
		t.Errorf("DomainError.Unwrap() = %v, want %v", got, underlying)
	}
}

func TestDomainError_WithContext(t *testing.T) {
	err := NewValidationError("test error", nil)
	err = err.WithContext("key1", "value1").WithContext("key2", 42)

	if len(err.Context) != 2 {
		t.Errorf("Expected 2 context entries, got %d", len(err.Context))
	}

	if err.Context["key1"] != "value1" {
		t.Errorf("Expected key1=value1, got %v", err.Context["key1"])
	}

	if err.Context["key2"] != 42 {
		t.Errorf("Expected key2=42, got %v", err.Context["key2"])
	}
}

func TestNewErrorFunctions(t *testing.T) {
	tests := []struct {
		name       string
		errFunc    func(string, error) *DomainError
		errType    ErrType
		message    string
		wrappedErr error
	}{
		{
			name:       "NewValidationError",
			errFunc:    NewValidationError,
			errType:    ErrTypeValidation,
			message:    "validation failed",
			wrappedErr: errors.New("wrapped"),
		},
		{
			name:       "NewNotFoundError",
			errFunc:    NewNotFoundError,
			errType:    ErrTypeNotFound,
			message:    "not found",
			wrappedErr: nil,
		},
		{
			name:       "NewInternalError",
			errFunc:    NewInternalError,
			errType:    ErrTypeInternal,
			message:    "internal error",
			wrappedErr: errors.New("wrapped internal"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errFunc(tt.message, tt.wrappedErr)

			if err.Type != tt.errType {
				t.Errorf("Expected type %v, got %v", tt.errType, err.Type)
			}

			if err.Message != tt.message {
				t.Errorf("Expected message %v, got %v", tt.message, err.Message)
			}

			if err.Err != tt.wrappedErr {
				t.Errorf("Expected wrapped error %v, got %v", tt.wrappedErr, err.Err)
			}

			if err.Context == nil {
				t.Error("Expected Context to be initialized")
			}
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name    string
		err     *DomainError
		errType ErrType
	}{
		{"ErrElevatorNameEmpty", ErrElevatorNameEmpty, ErrTypeValidation},
		{"ErrElevatorFloorsSame", ErrElevatorFloorsSame, ErrTypeValidation},
		{"ErrFloorsSame", ErrFloorsSame, ErrTypeValidation},
		{"ErrFloorsOutOfRange", ErrFloorsOutOfRange, ErrTypeValidation},
		{"ErrNoElevatorFound", ErrNoElevatorFound, ErrTypeNotFound},
		{"ErrElevatorCreation", ErrElevatorCreation, ErrTypeInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Type != tt.errType {
				t.Errorf("Expected %s to have type %v, got %v", tt.name, tt.errType, tt.err.Type)
			}

			if tt.err.Context == nil {
				t.Errorf("Expected %s to have initialized Context", tt.name)
			}
		})
	}
}
