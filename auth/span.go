package auth

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// WithSpan sets auth attributes on the current span and stores the trace ID in context
func WithSpan(ctx context.Context) context.Context {
	span := trace.SpanFromContext(ctx)
	if sub, ok := Subject(ctx); ok {
		span.SetAttributes(attribute.String("subject", sub))
	}
	if admin, ok := Admin(ctx); ok {
		span.SetAttributes(attribute.Bool("admin", admin))
	}
	if readOnly, ok := ReadOnly(ctx); ok {
		span.SetAttributes(attribute.Bool("readOnly", readOnly))
	}
	return context.WithValue(ctx, RequestContextKey, span.SpanContext().TraceID().String())
}
