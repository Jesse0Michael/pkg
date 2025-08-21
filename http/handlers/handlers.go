package handlers

import (
	"net/http"
	"slices"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HandleNotFound returns a NotFound HTTP handler
func HandleNotFound() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"page not found"}]}`))
	})
}

// HandleNotAllowed returns a NotAllowed HTTP handler
func HandleNotAllowed() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte(`{"errors":[{"message":"method not allowed"}]}`))
	})
}

// HandleHealth is a basic HTTP handler to use as a health check
func HandleHealth() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message": "Health OK"}`))
	})
}

// ServeHealthCheckMetrics serves a health check and metrics endpoint from port 9999
func ServeHealthCheckMetrics() {
	http.Handle("/health", HandleHealth())
	http.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:              ":9999",
		ReadHeaderTimeout: 5 * time.Second,
	}

	_ = server.ListenAndServe()
}

// HandleWithMiddleware allows you to specify a HTTP handler that is to used with a set of middleware functions
func HandleWithMiddleware(handler http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	slices.Reverse(middleware)
	for _, mw := range middleware {
		handler = mw(handler)
	}

	return handler
}
