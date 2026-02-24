package auth

import (
	"context"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestWithSpan(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	tracer := tp.Tracer("test-tracer")

	tests := []struct {
		name             string
		ctx              context.Context
		wantSubjectAttr  bool
		wantAdminAttr    bool
		wantReadOnlyAttr bool
	}{
		{
			name: "empty context",
			ctx:  t.Context(),
		},
		{
			name: "with all values",
			ctx: func() context.Context {
				ctx := context.WithValue(t.Context(), SubjectContextKey, "test-subject")
				ctx = context.WithValue(ctx, AdminContextKey, true)
				ctx = context.WithValue(ctx, ReadOnlyContextKey, true)
				return ctx
			}(),
			wantSubjectAttr:  true,
			wantAdminAttr:    true,
			wantReadOnlyAttr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exporter.Reset()
			ctx, span := tracer.Start(tt.ctx, "test-span")
			ctx = WithSpan(ctx)
			span.End()

			spans := exporter.GetSpans()
			if len(spans) != 1 {
				t.Fatalf("expected 1 span, got %d", len(spans))
			}

			attrs := spans[0].Attributes
			hasAttr := func(key string) bool {
				for _, a := range attrs {
					if string(a.Key) == key {
						return true
					}
				}
				return false
			}

			if hasAttr("subject") != tt.wantSubjectAttr {
				t.Errorf("subject attr = %v, want %v", hasAttr("subject"), tt.wantSubjectAttr)
			}
			if hasAttr("admin") != tt.wantAdminAttr {
				t.Errorf("admin attr = %v, want %v", hasAttr("admin"), tt.wantAdminAttr)
			}
			if hasAttr("readOnly") != tt.wantReadOnlyAttr {
				t.Errorf("readOnly attr = %v, want %v", hasAttr("readOnly"), tt.wantReadOnlyAttr)
			}

			traceID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
			reqID, _ := ctx.Value(RequestContextKey).(string)
			if reqID != traceID {
				t.Errorf("RequestContextKey = %v, want %v", reqID, traceID)
			}
		})
	}
}
