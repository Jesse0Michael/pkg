package logger

import (
	"context"
	"log/slog"
)

// ErrorCheckFunction is a function that checks if an error matches a condition.
type ErrorCheckFunction func(error) bool

// ErrorWarnHandler is a slog.Handler that will downgrade the log level of a record to WARN
// if there is an "error" attr that match one of the error check functions.
type ErrorWarnHandler struct {
	slog.Handler
	f   []ErrorCheckFunction
	err error
}

// NewErrorWarnHandler returns a new ErrorWarnHandler.
func NewErrorWarnHandler(handler slog.Handler, funcs ...ErrorCheckFunction) slog.Handler {
	return &ErrorWarnHandler{
		Handler: handler,
		f:       funcs,
	}
}

// WithAttrs returns a new ContextHandler with the given attributes.
func (h ErrorWarnHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var err error
	for _, attr := range attrs {
		if e, ok := attr.Value.Any().(error); ok && e != nil {
			err = e
		}
	}
	return &ErrorWarnHandler{Handler: h.Handler.WithAttrs(attrs), f: h.f, err: err}
}

// WithGroup returns a new ContextHandler with the given group name.
func (h ErrorWarnHandler) WithGroup(name string) slog.Handler {
	return &ErrorWarnHandler{Handler: h.Handler.WithGroup(name), f: h.f}
}

// Handle checks the log record for an error attr and compares it against the ErrorCheckFunctions.
// If any of the ErrorCheckFunctions return true, the log level will be overridden to WARN.
func (h ErrorWarnHandler) Handle(ctx context.Context, record slog.Record) error {
	if record.Level == slog.LevelError {
		record.Attrs(func(attr slog.Attr) bool {
			if attr.Key == "error" {
				if e, ok := attr.Value.Any().(error); ok && e != nil {
					h.err = e
					return false
				}
			}
			return true
		})

		if h.err != nil {
			for _, f := range h.f {
				if match := f(h.err); match {
					record.Level = slog.LevelWarn
					return h.Handler.Handle(ctx, record)
				}
			}
		}
	}

	return h.Handler.Handle(ctx, record)
}

// SetErrorWarnHandler will wrap the default slog.Logger's handler with an ErrorWarnHandler.
func SetErrorWarnHandler(funcs ...ErrorCheckFunction) {
	handler := NewErrorWarnHandler(slog.Default().Handler(), funcs...)
	slog.SetDefault(slog.New(handler))
}
