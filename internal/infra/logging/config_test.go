package logging

import (
	"log/slog"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected slog.Level
	}{
		{
			name:     "debug level",
			input:    "DEBUG",
			expected: slog.LevelDebug,
		},
		{
			name:     "debug level lowercase",
			input:    "debug",
			expected: slog.LevelDebug,
		},
		{
			name:     "info level",
			input:    "INFO",
			expected: slog.LevelInfo,
		},
		{
			name:     "info level lowercase",
			input:    "info",
			expected: slog.LevelInfo,
		},
		{
			name:     "warn level",
			input:    "WARN",
			expected: slog.LevelWarn,
		},
		{
			name:     "warning level",
			input:    "WARNING",
			expected: slog.LevelWarn,
		},
		{
			name:     "warn level lowercase",
			input:    "warn",
			expected: slog.LevelWarn,
		},
		{
			name:     "error level",
			input:    "ERROR",
			expected: slog.LevelError,
		},
		{
			name:     "error level lowercase",
			input:    "error",
			expected: slog.LevelError,
		},
		{
			name:     "invalid level defaults to info",
			input:    "INVALID",
			expected: slog.LevelInfo,
		},
		{
			name:     "empty string defaults to info",
			input:    "",
			expected: slog.LevelInfo,
		},
		{
			name:     "mixed case",
			input:    "DeBuG",
			expected: slog.LevelDebug,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
	}{
		{
			name:     "init with debug level",
			logLevel: "DEBUG",
		},
		{
			name:     "init with info level",
			logLevel: "INFO",
		},
		{
			name:     "init with warn level",
			logLevel: "WARN",
		},
		{
			name:     "init with error level",
			logLevel: "ERROR",
		},
		{
			name:     "init with invalid level",
			logLevel: "INVALID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test mainly ensures InitLogger doesn't panic
			// and can be called with different log levels
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("InitLogger(%q) panicked: %v", tt.logLevel, r)
				}
			}()

			InitLogger(tt.logLevel)
		})
	}
}
