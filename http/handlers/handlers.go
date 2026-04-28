package handlers

import (
	"context"
	"net/http"
	"net/http/pprof"
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

// HandlePprof returns a handler that serves runtime profiling data.
func HandlePprof() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", pprof.Index)
	mux.HandleFunc("/cmdline", pprof.Cmdline)
	mux.HandleFunc("/profile", pprof.Profile)
	mux.HandleFunc("/symbol", pprof.Symbol)
	mux.HandleFunc("/trace", pprof.Trace)
	return mux
}

// HealthChecker reports whether a dependency is healthy.
type HealthChecker interface {
	Healthy(ctx context.Context) error
}

// HandleHealth returns a health check handler. When called with no checkers it
// always returns 200. When checkers are provided, each is called and any
// failure results in a 503.
func HandleHealth(checkers ...HealthChecker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		for _, c := range checkers {
			if err := c.Healthy(r.Context()); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte(`{"message": "Health Unavailable"}`))
				return
			}
		}
		_, _ = w.Write([]byte(`{"message": "Health OK"}`))
	})
}

// ServeHealthCheckMetrics serves a health check and metrics endpoint from port 9999
// It runs in a goroutine and shuts down when the context is done
func ServeHealthCheckMetrics(ctx context.Context) {
	mux := http.NewServeMux()
	mux.Handle("/health", HandleHealth())
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:              ":9999",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		_ = server.ListenAndServe()
	}()

	<-ctx.Done()
	_ = server.Shutdown(context.Background())
}

// HandleWithMiddleware allows you to specify a HTTP handler that is to used with a set of middleware functions
func HandleWithMiddleware(handler http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	slices.Reverse(middleware)
	for _, mw := range middleware {
		handler = mw(handler)
	}

	return handler
}
