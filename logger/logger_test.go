package logger

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

func ExampleNewLogger() {
	// Setup test
	f, _ := os.CreateTemp("", "out")
	origStderr := os.Stderr
	os.Stderr = f
	defer func() {
		os.Stderr = origStderr
		_ = f.Close()
	}()

	// Predefine environment
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("HOSTNAME", "local")
	ctx := context.Background()

	// Example
	NewLogger(Config{Output: "STDERR", Source: new(false)})

	slog.With("key", "value").InfoContext(ctx, "writing logs")
	slog.With("error", errors.New("error")).ErrorContext(ctx, "writing errors")

	// Post process output
	ReplaceTimestamps(f, os.Stdout)

	// Output:
	// {"time":"TIMESTAMP","level":"INFO","msg":"writing logs","host":"local","env":"test","key":"value"}
	// {"time":"TIMESTAMP","level":"ERROR","msg":"writing errors","host":"local","env":"test","error":"error"}
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name  string
		setup func()
		cfg   Config
		log   string
	}{
		{
			name:  "set log",
			setup: func() {},
			cfg:   Config{},
			log:   `{"level":"DEBUG","msg":"message"}`,
		},
		{
			name: "set log with environment",
			setup: func() {
				t.Setenv("ENVIRONMENT", "test")
				t.Setenv("HOSTNAME", "local")
			},
			cfg: Config{Level: "DEBUG"},
			log: `{"level":"DEBUG","msg":"message","host":"local","env":"test"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.CreateTemp("", "out")
			if err != nil {
				t.Fatalf("create temp stdout: %v", err)
			}
			origStdout := os.Stdout
			os.Stdout = f
			t.Cleanup(func() {
				os.Stdout = origStdout
				_ = f.Close()
				_ = os.Remove(f.Name())
			})

			os.Clearenv()
			tt.setup()
			NewLogger(tt.cfg)

			_ = slog.Default().Handler().Handle(t.Context(), slog.NewRecord(time.Time{}, slog.LevelDebug, "message", 0))
			_, _ = f.Seek(0, 0)
			b, _ := io.ReadAll(f)
			if !reflect.DeepEqual(tt.log, strings.TrimSpace((string(b)))) {
				t.Errorf("slog().Handle = %v, want %v", string(b), tt.log)
			}
		})
	}
}

func TestConfig_LogLevel(t *testing.T) {
	tests := []struct {
		name  string
		cfg   Config
		level slog.Leveler
	}{
		{
			name:  "log level: empty",
			cfg:   Config{},
			level: slog.LevelInfo,
		},
		{
			name:  "log level: debug",
			cfg:   Config{Level: "debug"},
			level: slog.LevelDebug,
		},
		{
			name:  "log level: INFO",
			cfg:   Config{Level: "INFO"},
			level: slog.LevelInfo,
		},
		{
			name:  "log level: WARN",
			cfg:   Config{Level: "WARN"},
			level: slog.LevelWarn,
		},
		{
			name:  "log level: ErRoR",
			cfg:   Config{Level: "ErRoR"},
			level: slog.LevelError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.LogLevel(); !reflect.DeepEqual(got, tt.level) {
				t.Errorf("LogLevel() = %v, want %v", got, tt.level)
			}
		})
	}
}

func TestConfig_LogOutput(t *testing.T) {
	tests := []struct {
		name   string
		cfg    Config
		output io.Writer
	}{
		{
			name:   "log output: empty",
			cfg:    Config{},
			output: os.Stdout,
		},
		{
			name:   "log output: stderr",
			cfg:    Config{Output: "stderr"},
			output: os.Stderr,
		},
		{
			name:   "log output: stdout",
			cfg:    Config{Output: "STDOUT"},
			output: os.Stdout,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.LogOutput(); !reflect.DeepEqual(got, tt.output) {
				t.Errorf("LogOutput() = %v, want %v", got, tt.output)
			}
		})
	}
}

func TestConfig_LogOutput_file(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "test.log")
	cfg := Config{Output: logFile}

	got := cfg.LogOutput()
	lj, ok := got.(*lumberjack.Logger)
	if !ok {
		t.Fatalf("LogOutput() = %T, want *lumberjack.Logger", got)
	}
	if lj.Filename != logFile {
		t.Errorf("Filename = %v, want %v", lj.Filename, logFile)
	}
}

func TestConfig_LogOutput_file_writes(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "test.log")
	cfg := Config{Output: logFile}

	l := slog.New(slog.NewJSONHandler(cfg.LogOutput(), &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	l.InfoContext(t.Context(), "test-message", "key", "test-value")

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "test-message") {
		t.Errorf("log file missing message, got: %s", content)
	}
	if !strings.Contains(content, "test-value") {
		t.Errorf("log file missing attribute, got: %s", content)
	}
}

func TestConfig_LogSource(t *testing.T) {
	tests := []struct {
		name   string
		cfg    Config
		source bool
	}{
		{
			name:   "log source: nil (default true)",
			cfg:    Config{},
			source: true,
		},
		{
			name:   "log source: true",
			cfg:    Config{Source: new(true)},
			source: true,
		},
		{
			name:   "log source: false",
			cfg:    Config{Source: new(false)},
			source: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.LogSource(); !reflect.DeepEqual(got, tt.source) {
				t.Errorf("LogSource() = %v, want %v", got, tt.source)
			}
		})
	}
}

func TestConfig_LogFormat(t *testing.T) {
	tests := []struct {
		name   string
		cfg    Config
		format slog.Handler
	}{
		{
			name:   "log format: empty",
			cfg:    Config{},
			format: slog.NewJSONHandler(os.Stdout, nil),
		},
		{
			name:   "log format: JSON",
			cfg:    Config{Format: "JSON"},
			format: slog.NewJSONHandler(os.Stdout, nil),
		},
		{
			name:   "log format: TEXT",
			cfg:    Config{Format: "TEXT"},
			format: slog.NewTextHandler(os.Stdout, nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.LogFormat(); reflect.TypeOf(got) != reflect.TypeOf(tt.format) {
				t.Errorf("LogFormat() = %T, want %T", got, tt.format)
			}
		})
	}
}

func ReplaceTimestamps(input *os.File, output io.Writer) {
	_, _ = input.Seek(0, 0)
	reader := bufio.NewReader(input)
	writer := bufio.NewWriter(output)

	// Regular expression to match RFC3339Nano timestamps
	timestampRegex := regexp.MustCompile(`((?:(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2}(?:\.\d+)?))(Z|[\+-]\d{2}:\d{2})?)`)

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Println("Error reading:", err)
			return
		}

		// Check for the end of input
		if err == io.EOF {
			break
		}

		// Find all timestamps in the line
		matches := timestampRegex.FindAllStringIndex(line, -1)
		if matches != nil {
			var modifiedLine string
			lastIndex := 0

			// Replace each timestamp with "TIMESTAMP"
			for _, match := range matches {
				modifiedLine += line[lastIndex:match[0]] + "TIMESTAMP"
				lastIndex = match[1]
			}
			modifiedLine += line[lastIndex:]

			// Write the modified line to the output
			_, err := writer.WriteString(modifiedLine)
			if err != nil {
				fmt.Println("Error writing:", err)
				return
			}
		} else {
			// If no timestamps found, write the original line to the output
			_, err := writer.WriteString(line)
			if err != nil {
				fmt.Println("Error writing:", err)
				return
			}
		}
	}

	writer.Flush()
}
