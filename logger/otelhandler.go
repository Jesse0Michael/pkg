package logger

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// OtelHandler is a slog.Handler that adds attributes to the log record from the context.
type OtelHandler struct {
	slog.Handler
}

// NewOtelHandler returns a new OtelHandler.
func NewOtelHandler(handler slog.Handler) slog.Handler {
	return &OtelHandler{
		Handler: handler,
	}
}

// WithAttrs returns a new OtelHandler with the given attributes.
func (h OtelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &OtelHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroup returns a new OtelHandler with the given group name.
func (h OtelHandler) WithGroup(name string) slog.Handler {
	return &OtelHandler{Handler: h.Handler.WithGroup(name)}
}

// Handle adds attributes from the context to the log record.
func (h OtelHandler) Handle(ctx context.Context, record slog.Record) error {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		if span.SpanContext().HasTraceID() {
			record.AddAttrs(slog.String("trace_id", span.SpanContext().TraceID().String()))
		}
		if span.SpanContext().HasSpanID() {
			record.AddAttrs(slog.String("span_id", span.SpanContext().SpanID().String()))
		}
	}
	return h.Handler.Handle(ctx, record)
}
