package http

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/slavakukuyev/elevator-go/internal/domain"
)

func TestNewResponseWriter(t *testing.T) {
	w := httptest.NewRecorder()
	logger := slog.Default()
	requestID := "test-123"

	rw := NewResponseWriter(w, logger, requestID)

	assert.NotNil(t, rw)
	assert.Equal(t, w, rw.ResponseWriter)
	assert.Equal(t, logger, rw.logger)
	assert.Equal(t, requestID, rw.requestID)
	assert.WithinDuration(t, time.Now(), rw.startTime, time.Second)
}

func TestResponseWriter_WriteJSON(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		data          interface{}
		wantStatus    int
		checkResponse func(t *testing.T, response APIResponse)
	}{
		{
			name:       "success response with data",
			statusCode: http.StatusOK,
			data:       map[string]string{"message": "success"},
			wantStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response APIResponse) {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
				assert.Nil(t, response.Error)
				assert.NotNil(t, response.Meta)
				assert.Equal(t, "test-123", response.Meta.RequestID)
				assert.Equal(t, "v1", response.Meta.Version)
			},
		},
		{
			name:       "created response",
			statusCode: http.StatusCreated,
			data:       map[string]interface{}{"id": 1, "name": "test"},
			wantStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, response APIResponse) {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Data)
			},
		},
		{
			name:       "client error response",
			statusCode: http.StatusBadRequest,
			data:       nil,
			wantStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response APIResponse) {
				assert.False(t, response.Success)
			},
		},
		{
			name:       "server error response",
			statusCode: http.StatusInternalServerError,
			data:       nil,
			wantStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, response APIResponse) {
				assert.False(t, response.Success)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			logger := slog.Default()
			requestID := "test-123"

			rw := NewResponseWriter(w, logger, requestID)
			rw.WriteJSON(tt.statusCode, tt.data)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.Equal(t, requestID, w.Header().Get("X-Request-ID"))

			var response APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			tt.checkResponse(t, response)
			assert.WithinDuration(t, time.Now(), response.Timestamp, 5*time.Second)
		})
	}
}

func TestResponseWriter_WriteError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		errorCode  string
		message    string
		details    string
		wantStatus int
	}{
		{
			name:       "validation error",
			statusCode: http.StatusBadRequest,
			errorCode:  ErrorCodeValidation,
			message:    "Invalid input",
			details:    "Floor must be between -100 and 200",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not found error",
			statusCode: http.StatusNotFound,
			errorCode:  ErrorCodeNotFound,
			message:    "Resource not found",
			details:    "Elevator not found",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "internal error",
			statusCode: http.StatusInternalServerError,
			errorCode:  ErrorCodeInternal,
			message:    "Internal server error",
			details:    "Database connection failed",
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "rate limit error",
			statusCode: http.StatusTooManyRequests,
			errorCode:  ErrorCodeRateLimit,
			message:    "Rate limit exceeded",
			details:    "Too many requests from IP",
			wantStatus: http.StatusTooManyRequests,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			logger := slog.Default()
			requestID := "test-456"

			rw := NewResponseWriter(w, logger, requestID)
			rw.WriteError(tt.statusCode, tt.errorCode, tt.message, tt.details)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.Equal(t, requestID, w.Header().Get("X-Request-ID"))

			var response APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.False(t, response.Success)
			assert.Nil(t, response.Data)
			assert.NotNil(t, response.Error)
			assert.Equal(t, tt.errorCode, response.Error.Code)
			assert.Equal(t, tt.message, response.Error.Message)
			assert.Equal(t, tt.details, response.Error.Details)
			assert.Equal(t, requestID, response.Error.RequestID)
			assert.NotEmpty(t, response.Error.UserMessage)

			assert.NotNil(t, response.Meta)
			assert.Equal(t, requestID, response.Meta.RequestID)
			assert.Equal(t, "v1", response.Meta.Version)
		})
	}
}

func TestResponseWriter_WriteDomainError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
		expectedMsg    string
	}{
		{
			name:           "validation domain error",
			err:            domain.NewValidationError("invalid floor", nil),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "VALIDATION_ERROR",
			expectedMsg:    "Invalid input provided",
		},
		{
			name:           "not found domain error",
			err:            domain.NewNotFoundError("elevator not found", nil),
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
			expectedMsg:    "Resource not found",
		},
		{
			name:           "conflict domain error",
			err:            domain.NewConflictError("elevator already exists", nil),
			expectedStatus: http.StatusConflict,
			expectedCode:   "CONFLICT",
			expectedMsg:    "Resource conflict",
		},
		{
			name:           "internal domain error",
			err:            domain.NewInternalError("database error", nil),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_ERROR",
			expectedMsg:    "Internal server error",
		},
		{
			name:           "generic error",
			err:            assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_ERROR",
			expectedMsg:    "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			logger := slog.Default()
			requestID := "test-789"

			rw := NewResponseWriter(w, logger, requestID)
			rw.WriteDomainError(tt.err)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.False(t, response.Success)
			assert.NotNil(t, response.Error)
			assert.Equal(t, tt.expectedCode, response.Error.Code)
			assert.Equal(t, tt.expectedMsg, response.Error.Message)
			assert.Equal(t, tt.err.Error(), response.Error.Details)
		})
	}
}

func TestGetUserFriendlyMessage(t *testing.T) {
	tests := []struct {
		errorCode string
		expected  string
	}{
		{
			errorCode: ErrorCodeValidation,
			expected:  "Please check your input and try again.",
		},
		{
			errorCode: ErrorCodeNotFound,
			expected:  "The requested resource was not found.",
		},
		{
			errorCode: ErrorCodeConflict,
			expected:  "The requested operation conflicts with existing data.",
		},
		{
			errorCode: ErrorCodeInternal,
			expected:  "Something went wrong on our end. Please try again later.",
		},
		{
			errorCode: ErrorCodeMethodNotAllowed,
			expected:  "This HTTP method is not supported for this endpoint.",
		},
		{
			errorCode: ErrorCodeInvalidJSON,
			expected:  "The provided JSON is malformed.",
		},
		{
			errorCode: ErrorCodeRateLimit,
			expected:  "Too many requests. Please slow down.",
		},
		{
			errorCode: "UNKNOWN_ERROR",
			expected:  "An error occurred while processing your request.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.errorCode, func(t *testing.T) {
			result := getUserFriendlyMessage(tt.errorCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResponseWriter_JSONEncodingError(t *testing.T) {
	// Test case where JSON encoding might fail
	w := httptest.NewRecorder()

	// Create a logger that captures output
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))

	requestID := "test-encoding"
	rw := NewResponseWriter(w, logger, requestID)

	// Try to encode something that can't be encoded to JSON
	invalidData := make(chan int) // channels can't be JSON encoded
	rw.WriteJSON(http.StatusOK, invalidData)

	// Should fall back to simple error response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal server error")
}

func TestResponseWriter_Headers(t *testing.T) {
	w := httptest.NewRecorder()
	logger := slog.Default()
	requestID := "test-headers"

	rw := NewResponseWriter(w, logger, requestID)
	rw.WriteJSON(http.StatusOK, map[string]string{"test": "data"})

	// Check that proper headers are set
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, requestID, w.Header().Get("X-Request-ID"))
}

func TestResponseWriter_TimingInfo(t *testing.T) {
	w := httptest.NewRecorder()
	logger := slog.Default()
	requestID := "test-timing"

	rw := NewResponseWriter(w, logger, requestID)

	// Add a small delay to test timing
	time.Sleep(10 * time.Millisecond)

	rw.WriteJSON(http.StatusOK, map[string]string{"test": "data"})

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response.Meta)
	assert.NotEmpty(t, response.Meta.Duration)

	// Duration should be parseable and reasonable
	duration, err := time.ParseDuration(response.Meta.Duration)
	require.NoError(t, err)
	assert.True(t, duration >= 10*time.Millisecond)
	assert.True(t, duration < time.Second) // Should be much less than a second
}
