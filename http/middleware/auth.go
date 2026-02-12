package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/jesse0michael/pkg/auth"
)

type Authenticator interface {
	VerifyAccessToken(token string) (*auth.Claim, error)
}

func Auth(a Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, ok := authenticate(r, a)
			if !ok {
				w.WriteHeader(http.StatusForbidden)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"errors":[{"message":"forbidden"}]}`))
				return
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func authenticate(r *http.Request, a Authenticator) (context.Context, bool) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	claims, err := a.VerifyAccessToken(token)
	if err != nil {
		return r.Context(), false
	}

	ctx := context.WithValue(r.Context(), auth.AuthorizationContextKey, token)
	ctx = auth.WithClaims(ctx, claims)
	ctx = auth.WithSpan(ctx)
	return ctx, true
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
