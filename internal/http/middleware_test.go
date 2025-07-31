package http

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/slavakukuyev/elevator-go/internal/infra/logging"
	"github.com/stretchr/testify/assert"
)

func TestChainMiddleware(t *testing.T) {
	// Create test middlewares that add headers to track execution order
	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-Test", "middleware1")
			next.ServeHTTP(w, r)
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-Test", "middleware2")
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})

	// Chain middlewares
	chainedHandler := ChainMiddleware(middleware1, middleware2)(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	chainedHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())

	// Check that middlewares were applied in correct order
	testHeaders := w.Header()["X-Test"]
	assert.Equal(t, []string{"middleware1", "middleware2"}, testHeaders)
}

func TestRequestIDMiddleware(t *testing.T) {
	middleware := RequestIDMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := logging.GetRequestID(r.Context())
		assert.NotEmpty(t, requestID)
		w.WriteHeader(http.StatusOK)
	})

	t.Run("generates new request ID when none provided", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)

		wrappedHandler := middleware(handler)
		wrappedHandler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		requestID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID)
	})

	t.Run("uses existing request ID when provided", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		existingID := "existing-123"
		r.Header.Set("X-Request-ID", existingID)

		wrappedHandler := middleware(handler)
		wrappedHandler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		requestID := w.Header().Get("X-Request-ID")
		assert.Equal(t, existingID, requestID)
	})
}

func TestLoggingMiddleware(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	middleware := LoggingMiddleware(logger)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("test response")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test/path", nil)
	r.Header.Set("User-Agent", "test-agent")

	wrappedHandler := middleware(handler)
	wrappedHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())

	logOutput := logBuf.String()
	assert.Contains(t, logOutput, "HTTP request started")
	assert.Contains(t, logOutput, "HTTP request completed")
	assert.Contains(t, logOutput, "GET")
	assert.Contains(t, logOutput, "/test/path")
	assert.Contains(t, logOutput, "test-agent")
}

func TestRecoveryMiddleware(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	middleware := RecoveryMiddleware(logger)

	t.Run("handles panic gracefully", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		ctx := logging.WithRequestID(r.Context(), "test-123")
		r = r.WithContext(ctx)

		wrappedHandler := middleware(handler)

		// This should not panic
		assert.NotPanics(t, func() {
			wrappedHandler.ServeHTTP(w, r)
		})

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, "HTTP handler panic recovered")
		assert.Contains(t, logOutput, "test panic")
		assert.Contains(t, logOutput, "test-123")
	})

	t.Run("passes through normal requests", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("normal response")); err != nil {
				t.Errorf("failed to write response: %v", err)
			}
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)

		wrappedHandler := middleware(handler)
		wrappedHandler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "normal response", w.Body.String())
	})
}

func TestCORSMiddleware(t *testing.T) {
	middleware := CORSMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})

	t.Run("adds CORS headers to regular requests", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)

		wrappedHandler := middleware(handler)
		wrappedHandler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "OK", w.Body.String())

		// Check CORS headers
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization, X-Request-ID", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "X-Request-ID", w.Header().Get("Access-Control-Expose-Headers"))
		assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
	})

	t.Run("handles OPTIONS preflight requests", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("OPTIONS", "/test", nil)

		wrappedHandler := middleware(handler)
		wrappedHandler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Body.String())

		// CORS headers should still be present
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	middleware := SecurityHeadersMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	wrappedHandler := middleware(handler)
	wrappedHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())

	// Check security headers
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.Equal(t, "default-src 'self'", w.Header().Get("Content-Security-Policy"))
}

func TestNewRateLimitMiddleware(t *testing.T) {
	logger := slog.Default()

	t.Run("creates rate limiter with correct settings", func(t *testing.T) {
		rl := NewRateLimitMiddleware(100, logger)

		assert.NotNil(t, rl)
		assert.Equal(t, 100, rl.limit)
		assert.Equal(t, time.Minute, rl.window)
		assert.Equal(t, logger, rl.logger)
		assert.NotNil(t, rl.requests)
	})
}

func TestRateLimitMiddleware_Handler(t *testing.T) {
	logger := slog.Default()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})

	t.Run("allows requests under limit", func(t *testing.T) {
		rl := NewRateLimitMiddleware(5, logger)
		middleware := rl.Handler()
		wrappedHandler := middleware(handler)

		// Make 5 requests (under limit)
		for i := 0; i < 5; i++ {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			r.RemoteAddr = "192.168.1.1:12345"

			wrappedHandler.ServeHTTP(w, r)
			assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
		}
	})

	t.Run("blocks requests over limit", func(t *testing.T) {
		rl := NewRateLimitMiddleware(2, logger)
		middleware := rl.Handler()
		wrappedHandler := middleware(handler)

		// Make 2 requests (at limit)
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			r.RemoteAddr = "192.168.1.2:12345"

			wrappedHandler.ServeHTTP(w, r)
			assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
		}

		// Third request should be blocked
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		r.RemoteAddr = "192.168.1.2:12345"

		wrappedHandler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	})

	t.Run("different IPs have separate limits", func(t *testing.T) {
		rl := NewRateLimitMiddleware(1, logger)
		middleware := rl.Handler()
		wrappedHandler := middleware(handler)

		// Request from first IP
		w1 := httptest.NewRecorder()
		r1 := httptest.NewRequest("GET", "/test", nil)
		r1.RemoteAddr = "192.168.1.3:12345"
		wrappedHandler.ServeHTTP(w1, r1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Request from second IP should still work
		w2 := httptest.NewRecorder()
		r2 := httptest.NewRequest("GET", "/test", nil)
		r2.RemoteAddr = "192.168.1.4:12345"
		wrappedHandler.ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusOK, w2.Code)
	})
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name         string
		setupRequest func(*http.Request)
		expectedIP   string
	}{
		{
			name: "X-Forwarded-For header",
			setupRequest: func(r *http.Request) {
				r.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
			},
			expectedIP: "203.0.113.1",
		},
		{
			name: "X-Real-IP header",
			setupRequest: func(r *http.Request) {
				r.Header.Set("X-Real-IP", "203.0.113.2")
			},
			expectedIP: "203.0.113.2",
		},
		{
			name: "RemoteAddr fallback",
			setupRequest: func(r *http.Request) {
				r.RemoteAddr = "203.0.113.3:12345"
			},
			expectedIP: "203.0.113.3",
		},
		{
			name: "X-Forwarded-For takes precedence",
			setupRequest: func(r *http.Request) {
				r.Header.Set("X-Forwarded-For", "203.0.113.4")
				r.Header.Set("X-Real-IP", "203.0.113.5")
				r.RemoteAddr = "203.0.113.6:12345"
			},
			expectedIP: "203.0.113.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/test", nil)
			tt.setupRequest(r)

			ip := getClientIP(r)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}

func TestGenerateRequestID(t *testing.T) {
	// Generate multiple request IDs
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := generateRequestID()
		assert.NotEmpty(t, id)
		assert.False(t, ids[id], "Request ID should be unique: %s", id)
		ids[id] = true
	}
}

func TestGetRequestID(t *testing.T) {
	t.Run("returns request ID from context", func(t *testing.T) {
		expectedID := "context-123"
		ctx := context.WithValue(context.Background(), "request_id", expectedID)
		r := httptest.NewRequest("GET", "/test", nil)
		r = r.WithContext(ctx)

		id := getRequestID(r)
		assert.Equal(t, expectedID, id)
	})

	t.Run("generates new ID when not in context", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)

		id := getRequestID(r)
		assert.NotEmpty(t, id)
	})
}

func TestStatusResponseWriter(t *testing.T) {
	w := httptest.NewRecorder()
	srw := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	// Test default status code
	assert.Equal(t, http.StatusOK, srw.statusCode)

	// Test WriteHeader updates status code
	srw.WriteHeader(http.StatusNotFound)
	assert.Equal(t, http.StatusNotFound, srw.statusCode)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Test multiple calls (should only set once)
	srw.WriteHeader(http.StatusInternalServerError)
	assert.Equal(t, http.StatusNotFound, w.Code) // Should remain the first status
}

func TestMiddlewareIntegration(t *testing.T) {
	// Test that all middlewares work together
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request ID is available
		requestID := getRequestID(r)
		assert.NotEmpty(t, requestID)

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("integration test")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})

	// Create middleware chain
	rl := NewRateLimitMiddleware(10, logger)
	middlewareChain := ChainMiddleware(
		RequestIDMiddleware(),
		LoggingMiddleware(logger),
		RecoveryMiddleware(logger),
		CORSMiddleware(),
		SecurityHeadersMiddleware(),
		rl.Handler(),
	)

	wrappedHandler := middlewareChain(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/integration", nil)
	r.Header.Set("Origin", "https://example.com")

	wrappedHandler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "integration test", w.Body.String())

	// Check that all middleware functionality is working
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))

	// Check logging occurred
	logOutput := logBuf.String()
	assert.Contains(t, logOutput, "HTTP request started")
	assert.Contains(t, logOutput, "HTTP request completed")
}
