package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// NewLogger sets and returns a new default logger.
func NewLogger() *slog.Logger {
	handler := NewContextHandler(
		NewOtelHandler(
			slog.NewJSONHandler(LogOutput(), &slog.HandlerOptions{
				Level: LogLevel(),
			}),
		),
	)

	logger := slog.New(handler)

	// Default log attributes taken from environment variables.
	if host, ok := os.LookupEnv("HOSTNAME"); ok {
		logger = logger.With("host", host)
	}
	if env, ok := os.LookupEnv("ENVIRONMENT"); ok {
		logger = logger.With("environment", env)
	}

	slog.SetDefault(logger)
	return logger
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
	case "STDERR":
		return os.Stderr
	default:
		return os.Stdout
	}
}
