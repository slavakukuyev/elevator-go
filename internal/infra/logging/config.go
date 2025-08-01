package logging

import (
	"log/slog"
	"os"
	"strings"
)

// InitLogger configures the global slog logger with JSON handler
func InitLogger(logLevel string) {
	level := parseLogLevel(logLevel)

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Rename default keys to match common observability standards
			if a.Key == slog.TimeKey {
				a.Key = "timestamp"
			}
			if a.Key == slog.LevelKey {
				a.Key = "level"
			}
			if a.Key == slog.MessageKey {
				a.Key = "message"
			}
			return a
		},
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// parseLogLevel converts string log level to slog.Level
// Defaults to INFO for production safety
func parseLogLevel(logLevel string) slog.Level {
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		// Default to INFO for production safety
		return slog.LevelInfo
	}
}
