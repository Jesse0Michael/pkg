package middleware

import (
	"context"
	"net/http"

	"github.com/jesse0michael/pkg/http/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Authenticator interface {
	Authenticate(r *http.Request) (bool, context.Context)
}

func Auth(a Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorized, ctx := a.Authenticate(r)
			if !authorized {
				w.WriteHeader(http.StatusUnauthorized)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"errors":[{"message":"unauthorized"}]}`))
				return
			}
			span := trace.SpanFromContext(r.Context())
			if sub, ok := auth.Subject(ctx); ok {
				span.SetAttributes(attribute.String("subject", sub))
			}
			if admin, ok := auth.Admin(ctx); ok {
				span.SetAttributes(attribute.Bool("admin", admin))
			}
			ctx = context.WithValue(ctx, auth.RequestContextKey, span.SpanContext().TraceID().String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RejectReadOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		readOnly, _ := auth.ReadOnly(r.Context())
		if readOnly {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"errors":[{"message":"unauthorized"}]}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}
