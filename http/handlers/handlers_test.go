package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHandleNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://www.example.com/invalid-page", nil)

	HandleNotFound().ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusNotFound {
		t.Errorf("HTTP status code should match\n\tExpected: %d\n\tReceived: %d", http.StatusNotFound, result.StatusCode)
	}

	expected := `{"errors":[{"message":"page not found"}]}`
	if w.Body.String() != expected {
		t.Errorf("Body should match\n\tExpected: %s\n\tReceived: %s", expected, w.Body.String())
	}
}

func TestHandleNotAllowed(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://www.example.com/invalid-request-method", nil)

	HandleNotAllowed().ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf(
			"HTTP status code should match\n\tExpected: %d\n\tReceived: %d",
			http.StatusMethodNotAllowed,
			result.StatusCode,
		)
	}

	expected := `{"errors":[{"message":"method not allowed"}]}`
	if w.Body.String() != expected {
		t.Errorf("Body should match\n\tExpected: %s\n\tReceived: %s", expected, w.Body.String())
	}
}

func Test_HandleHealth(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)

	HandleHealth().ServeHTTP(w, req)

	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusOK {
		t.Errorf(
			"HTTP status code should match\n\tExpected: %d\n\tReceived: %d",
			http.StatusMethodNotAllowed,
			result.StatusCode,
		)
	}

	expected := `{"message": "Health OK"}`
	if w.Body.String() != expected {
		t.Errorf("Body should match\n\tExpected: %s\n\tReceived: %s", expected, w.Body.String())
	}
}

func TestServeHealthCheckMetrics(t *testing.T) {
	go ServeHealthCheckMetrics(t.Context())

	resp, err := http.Get("http://localhost:9999/health")
	if err != nil {
		t.Errorf("Should not fail to make health request: %s", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("/health returned status code: %d expected: %d", resp.StatusCode, http.StatusOK)
	}
	b, _ := io.ReadAll(resp.Body)
	if string(b) != `{"message": "Health OK"}` {
		t.Errorf("/health body did not match expected: %s", string(b))
	}
	resp.Body.Close()

	resp, err = http.Get("http://localhost:9999/metrics")
	if err != nil {
		t.Errorf("Should not fail to make metrics request: %s", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("/metrics returned status code: %d expected: %d", resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()
}

func TestHandleWithMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte("."))
	})
	tests := []struct {
		name       string
		middleware []func(http.Handler) http.Handler
		wantBody   string
	}{
		{
			name:       "without middleware",
			middleware: []func(http.Handler) http.Handler{},
			wantBody:   ".",
		},
		{
			name: "with middleware",
			middleware: []func(http.Handler) http.Handler{
				func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						_, _ = w.Write([]byte("man in the middle"))
						next.ServeHTTP(w, r)
					})
				},
			},
			wantBody: "man in the middle.",
		},
		{
			name: "with many middlewares",
			middleware: []func(http.Handler) http.Handler{
				func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						_, _ = w.Write([]byte("I'm looking at the "))
						next.ServeHTTP(w, r)
					})
				},
				func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						_, _ = w.Write([]byte("man in the middle"))
						next.ServeHTTP(w, r)
					})
				},
			},
			wantBody: "I'm looking at the man in the middle.",
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "http://www.example.com/", nil)

			wrappedHandler := HandleWithMiddleware(handler, tt.middleware...)

			wrappedHandler.ServeHTTP(w, req)

			body, _ := io.ReadAll(w.Body)
			if string(body) != tt.wantBody {
				t.Errorf("Response body should match\n\tExpected: %s\n\tReceived: %s", tt.wantBody, string(body))
			}
		})
	}
}
