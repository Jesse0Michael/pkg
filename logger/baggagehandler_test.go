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

	"go.opentelemetry.io/otel/baggage"
)

func ExampleBaggageHandler() {
	// Setup test
	f, _ := os.CreateTemp("", "out")
	origStderr := os.Stderr
	os.Stderr = f
	defer func() {
		os.Stderr = origStderr
		_ = f.Close()
	}()
	ctx := context.Background()

	// Example
	h := NewBaggageHandler(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{}))
	slog.SetDefault(slog.New(h))

	member, _ := baggage.NewMember("accountID", "12345")
	bag, _ := baggage.New(member)
	ctx = baggage.ContextWithBaggage(ctx, bag)

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

func TestBaggageHandler_Handle(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		log  string
	}{
		{
			name: "no baggage",
			ctx:  context.TODO(),
			log:  `{"level":"DEBUG","msg":"message"}`,
		},
		{
			name: "single member",
			ctx: func() context.Context {
				m, _ := baggage.NewMember("accountID", "12345")
				b, _ := baggage.New(m)
				return baggage.ContextWithBaggage(context.TODO(), b)
			}(),
			log: `{"level":"DEBUG","msg":"message","accountID":"12345"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _ := os.CreateTemp("", "out")
			os.Stderr = f

			h := NewBaggageHandler(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{}))
			_ = h.Handle(tt.ctx, slog.NewRecord(time.Time{}, slog.LevelDebug, "message", 0))

			_, _ = f.Seek(0, 0)
			b, _ := io.ReadAll(f)
			if !reflect.DeepEqual(tt.log, strings.TrimSpace(string(b))) {
				t.Errorf("BaggageHandler().Handle = %v, want %v", string(b), tt.log)
			}

			f.Close()
		})
	}
}
