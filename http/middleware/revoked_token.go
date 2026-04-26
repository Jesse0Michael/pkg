package middleware

import (
	"log/slog"
	"net/http"

	"github.com/jesse0michael/pkg/auth"
)

// RevokedToken returns HTTP middleware that rejects requests whose JWT (by JTI)
// or subject has been revoked according to the provided checker.
func RevokedToken(checker auth.RevokedTokenChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			jti, ok := auth.JTI(r.Context())
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"errors":[{"message":"unauthorized"}]}`))
				return
			}

			revoked, err := checker.IsRevoked(r.Context(), jti)
			if err != nil {
				slog.ErrorContext(r.Context(), "failed to check token revocation", "err", err, "jti", jti)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			if revoked {
				slog.WarnContext(r.Context(), "revoked token rejected", "jti", jti)
				w.WriteHeader(http.StatusUnauthorized)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"errors":[{"message":"unauthorized"}]}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
