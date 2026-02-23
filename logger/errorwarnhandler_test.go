package logger

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func ExampleErrorWarnHandler() {
	// Setup test
	f, _ := os.CreateTemp("", "out")
	os.Stderr = f
	defer f.Close()

	// Predefine environment
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("HOSTNAME", "local")
	os.Setenv("LOG_OUTPUT", "stderr")
	os.Setenv("LOG_SOURCE", "false")
	ctx := context.Background()
	warnCheck := func(err error) bool { return err != nil && err.Error() == "warn" }

	// Example
	NewLogger()
	SetErrorWarnHandler(warnCheck)

	slog.ErrorContext(ctx, "writing errors", "error", errors.New("error"))
	slog.ErrorContext(ctx, "writing errors that should be warnings", "error", errors.New("warn"))
	slog.Default().WithGroup("request").With("path", "/", "verb", "GET").ErrorContext(ctx, "writing errors that should be warnings with group", "error", errors.New("warn"))

	// Post process output
	ReplaceTimestamps(f, os.Stdout)

	// Output:
	// {"time":"TIMESTAMP","level":"ERROR","msg":"writing errors","host":"local","env":"test","error":"error"}
	// {"time":"TIMESTAMP","level":"WARN","msg":"writing errors that should be warnings","host":"local","env":"test","error":"warn"}
	// {"time":"TIMESTAMP","level":"WARN","msg":"writing errors that should be warnings with group","host":"local","env":"test","request":{"path":"/","verb":"GET","error":"warn"}}
}

func TestErrorWarnHandler_Handle(t *testing.T) {
	tests := []struct {
		name  string
		level slog.Level
		attrs []slog.Attr
		log   string
	}{
		{
			name:  "debug record",
			attrs: []slog.Attr{},
			level: slog.LevelDebug,
			log:   `{"level":"DEBUG","msg":"message"}`,
		},
		{
			name:  "no error attribute",
			attrs: []slog.Attr{slog.Bool("flag", true), slog.Float64("thing", 0.99), slog.String("key", "value")},
			level: slog.LevelError,
			log:   `{"level":"ERROR","msg":"message","flag":true,"thing":0.99,"key":"value"}`,
		},
		{
			name:  "error attribute",
			attrs: []slog.Attr{slog.Any("error", errors.New("test-error"))},
			level: slog.LevelError,
			log:   `{"level":"ERROR","msg":"message","error":"test-error"}`,
		},
		{
			name:  "warn error attribute",
			attrs: []slog.Attr{slog.Any("error", errors.New("warn"))},
			level: slog.LevelError,
			log:   `{"level":"WARN","msg":"message","error":"warn"}`,
		},
		{
			name:  "warn error with non-error key",
			attrs: []slog.Attr{slog.Any("cause", errors.New("warn"))},
			level: slog.LevelError,
			log:   `{"level":"WARN","msg":"message","cause":"warn"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _ := os.CreateTemp("", "out")
			os.Stderr = f

			h := NewErrorWarnHandler(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{}),
				func(err error) bool { return err.Error() == "warn" })
			record := slog.NewRecord(time.Time{}, tt.level, "message", 0)
			record.AddAttrs(tt.attrs...)
			_ = h.Handle(context.TODO(), record)

			_, _ = f.Seek(0, 0)
			b, _ := io.ReadAll(f)
			if !reflect.DeepEqual(tt.log, strings.TrimSpace((string(b)))) {
				t.Errorf("ErrorWarnHandler().Handle = %v, want %v", string(b), tt.log)
			}

			f.Close()
		})
	}
}
