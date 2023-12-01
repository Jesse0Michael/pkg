package logger

import (
	"context"
	"log/slog"
)

type contextKey string

const contextHandlerKey = contextKey("attrs")

// ContextHandler is a slog.Handler that adds attributes to the log record from the context.
type ContextHandler struct {
	slog.Handler
}

// NewContextHandler returns a new ContextHandler.
func NewContextHandler(handler slog.Handler) slog.Handler {
	return &ContextHandler{
		Handler: handler,
	}
}

// WithAttrs returns a new ContextHandler with the given attributes.
func (h ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroup returns a new ContextHandler with the given group name.
func (h ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{Handler: h.Handler.WithGroup(name)}
}

// Handle adds attributes from the context to the log record.
func (h ContextHandler) Handle(ctx context.Context, record slog.Record) error {
	if v, ok := ctx.Value(contextHandlerKey).(map[string]any); ok {
		for key, val := range v {
			record.AddAttrs(slog.Any(key, val))
		}
	}
	return h.Handler.Handle(ctx, record)
}

// AddAttrs adds attributes to the context.
func AddAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	v, ok := ctx.Value(contextHandlerKey).(map[string]any)
	if !ok {
		v = map[string]any{}
	}
	for _, attr := range attrs {
		v[attr.Key] = attr.Value.Any()
	}
	return context.WithValue(ctx, contextHandlerKey, v)
}
