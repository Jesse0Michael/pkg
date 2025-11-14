package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareDefault(t *testing.T) {
	tests := []struct {
		name string
		next http.HandlerFunc
	}{
		{
			next: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"message": "Success"}`))
			},
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			Default(tt.next).ServeHTTP(w, req)

			if w.Header().Get("X-Content-Type-Options") != "nosniff" {
				t.Errorf(
					"X-Content-Type-Options header should match\n\tExpected: nosniff\n\tReceived: %s",
					w.Header().Get("X-Content-Type-Options"),
				)
			}

			if w.Header().Get("Connection") != "keep-alive" {
				t.Errorf("Connection header should match\n\tExpected: keep-alive\n\tReceived: %s", w.Header().Get("Connection"))
			}

			if w.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
				t.Errorf("Content-Type header should match\n\tExpected: text/plain; charset=utf-8\n\tReceived: %s", w.Header().Get("Content-Type"))
			}

			expectedBody := `{"message": "Success"}`

			if w.Body.String() != expectedBody {
				t.Errorf("Body should match\n\tExpected: %s\n\tReceived: %s", expectedBody, w.Body.String())
			}
		})
	}
}
