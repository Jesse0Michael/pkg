package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TagParams adds any query parameters as attributes to the span
func TagParams(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())
		values := r.URL.Query()
		for k := range values {
			span.SetAttributes(attribute.String(k, values.Get(k)))
		}
		if r.Header.Get("Content-Length") != "" {
			span.SetAttributes(attribute.String("Content-Length", r.Header.Get("Content-Length")))
		}
		if r.Header.Get("Accept-Language") != "" {
			span.SetAttributes(attribute.String("Accept-Language", r.Header.Get("Accept-Language")))
		}

		next.ServeHTTP(w, r)
	})
}
