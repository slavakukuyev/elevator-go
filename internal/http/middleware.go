package http

import (
	"bufio"
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/slavakukuyev/elevator-go/internal/constants"
	"github.com/slavakukuyev/elevator-go/internal/infra/logging"
	"github.com/slavakukuyev/elevator-go/metrics"
)

// Middleware represents a middleware function
type Middleware func(http.Handler) http.Handler

// ChainMiddleware chains multiple middleware functions
func ChainMiddleware(middlewares ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = logging.GenerateCorrelationID()
			}

			ctx := logging.WithRequestID(r.Context(), requestID)
			ctx = logging.WithCorrelationID(ctx, requestID)

			w.Header().Set("X-Request-ID", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// LoggingMiddleware logs HTTP requests with structured logging
func LoggingMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			requestID := logging.GetRequestID(r.Context())
			correlationID := logging.GetCorrelationID(r.Context())

			// Wrap the response writer to capture status code and response size
			wrapper := &responseWriterWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				bytesWritten:   0,
			}

			// Log request start
			logger.InfoContext(r.Context(), "HTTP request started",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", requestID),
				slog.String("correlation_id", correlationID),
				slog.Int64("content_length", r.ContentLength),
				slog.String("component", constants.ComponentHTTPServer))

			next.ServeHTTP(wrapper, r)

			// Calculate metrics
			duration := time.Since(startTime)
			endpoint := sanitizeEndpoint(r.URL.Path)
			statusCode := strconv.Itoa(wrapper.statusCode)

			// Record HTTP metrics
			metrics.RecordHTTPRequest(r.Method, endpoint, statusCode, duration.Seconds())

			// Record performance metrics
			if duration.Seconds() > 1.0 { // Log slow requests
				logger.WarnContext(r.Context(), "slow request detected",
					slog.String("method", r.Method),
					slog.String("endpoint", endpoint),
					slog.String("request_id", requestID),
					slog.Float64("duration_seconds", duration.Seconds()),
					slog.Int("status_code", wrapper.statusCode),
					slog.Int64("response_bytes", wrapper.bytesWritten))
			}

			// Update system performance metrics
			if endpoint == "/v1/floors/request" || endpoint == "/floor" {
				metrics.SetAvgResponseTime("elevator_request", duration.Seconds())
			} else if endpoint == "/v1/health" || endpoint == "/health" {
				metrics.SetAvgResponseTime("health_check", duration.Seconds())
			}

			// Track error rates
			if wrapper.statusCode >= 400 {
				errorType := "client_error"
				if wrapper.statusCode >= 500 {
					errorType = "server_error"
				}
				metrics.IncError(errorType, "http_handler")
			}

			// Log request completion
			logLevel := slog.LevelInfo
			if wrapper.statusCode >= 500 {
				logLevel = slog.LevelError
			} else if wrapper.statusCode >= 400 {
				logLevel = slog.LevelWarn
			}

			logger.Log(r.Context(), logLevel, "HTTP request completed",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status_code", wrapper.statusCode),
				slog.Float64("duration_seconds", duration.Seconds()),
				slog.Int64("response_bytes", wrapper.bytesWritten),
				slog.String("request_id", requestID),
				slog.String("correlation_id", correlationID),
				slog.String("component", constants.ComponentHTTPServer))
		})
	}
}

// RecoveryMiddleware handles panics and returns a proper error response
func RecoveryMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					requestID := logging.GetRequestID(r.Context())
					correlationID := logging.GetCorrelationID(r.Context())

					// Convert panic value to string for logging
					var errorMsg string
					if e, ok := err.(error); ok {
						errorMsg = e.Error()
					} else {
						errorMsg = fmt.Sprintf("%v", err)
					}

					// Capture stack trace
					stack := make([]byte, 4096)
					length := runtime.Stack(stack, false)

					logger.ErrorContext(r.Context(), "HTTP handler panic recovered",
						slog.String("error", errorMsg),
						slog.String("request_id", requestID),
						slog.String("correlation_id", correlationID),
						slog.String("path", r.URL.Path),
						slog.String("method", r.Method),
						slog.String("stack_trace", string(stack[:length])),
						slog.String("component", constants.ComponentHTTPServer))

					// Record panic as error metric
					metrics.IncError("panic", "http_handler")

					rw := NewResponseWriter(w, logger, requestID)
					rw.WriteError(http.StatusInternalServerError, ErrorCodeInternal,
						"Internal server error", "An unexpected error occurred")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
			w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware implements simple in-memory rate limiting
type RateLimitMiddleware struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
	logger   *slog.Logger
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(requestsPerMinute int, logger *slog.Logger) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		requests: make(map[string][]time.Time),
		limit:    requestsPerMinute,
		window:   time.Minute,
		logger:   logger,
	}
}

// Handler returns the middleware handler function
func (rl *RateLimitMiddleware) Handler() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)
			if !rl.isAllowed(clientIP) {
				requestID := getRequestID(r)
				rl.logger.WarnContext(r.Context(), "Rate limit exceeded",
					slog.String("client_ip", clientIP),
					slog.String("request_id", requestID),
					slog.String("component", constants.ComponentHTTPServer))

				rw := NewResponseWriter(w, rl.logger, requestID)
				rw.WriteError(http.StatusTooManyRequests, ErrorCodeRateLimit,
					"Rate limit exceeded", "Too many requests from this IP address")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isAllowed checks if a request is allowed based on rate limits
func (rl *RateLimitMiddleware) isAllowed(clientIP string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	requests := rl.requests[clientIP]

	// Remove requests outside the time window
	var validRequests []time.Time
	for _, requestTime := range requests {
		if now.Sub(requestTime) < rl.window {
			validRequests = append(validRequests, requestTime)
		}
	}

	// Check if we're under the limit
	if len(validRequests) >= rl.limit {
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[clientIP] = validRequests

	return true
}

// SecurityHeadersMiddleware adds common security headers
func SecurityHeadersMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")

			next.ServeHTTP(w, r)
		})
	}
}

// MetricsMiddleware updates system metrics
func MetricsMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Update system resource metrics
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			metrics.SetMemoryUsage("alloc", float64(m.Alloc))
			metrics.SetMemoryUsage("sys", float64(m.Sys))
			metrics.SetMemoryUsage("heap_objects", float64(m.HeapObjects))

			next.ServeHTTP(w, r)

			// Record system performance
			duration := time.Since(start)

			// Track system-wide response times
			if strings.HasPrefix(r.URL.Path, "/v1/") || strings.HasPrefix(r.URL.Path, "/health") || strings.HasPrefix(r.URL.Path, "/metrics") {
				metrics.SetAvgResponseTime("system", duration.Seconds())
			}
		})
	}
}

// Helper functions

// statusResponseWriter wraps http.ResponseWriter to capture status code
type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Hijack implements http.Hijacker interface for WebSocket support
func (w *statusResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("ResponseWriter does not implement http.Hijacker")
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// Use current time + random component to ensure uniqueness
	timestamp := time.Now().UnixNano()
	randomComponent := rand.Int63()
	return strconv.FormatInt(timestamp^randomComponent, 36)
}

// getRequestID extracts request ID from context or request
func getRequestID(r *http.Request) string {
	if requestID := r.Context().Value("request_id"); requestID != nil {
		return requestID.(string)
	}
	return generateRequestID()
}

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for reverse proxies)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header (for some reverse proxies)
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to remote address
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

// responseWriterWrapper wraps http.ResponseWriter to capture response details
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterWrapper) Write(data []byte) (int, error) {
	w.bytesWritten += int64(len(data))
	return w.ResponseWriter.Write(data)
}

// Hijack implements http.Hijacker interface for WebSocket support
func (w *responseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("ResponseWriter does not implement http.Hijacker")
}

// Flush implements http.Flusher interface
func (w *responseWriterWrapper) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// CloseNotify implements http.CloseNotifier interface (deprecated but might be needed)
func (w *responseWriterWrapper) CloseNotify() <-chan bool {
	if notifier, ok := w.ResponseWriter.(http.CloseNotifier); ok {
		return notifier.CloseNotify()
	}
	// Return a channel that will never receive a value
	return make(<-chan bool)
}

// Push implements http.Pusher interface for HTTP/2 server push
func (w *responseWriterWrapper) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

// sanitizeEndpoint normalizes endpoints for metrics
func sanitizeEndpoint(path string) string {
	// Replace dynamic parts with placeholders
	if strings.HasPrefix(path, "/v1/") {
		return path
	}

	// Legacy endpoints
	switch path {
	case "/floor":
		return "/floor"
	case "/elevator":
		return "/elevator"
	case "/health":
		return "/health"
	case "/metrics":
		return "/metrics"
	case "/ws/status":
		return "/ws/status"
	default:
		return "/other"
	}
}
