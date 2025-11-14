package logger

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
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
	os.Setenv("LOG_OUTPUT", "stderr")
	ctx := context.Background()

	// Example
	NewLogger()

	slog.With("key", "value").InfoContext(ctx, "writing logs")
	slog.With("error", errors.New("error")).ErrorContext(ctx, "writing errors")

	// Post process output
	ReplaceTimestamps(f, os.Stdout)

	// Output:
	// {"time":"TIMESTAMP","level":"INFO","msg":"writing logs","host":"local","environment":"test","key":"value"}
	// {"time":"TIMESTAMP","level":"ERROR","msg":"writing errors","host":"local","environment":"test","error":"error"}
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name  string
		setup func()
		log   string
	}{
		{
			name:  "set log",
			setup: func() {},
			log:   `{"level":"DEBUG","msg":"message"}`,
		},
		{
			name: "set log with environment",
			setup: func() {
				t.Setenv("LOG_LEVEL", "debug")
				t.Setenv("ENVIRONMENT", "test")
				t.Setenv("HOSTNAME", "local")
			},
			log: `{"level":"DEBUG","msg":"message","host":"local","environment":"test"}`,
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
			NewLogger()

			_ = slog.Default().Handler().Handle(context.TODO(), slog.NewRecord(time.Time{}, slog.LevelDebug, "message", 0))
			_, _ = f.Seek(0, 0)
			b, _ := io.ReadAll(f)
			if !reflect.DeepEqual(tt.log, strings.TrimSpace((string(b)))) {
				t.Errorf("slog().Handle = %v, want %v", string(b), tt.log)
			}
		})
	}
}

func Test_slogLevel(t *testing.T) {
	tests := []struct {
		name  string
		setup func()
		level slog.Leveler
	}{
		{
			name:  "log level: empty",
			setup: func() {},
			level: slog.LevelInfo,
		},
		{
			name: "log level: debug",
			setup: func() {
				t.Setenv("LOG_LEVEL", "debug")
			},
			level: slog.LevelDebug,
		},
		{
			name: "log level: INFO",
			setup: func() {
				t.Setenv("LOG_LEVEL", "INFO")
			},
			level: slog.LevelInfo,
		},
		{
			name: "log level: WARN",
			setup: func() {
				t.Setenv("LOG_LEVEL", "WARN")
			},
			level: slog.LevelWarn,
		},
		{
			name: "log level: ErRoR",
			setup: func() {
				t.Setenv("LOG_LEVEL", "ErRoR")
			},
			level: slog.LevelError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			tt.setup()
			if got := LogLevel(); !reflect.DeepEqual(got, tt.level) {
				t.Errorf("slogLevel() = %v, want %v", got, tt.level)
			}
		})
	}
}

func Test_slogOut(t *testing.T) {
	tests := []struct {
		name   string
		setup  func()
		output io.Writer
	}{
		{
			name:   "log output: empty",
			setup:  func() {},
			output: os.Stdout,
		},
		{
			name: "log output: stderr",
			setup: func() {
				t.Setenv("LOG_OUTPUT", "stderr")
			},
			output: os.Stderr,
		},
		{
			name: "log output: stdout",
			setup: func() {
				t.Setenv("LOG_OUTPUT", "STDOUT")
			},
			output: os.Stdout,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			tt.setup()
			if got := LogOutput(); !reflect.DeepEqual(got, tt.output) {
				t.Errorf("slogOutput() = %v, want %v", got, tt.output)
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
