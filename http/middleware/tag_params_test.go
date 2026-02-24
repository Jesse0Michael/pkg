package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
)

func TestTagParams(t *testing.T) {
	tests := []struct {
		name    string
		span    bool
		headers map[string]string
	}{
		{
			name: "no transaction",
			span: false,
		},
		{
			name: "tag transaction",
			span: true,
		},
		{
			name: "tag content-length",
			span: true,
			headers: map[string]string{
				"Content-Length": "192",
			},
		},
		{
			name: "tag accept-language",
			span: true,
			headers: map[string]string{
				"Accept-Language": "fr",
			},
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			if tt.span {
				ctx, _ = otel.Tracer("test").Start(t.Context(), "test")
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/?test=best", nil).WithContext(ctx)
			for k, v := range tt.headers {
				req.Header.Add(k, v)
			}
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			TagParams(next).ServeHTTP(w, req)
		})
	}
}
