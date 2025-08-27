package middleware

import (
	"context"
	"net/http"

	"github.com/jesse0michael/pkg/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Authenticator interface {
	Authenticate(r *http.Request) (bool, context.Context)
}

func Auth(a Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ok, ctx := a.Authenticate(r)
			if !ok {
				w.WriteHeader(http.StatusForbidden)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"errors":[{"message":"forbidden"}]}`))
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
