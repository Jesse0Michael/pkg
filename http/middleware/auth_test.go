package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jesse0michael/pkg/auth"
)

type MockAuthenticator struct {
	authorized bool
}

func (d *MockAuthenticator) Authenticate(r *http.Request) (bool, context.Context) {
	return d.authorized, r.Context()
}

func TestAuth(t *testing.T) {
	ctx := context.WithValue(context.TODO(), auth.SubjectContextKey, "auth")
	ctx = context.WithValue(ctx, auth.AdminContextKey, true)
	tests := []struct {
		name          string
		ctx           context.Context
		next          http.HandlerFunc
		authenticator Authenticator
		expectedBody  string
		expectedCode  int
	}{
		{
			name: "auth success",
			ctx:  context.TODO(),
			next: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"message": "Success"}`))
			},
			authenticator: &MockAuthenticator{authorized: true},
			expectedBody:  `{"message": "Success"}`,
			expectedCode:  http.StatusOK,
		},
		{
			name: "auth success - with auth context",
			ctx:  ctx,
			next: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"message": "Success"}`))
			},
			authenticator: &MockAuthenticator{authorized: true},
			expectedBody:  `{"message": "Success"}`,
			expectedCode:  http.StatusOK,
		},
		{
			name: "auth unauthorized",
			ctx:  context.TODO(),
			next: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"message": "Success"}`))
			},
			authenticator: &MockAuthenticator{authorized: false},
			expectedBody:  `{"errors":[{"message":"forbidden"}]}`,
			expectedCode:  http.StatusForbidden,
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil).WithContext(tt.ctx)
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
			ctx:  context.WithValue(context.TODO(), auth.ReadOnlyContextKey, false),
			next: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"message": "Success"}`))
			},
			expectedBody: `{"message": "Success"}`,
			expectedCode: http.StatusOK,
		},
		{
			name: "missing read only",
			ctx:  context.TODO(),
			next: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"message": "Success"}`))
			},
			expectedBody: `{"message": "Success"}`,
			expectedCode: http.StatusOK,
		},
		{
			name: "read only rejected",
			ctx:  context.WithValue(context.TODO(), auth.ReadOnlyContextKey, true),
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
