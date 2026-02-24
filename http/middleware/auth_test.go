package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jesse0michael/pkg/auth"
)

type MockAuthenticator struct {
	claim *auth.Claim
	err   error
}

func (d *MockAuthenticator) VerifyAccessToken(token string) (*auth.Claim, error) {
	return d.claim, d.err
}

func TestAuth(t *testing.T) {
	tests := []struct {
		name          string
		next          http.HandlerFunc
		authenticator Authenticator
		expectedBody  string
		expectedCode  int
	}{
		{
			name: "auth success",
			next: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"message": "Success"}`))
			},
			authenticator: &MockAuthenticator{claim: &auth.Claim{}},
			expectedBody:  `{"message": "Success"}`,
			expectedCode:  http.StatusOK,
		},
		{
			name: "auth success - with claims",
			next: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"message": "Success"}`))
			},
			authenticator: &MockAuthenticator{claim: &auth.Claim{Admin: true}},
			expectedBody:  `{"message": "Success"}`,
			expectedCode:  http.StatusOK,
		},
		{
			name: "auth unauthorized",
			next: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"message": "Success"}`))
			},
			authenticator: &MockAuthenticator{err: fmt.Errorf("invalid")},
			expectedBody:  `{"errors":[{"message":"forbidden"}]}`,
			expectedCode:  http.StatusForbidden,
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			Auth(tt.authenticator)(tt.next).ServeHTTP(w, req)

			if w.Body.String() != tt.expectedBody {
				t.Errorf("Body should match\n\tExpected: %s\n\tReceived: %s", tt.expectedBody, w.Body.String())
			}

			if w.Code != tt.expectedCode {
				t.Errorf("Response code should match\n\tExpected: %d\n\tReceived: %d", tt.expectedCode, w.Code)
			}
		})
	}
}

func TestRejectReadOnly(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		next         http.HandlerFunc
		expectedBody string
		expectedCode int
	}{
		{
			name: "not read only",
			ctx:  context.WithValue(t.Context(), auth.ReadOnlyContextKey, false),
			next: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"message": "Success"}`))
			},
			expectedBody: `{"message": "Success"}`,
			expectedCode: http.StatusOK,
		},
		{
			name: "missing read only",
			ctx:  t.Context(),
			next: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"message": "Success"}`))
			},
			expectedBody: `{"message": "Success"}`,
			expectedCode: http.StatusOK,
		},
		{
			name: "read only rejected",
			ctx:  context.WithValue(t.Context(), auth.ReadOnlyContextKey, true),
			next: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"message": "Success"}`))
			},
			expectedBody: `{"errors":[{"message":"unauthorized"}]}`,
			expectedCode: http.StatusUnauthorized,
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil).WithContext(tt.ctx)
			RejectReadOnly(tt.next).ServeHTTP(w, req)

			if w.Body.String() != tt.expectedBody {
				t.Errorf("Body should match\n\tExpected: %s\n\tReceived: %s", tt.expectedBody, w.Body.String())
			}

			if w.Code != tt.expectedCode {
				t.Errorf("Response code should match\n\tExpected: %d\n\tReceived: %d", tt.expectedCode, w.Code)
			}
		})
	}
}
