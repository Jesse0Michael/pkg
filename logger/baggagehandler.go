package logger

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/baggage"
)

// BaggageHandler is a slog.Handler that adds baggage attributes to the log record from the context.
type BaggageHandler struct {
	slog.Handler
}

// NewBaggageHandler returns a new BaggageHandler.
func NewBaggageHandler(handler slog.Handler) slog.Handler {
	return &BaggageHandler{
		Handler: handler,
	}
}

// WithAttrs returns a new BaggageHandler with the given attributes.
func (h BaggageHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &BaggageHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroup returns a new BaggageHandler with the given group name.
func (h BaggageHandler) WithGroup(name string) slog.Handler {
	return &BaggageHandler{Handler: h.Handler.WithGroup(name)}
}

// Handle adds baggage attributes from the context to the log record.
func (h BaggageHandler) Handle(ctx context.Context, record slog.Record) error {
	b := baggage.FromContext(ctx)
	for _, m := range b.Members() {
		record.AddAttrs(slog.String(m.Key(), m.Value()))
	}
	return h.Handler.Handle(ctx, record)
}
