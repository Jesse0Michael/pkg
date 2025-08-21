package middleware

import (
	"net/http"
)

// Default sets the HTTP response headers applicable to the any route by default.
func Default(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Content-Type", "application/json")

		next.ServeHTTP(w, r)
	})
}
