package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Config holds logger configuration with support for env vars, JSON, and YAML.
type Config struct {
	Level  string `envconfig:"LOG_LEVEL"  default:"INFO"  json:"level"  yaml:"level"`
	Output string `envconfig:"LOG_OUTPUT" default:"STDOUT" json:"output" yaml:"output"`
	Format string `envconfig:"LOG_FORMAT" default:"JSON"  json:"format" yaml:"format"`
	Source *bool  `envconfig:"LOG_SOURCE" default:"true"  json:"source" yaml:"source"`
}

// NewLogger sets and returns a new default logger configured by the provided Config.
func NewLogger(cfg Config) *slog.Logger {
	handler := NewBaggageHandler(
		NewOtelHandler(
			cfg.LogFormat(),
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

// LogLevel returns the slog level from the config. Defaults to INFO.
func (c Config) LogLevel() slog.Leveler {
	switch strings.ToUpper(c.Level) {
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

// LogOutput returns the slog output writer from the config.
// Supports "STDOUT", "STDERR", or a file path with automatic log rotation.
// Defaults to os.Stdout.
func (c Config) LogOutput() io.Writer {
	switch strings.ToUpper(c.Output) {
	case "STDOUT", "":
		return os.Stdout
	case "STDERR":
		return os.Stderr
	default:
		return &lumberjack.Logger{
			Filename:   c.Output,
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		}
	}
}

// LogSource returns whether to include source information in logs. Defaults to true.
func (c Config) LogSource() bool {
	if c.Source == nil {
		return true
	}
	return *c.Source
}

// LogFormat returns the slog handler based on the config format. Defaults to JSON.
func (c Config) LogFormat() slog.Handler {
	switch strings.ToUpper(c.Format) {
	case "TEXT":
		return slog.NewTextHandler(c.LogOutput(), &slog.HandlerOptions{
			Level:     c.LogLevel(),
			AddSource: c.LogSource(),
		})
	default:
		return slog.NewJSONHandler(c.LogOutput(), &slog.HandlerOptions{
			Level:     c.LogLevel(),
			AddSource: c.LogSource(),
		})
	}
}
