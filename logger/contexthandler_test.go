package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func ExampleContextHandler() {
	// Setup test
	f, _ := os.CreateTemp("", "out")
	os.Stderr = f
	defer f.Close()

	// Predefine environment
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("HOSTNAME", "local")
	ctx := context.Background()

	// Example
	handler := NewContextHandler(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{}))
	slog.SetDefault(slog.New(handler))

	ctx = AddAttrs(ctx, slog.String("accountID", "12345"))

	slog.InfoContext(ctx, "writing logs")
	slog.With("key", "value").InfoContext(ctx, "writing logs with attributes")
	slog.Default().WithGroup("request").With("path", "/", "verb", "GET").InfoContext(ctx, "writing logs with group")

	// Post process output
	ReplaceTimestamps(f, os.Stdout)

	// Output:
	// {"time":"TIMESTAMP","level":"INFO","msg":"writing logs","accountID":"12345"}
	// {"time":"TIMESTAMP","level":"INFO","msg":"writing logs with attributes","key":"value","accountID":"12345"}
	// {"time":"TIMESTAMP","level":"INFO","msg":"writing logs with group","request":{"path":"/","verb":"GET","accountID":"12345"}}
}

func TestContextHandler_Handle(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		log  string
	}{
		{
			name: "nil context",
			ctx:  t.Context(),
			log:  `{"level":"DEBUG","msg":"message"}`,
		},
		{
			name: "empty context",
			ctx:  context.WithValue(t.Context(), contextHandlerKey, map[string]any{}),
			log:  `{"level":"DEBUG","msg":"message"}`,
		},
		{
			name: "attr context",
			ctx:  context.WithValue(t.Context(), contextHandlerKey, map[string]any{"key": "value"}),
			log:  `{"level":"DEBUG","msg":"message","key":"value"}`,
		},
		{
			name: "invalid context",
			ctx:  context.WithValue(t.Context(), contextHandlerKey, true),
			log:  `{"level":"DEBUG","msg":"message"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _ := os.CreateTemp("", "out")
			os.Stderr = f

			h := NewContextHandler(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{}))
			_ = h.Handle(tt.ctx, slog.NewRecord(time.Time{}, slog.LevelDebug, "message", 0))

			_, _ = f.Seek(0, 0)
			b, _ := io.ReadAll(f)
			if !reflect.DeepEqual(tt.log, strings.TrimSpace((string(b)))) {
				t.Errorf("ContextHandler().Handle = %v, want %v", string(b), tt.log)
			}

			f.Close()
		})
	}
}

func TestAddAttrs(t *testing.T) {
	tests := []struct {
		name  string
		ctx   context.Context
		attrs []slog.Attr
		want  map[string]any
	}{
		{
			name:  "nil context",
			ctx:   t.Context(),
			attrs: []slog.Attr{slog.String("key", "value")},
			want: map[string]any{
				"key": "value",
			},
		},
		{
			name:  "empty context",
			ctx:   context.WithValue(t.Context(), contextHandlerKey, map[string]any{}),
			attrs: []slog.Attr{slog.String("key", "value")},
			want: map[string]any{
				"key": "value",
			},
		},
		{
			name:  "invalid context",
			ctx:   context.WithValue(t.Context(), contextHandlerKey, true),
			attrs: []slog.Attr{slog.String("key", "value")},
			want: map[string]any{
				"key": "value",
			},
		},
		{
			name:  "attr context",
			ctx:   context.WithValue(t.Context(), contextHandlerKey, map[string]any{"key": "old", "other": "thing"}),
			attrs: []slog.Attr{slog.String("key", "value")},
			want: map[string]any{
				"key":   "value",
				"other": "thing",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := AddAttrs(tt.ctx, tt.attrs...)
			got := ctx.Value(contextHandlerKey).(map[string]any)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddAttrs() = %v, want %v", got, tt.want)
			}
		})
	}
}
