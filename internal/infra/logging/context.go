package logging

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"
)

// ContextKey represents a context key type
type ContextKey string

const (
	// CorrelationIDKey is the context key for correlation ID
	CorrelationIDKey ContextKey = "correlation_id"
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
)

// GenerateCorrelationID generates a new correlation ID
func GenerateCorrelationID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("corr_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", bytes)
}

// WithCorrelationID adds a correlation ID to the context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// GetCorrelationID retrieves the correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if correlationID := ctx.Value(CorrelationIDKey); correlationID != nil {
		return correlationID.(string)
	}
	return ""
}

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		return requestID.(string)
	}
	return ""
}

// NewContextWithCorrelation creates a new context with a correlation ID
func NewContextWithCorrelation(ctx context.Context) context.Context {
	correlationID := GenerateCorrelationID()
	return WithCorrelationID(ctx, correlationID)
}
