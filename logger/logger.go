package logger

import (
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// NewLogger sets and returns a new default logger.
func NewLogger() *slog.Logger {
	handler := NewBaggageHandler(
		NewOtelHandler(
			LogFormat(),
		),
	)

	logger := slog.New(handler)

	// Default log attributes taken from environment variables.
	if host, ok := os.LookupEnv("HOSTNAME"); ok {
		logger = logger.With("host", host)
	}
	if env, ok := os.LookupEnv("ENVIRONMENT"); ok {
		logger = logger.With("env", env)
	}

	slog.SetDefault(logger)
	return logger
}

// LogLevel returns the slog level from the LOG_LEVEL environment variable.
// Defaults to INFO.
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
// Supports "STDOUT", "STDERR", or a file path with automatic log rotation.
// Defaults to os.Stdout.
func LogOutput() io.Writer {
	output := os.Getenv("LOG_OUTPUT")
	switch strings.ToUpper(output) {
	case "STDOUT", "":
		return os.Stdout
	case "STDERR":
		return os.Stderr
	default:
		return &lumberjack.Logger{
			Filename:   output,
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		}
	}
}

// LogSource returns whether to include source information in logs from the LOG_SOURCE environment variable.
// Defaults to true.
func LogSource() bool {
	value, err := strconv.ParseBool(os.Getenv("LOG_SOURCE"))
	if err != nil {
		return true
	}
	return value
}

// LogFormat returns the slog handler to use based on the LOG_FORMAT environment variable.
// Defaults to JSON Handler.
func LogFormat() slog.Handler {
	switch strings.ToUpper(os.Getenv("LOG_FORMAT")) {
	case "TEXT":
		return slog.NewTextHandler(LogOutput(), &slog.HandlerOptions{
			Level:     LogLevel(),
			AddSource: LogSource(),
		})
	default:
		return slog.NewJSONHandler(LogOutput(), &slog.HandlerOptions{
			Level:     LogLevel(),
			AddSource: LogSource(),
		})
	}
}
