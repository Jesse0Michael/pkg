package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// SetLog sets the default logger.
func SetLog() {
	handler := NewContextHandler(
		slog.NewJSONHandler(LogOutput(), &slog.HandlerOptions{
			Level: LogLevel(),
		}),
	)

	logger := slog.New(handler)
	logger = logger.With(
		"host", os.Getenv("HOSTNAME"),
		"environment", os.Getenv("ENVIRONMENT"),
	)

	slog.SetDefault(logger)
}

// LogLevel returns the slog level from the LOG_LEVEL environment variable.
func LogLevel() slog.Leveler {
	switch strings.ToUpper(os.Getenv("LOG_LEVEL")) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// LogOutput returns the slog output from the LOG_OUTPUT environment variable.
func LogOutput() io.Writer {
	switch strings.ToUpper(os.Getenv("LOG_OUTPUT")) {
	case "STDOUT":
		return os.Stdout
	default:
		return os.Stderr
	}
}
