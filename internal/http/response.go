package http

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/slavakukuyev/elevator-go/internal/constants"
	"github.com/slavakukuyev/elevator-go/internal/domain"
)

// APIResponse represents the standard API response structure
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Meta      *APIMeta    `json:"meta,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// APIError represents error information in API responses
type APIError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Details     string `json:"details,omitempty"`
	RequestID   string `json:"request_id,omitempty"`
	UserMessage string `json:"user_message,omitempty"`
}

// APIMeta represents metadata in API responses
type APIMeta struct {
	RequestID string `json:"request_id,omitempty"`
	Version   string `json:"version,omitempty"`
	Duration  string `json:"duration,omitempty"`
}

// ResponseWriter wraps http.ResponseWriter with additional functionality
type ResponseWriter struct {
	http.ResponseWriter
	logger    *slog.Logger
	requestID string
	startTime time.Time
}

// Header returns the header map that will be sent by WriteHeader
func (rw *ResponseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

// Write writes the data to the connection as part of an HTTP reply
func (rw *ResponseWriter) Write(data []byte) (int, error) {
	return rw.ResponseWriter.Write(data)
}

// WriteHeader sends an HTTP response header with the provided status code
func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Hijack implements http.Hijacker interface for WebSocket support
func (rw *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("ResponseWriter does not implement http.Hijacker")
}

// NewResponseWriter creates a new ResponseWriter
func NewResponseWriter(w http.ResponseWriter, logger *slog.Logger, requestID string) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		logger:         logger,
		requestID:      requestID,
		startTime:      time.Now(),
	}
}

// WriteJSON writes a JSON response with the standard API format
func (rw *ResponseWriter) WriteJSON(statusCode int, data interface{}) {
	response := APIResponse{
		Success:   statusCode >= 200 && statusCode < 300,
		Data:      data,
		Timestamp: time.Now(),
		Meta: &APIMeta{
			RequestID: rw.requestID,
			Version:   "v1",
			Duration:  time.Since(rw.startTime).String(),
		},
	}

	rw.Header().Set("Content-Type", constants.ContentTypeJSON)
	rw.Header().Set("X-Request-ID", rw.requestID)

	// Try to encode first, then write headers if successful
	encoded, err := json.Marshal(response)
	if err != nil {
		rw.logger.Error("failed to encode JSON response",
			slog.String("error", err.Error()),
			slog.String("request_id", rw.requestID))
		// Write error response instead
		rw.WriteHeader(http.StatusInternalServerError)
		if _, writeErr := rw.Write([]byte(`{"success":false,"error":{"code":"INTERNAL_ERROR","message":"Internal server error"},"timestamp":"` + time.Now().Format(time.RFC3339) + `"}`)); writeErr != nil {
			rw.logger.Error("failed to write error response",
				slog.String("error", writeErr.Error()),
				slog.String("request_id", rw.requestID))
		}
		return
	}

	rw.WriteHeader(statusCode)
	if _, writeErr := rw.Write(encoded); writeErr != nil {
		rw.logger.Error("failed to write JSON response",
			slog.String("error", writeErr.Error()),
			slog.String("request_id", rw.requestID))
	}
}

// WriteError writes a JSON error response with the standard API format
func (rw *ResponseWriter) WriteError(statusCode int, errorCode, message, details string) {
	apiError := &APIError{
		Code:        errorCode,
		Message:     message,
		Details:     details,
		RequestID:   rw.requestID,
		UserMessage: getUserFriendlyMessage(errorCode),
	}

	response := APIResponse{
		Success:   false,
		Error:     apiError,
		Timestamp: time.Now(),
		Meta: &APIMeta{
			RequestID: rw.requestID,
			Version:   "v1",
			Duration:  time.Since(rw.startTime).String(),
		},
	}

	rw.Header().Set("Content-Type", constants.ContentTypeJSON)
	rw.Header().Set("X-Request-ID", rw.requestID)
	rw.WriteHeader(statusCode)

	if err := json.NewEncoder(rw).Encode(response); err != nil {
		rw.logger.Error("failed to encode error response",
			slog.String("error", err.Error()),
			slog.String("request_id", rw.requestID))
	}
}

// WriteDomainError writes a domain error as a JSON response
func (rw *ResponseWriter) WriteDomainError(err error) {
	statusCode := http.StatusInternalServerError
	errorCode := "INTERNAL_ERROR"
	message := "Internal server error"
	details := ""

	if domainErr, ok := err.(*domain.DomainError); ok {
		switch domainErr.Type {
		case domain.ErrTypeValidation:
			statusCode = http.StatusBadRequest
			errorCode = "VALIDATION_ERROR"
			message = "Invalid input provided"
		case domain.ErrTypeNotFound:
			statusCode = http.StatusNotFound
			errorCode = "NOT_FOUND"
			message = "Resource not found"
		case domain.ErrTypeConflict:
			statusCode = http.StatusConflict
			errorCode = "CONFLICT"
			message = "Resource conflict"
		case domain.ErrTypeInternal:
			statusCode = http.StatusInternalServerError
			errorCode = "INTERNAL_ERROR"
			message = "Internal server error"
		}
		details = domainErr.Error()
	} else {
		details = err.Error()
	}

	rw.WriteError(statusCode, errorCode, message, details)
}

// getUserFriendlyMessage returns user-friendly messages for error codes
func getUserFriendlyMessage(errorCode string) string {
	messages := map[string]string{
		"VALIDATION_ERROR":   "Please check your input and try again.",
		"NOT_FOUND":          "The requested resource was not found.",
		"CONFLICT":           "The requested operation conflicts with existing data.",
		"INTERNAL_ERROR":     "Something went wrong on our end. Please try again later.",
		"METHOD_NOT_ALLOWED": "This HTTP method is not supported for this endpoint.",
		"INVALID_JSON":       "The provided JSON is malformed.",
		"RATE_LIMITED":       "Too many requests. Please slow down.",
	}

	if msg, exists := messages[errorCode]; exists {
		return msg
	}
	return "An error occurred while processing your request."
}

// ErrorCode constants for consistent error handling
const (
	ErrorCodeValidation       = "VALIDATION_ERROR"
	ErrorCodeNotFound         = "NOT_FOUND"
	ErrorCodeConflict         = "CONFLICT"
	ErrorCodeInternal         = "INTERNAL_ERROR"
	ErrorCodeMethodNotAllowed = "METHOD_NOT_ALLOWED"
	ErrorCodeInvalidJSON      = "INVALID_JSON"
	ErrorCodeRateLimit        = "RATE_LIMITED"
)
