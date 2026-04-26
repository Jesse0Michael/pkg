package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jesse0michael/pkg/auth"
)

type mockRevokedTokenChecker struct {
	revoked bool
	err     error
}

func (m *mockRevokedTokenChecker) IsRevoked(_ context.Context, _ string) (bool, error) {
	return m.revoked, m.err
}

func TestRevokedToken(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		checker      auth.RevokedTokenChecker
		expectedBody string
		expectedCode int
	}{
		{
			name:         "not revoked",
			ctx:          context.WithValue(t.Context(), auth.JTIContextKey, "test-jti"),
			checker:      &mockRevokedTokenChecker{revoked: false},
			expectedBody: `{"message": "Success"}`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "revoked token",
			ctx:          context.WithValue(t.Context(), auth.JTIContextKey, "test-jti"),
			checker:      &mockRevokedTokenChecker{revoked: true},
			expectedBody: `{"errors":[{"message":"unauthorized"}]}`,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "checker error",
			ctx:          context.WithValue(t.Context(), auth.JTIContextKey, "test-jti"),
			checker:      &mockRevokedTokenChecker{err: errors.New("test-error")},
			expectedBody: "internal server error\n",
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "no jti in context",
			ctx:          t.Context(),
			checker:      &mockRevokedTokenChecker{revoked: true},
			expectedBody: `{"errors":[{"message":"unauthorized"}]}`,
			expectedCode: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil).WithContext(tt.ctx)

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"message": "Success"}`))
			})

			RevokedToken(tt.checker)(next).ServeHTTP(w, req)

			if w.Body.String() != tt.expectedBody {
				t.Errorf("Body should match\n\tExpected: %s\n\tReceived: %s", tt.expectedBody, w.Body.String())
			}

			if w.Code != tt.expectedCode {
				t.Errorf("Response code should match\n\tExpected: %d\n\tReceived: %d", tt.expectedCode, w.Code)
			}
		})
	}
}
