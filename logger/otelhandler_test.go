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

	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

type MockIDGenerator struct {
}

func (m MockIDGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	traceID := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	spanID := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	return trace.TraceID(traceID), trace.SpanID(spanID)
}

func (m MockIDGenerator) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	spanID := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	return trace.SpanID(spanID)
}

func TestOtelHandler_Handle(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		log  string
	}{
		{
			name: "nil context",
			ctx:  context.TODO(),
			log:  `{"level":"DEBUG","msg":"message"}`,
		},
		{
			name: "empty context",
			ctx: func() context.Context {

				ctx, _ := tracesdk.NewTracerProvider(tracesdk.WithIDGenerator(&MockIDGenerator{}), tracesdk.WithSpanProcessor(tracetest.NewSpanRecorder())).Tracer("test").Start(context.TODO(), "test")
				return ctx
			}(),
			log: `{"level":"DEBUG","msg":"message","trace_id":"0102030405060708090a0b0c0d0e0f10","span_id":"0102030405060708"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _ := os.CreateTemp("", "out")
			os.Stderr = f

			h := NewOtelHandler(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{}))
			_ = h.Handle(tt.ctx, slog.NewRecord(time.Time{}, slog.LevelDebug, "message", 0))

			_, _ = f.Seek(0, 0)
			b, _ := io.ReadAll(f)
			if !reflect.DeepEqual(tt.log, strings.TrimSpace((string(b)))) {
				t.Errorf("OtelHandler().Handle = %v, want %v", string(b), tt.log)
			}

			f.Close()
		})
	}
}
