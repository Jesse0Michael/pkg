package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jesse0michael/pkg/auth"
	"github.com/jesse0michael/pkg/auth/oidc"
)

type mockVerifier struct {
	claims *oidc.Claims
	err    error
}

func (m *mockVerifier) Verify(_ context.Context, _ string) (*oidc.Claims, error) {
	return m.claims, m.err
}

func TestHandleOIDCLogin(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		verifier     oidc.Verifier
		resolve      OIDCResolver
		expectedCode int
		expectedBody string
	}{
		{
			name:         "invalid request body",
			body:         `{invalid`,
			verifier:     &mockVerifier{},
			resolve:      func(context.Context, *oidc.Claims) ([]auth.TokenOption, error) { return nil, nil },
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "missing id_token",
			body:         `{}`,
			verifier:     &mockVerifier{},
			resolve:      func(context.Context, *oidc.Claims) ([]auth.TokenOption, error) { return nil, nil },
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errors":[{"message":"id_token is required"}]}`,
		},
		{
			name: "invalid id token",
			body: `{"id_token":"test-token"}`,
			verifier: &mockVerifier{
				err: fmt.Errorf("test-error"),
			},
			resolve:      func(context.Context, *oidc.Claims) ([]auth.TokenOption, error) { return nil, nil },
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"errors":[{"message":"invalid id token"}]}`,
		},
		{
			name: "resolver error",
			body: `{"id_token":"test-token"}`,
			verifier: &mockVerifier{
				claims: &oidc.Claims{Provider: "google", Subject: "test-subject", EmailVerified: true},
			},
			resolve: func(context.Context, *oidc.Claims) ([]auth.TokenOption, error) {
				return nil, fmt.Errorf("test-error")
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"errors":[{"message":"Internal Server Error"}]}`,
		},
		{
			name: "success",
			body: `{"id_token":"test-token"}`,
			verifier: &mockVerifier{
				claims: &oidc.Claims{Provider: "google", Subject: "test-subject", EmailVerified: true},
			},
			resolve: func(_ context.Context, c *oidc.Claims) ([]auth.TokenOption, error) {
				return []auth.TokenOption{auth.WithSubject(c.Subject)}, nil
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"subject":"test-subject","access_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0eXBlIjoiYWNjZXNzIiwic3ViIjoidGVzdC1zdWJqZWN0IiwiZXhwIjo5NDY2ODg0MDAsImlhdCI6OTQ2Njg0ODAwLCJqdGkiOiIwMTAyMDMwNC0wNTA2LTQ3MDgtODkwYS0wYjBjMGQwZTBmMTAifQ.lVYn7KmPQULeQbDTqY58c2Z9fjLo-7o7aDTDnOr4JgY","refresh_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0eXBlIjoicmVmcmVzaCIsInN1YiI6InRlc3Qtc3ViamVjdCIsImV4cCI6OTQ2Njg4NDAwLCJpYXQiOjk0NjY4NDgwMCwianRpIjoiMDEwMjAzMDQtMDUwNi00NzA4LTg5MGEtMGIwYzBkMGUwZjEwIn0.OCCT-HOEc1AzYLhdbz8rutGAFF2BYzcxbLBDpVYIMf4"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				uuid.SetRand(bytes.NewReader(bytes.Repeat([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, 8)))
				defer uuid.SetRand(nil)

				jwtAuth := auth.NewJWTAuth(auth.Config{SecretKey: []byte("test-secret"), AccessTokenTTL: time.Hour, RefreshTokenTTL: time.Hour}, jwt.SigningMethodHS256)
				handler := HandleOIDCLogin(tt.verifier, jwtAuth, tt.resolve)

				w := httptest.NewRecorder()
				r := httptest.NewRequest(http.MethodPost, "/auth/google", strings.NewReader(tt.body))
				handler.ServeHTTP(w, r)

				if w.Code != tt.expectedCode {
					t.Errorf("status code = %d, want %d", w.Code, tt.expectedCode)
				}
				if tt.expectedBody != "" {
					got := strings.TrimSpace(w.Body.String())
					if got != tt.expectedBody {
						t.Errorf("body = %s, want %s", got, tt.expectedBody)
					}
				}
			})
		})
	}
}

func TestHandleLogin(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		resolve      IdentityResolver
		expectedCode int
		expectedBody string
	}{
		{
			name:         "invalid request body",
			body:         `{invalid`,
			resolve:      func(context.Context, string, string) ([]auth.TokenOption, error) { return nil, nil },
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "missing identity",
			body:         `{"password":"test-password"}`,
			resolve:      func(context.Context, string, string) ([]auth.TokenOption, error) { return nil, nil },
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errors":[{"message":"identity and password are required"}]}`,
		},
		{
			name:         "missing password",
			body:         `{"identity":"test-identity"}`,
			resolve:      func(context.Context, string, string) ([]auth.TokenOption, error) { return nil, nil },
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errors":[{"message":"identity and password are required"}]}`,
		},
		{
			name: "resolver error",
			body: `{"identity":"test-identity","password":"test-password"}`,
			resolve: func(context.Context, string, string) ([]auth.TokenOption, error) {
				return nil, fmt.Errorf("test-error")
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"errors":[{"message":"Internal Server Error"}]}`,
		},
		{
			name: "success",
			body: `{"identity":"test-identity","password":"test-password"}`,
			resolve: func(_ context.Context, _, _ string) ([]auth.TokenOption, error) {
				return []auth.TokenOption{auth.WithSubject("test-subject")}, nil
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"subject":"test-subject","access_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0eXBlIjoiYWNjZXNzIiwic3ViIjoidGVzdC1zdWJqZWN0IiwiZXhwIjo5NDY2ODg0MDAsImlhdCI6OTQ2Njg0ODAwLCJqdGkiOiIwMTAyMDMwNC0wNTA2LTQ3MDgtODkwYS0wYjBjMGQwZTBmMTAifQ.lVYn7KmPQULeQbDTqY58c2Z9fjLo-7o7aDTDnOr4JgY","refresh_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0eXBlIjoicmVmcmVzaCIsInN1YiI6InRlc3Qtc3ViamVjdCIsImV4cCI6OTQ2Njg4NDAwLCJpYXQiOjk0NjY4NDgwMCwianRpIjoiMDEwMjAzMDQtMDUwNi00NzA4LTg5MGEtMGIwYzBkMGUwZjEwIn0.OCCT-HOEc1AzYLhdbz8rutGAFF2BYzcxbLBDpVYIMf4"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				uuid.SetRand(bytes.NewReader(bytes.Repeat([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, 8)))
				defer uuid.SetRand(nil)

				jwtAuth := auth.NewJWTAuth(auth.Config{SecretKey: []byte("test-secret"), AccessTokenTTL: time.Hour, RefreshTokenTTL: time.Hour}, jwt.SigningMethodHS256)
				handler := HandleLogin(jwtAuth, tt.resolve)

				w := httptest.NewRecorder()
				r := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(tt.body))
				handler.ServeHTTP(w, r)

				if w.Code != tt.expectedCode {
					t.Errorf("status code = %d, want %d", w.Code, tt.expectedCode)
				}
				if tt.expectedBody != "" {
					got := strings.TrimSpace(w.Body.String())
					if got != tt.expectedBody {
						t.Errorf("body = %s, want %s", got, tt.expectedBody)
					}
				}
			})
		})
	}
}

func TestHandleRefreshTokens(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		genToken     bool
		expectedCode int
		expectedBody string
	}{
		{
			name:         "invalid request body",
			body:         `{invalid`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "missing refresh_token",
			body:         `{}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errors":[{"message":"refresh_token is required"}]}`,
		},
		{
			name:         "invalid refresh token",
			body:         `{"refresh_token":"test-invalid-token"}`,
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"errors":[{"message":"invalid refresh token"}]}`,
		},
		{
			name:         "success",
			genToken:     true,
			expectedCode: http.StatusOK,
			expectedBody: `{"subject":"test-subject","access_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0eXBlIjoiYWNjZXNzIiwic3ViIjoidGVzdC1zdWJqZWN0IiwiZXhwIjo5NDY2ODg0MDAsImlhdCI6OTQ2Njg0ODAwLCJqdGkiOiIwMTAyMDMwNC0wNTA2LTQ3MDgtODkwYS0wYjBjMGQwZTBmMTAifQ.lVYn7KmPQULeQbDTqY58c2Z9fjLo-7o7aDTDnOr4JgY","refresh_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0eXBlIjoicmVmcmVzaCIsInN1YiI6InRlc3Qtc3ViamVjdCIsImV4cCI6OTQ2Njg4NDAwLCJpYXQiOjk0NjY4NDgwMCwianRpIjoiMDEwMjAzMDQtMDUwNi00NzA4LTg5MGEtMGIwYzBkMGUwZjEwIn0.OCCT-HOEc1AzYLhdbz8rutGAFF2BYzcxbLBDpVYIMf4"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				uuid.SetRand(bytes.NewReader(bytes.Repeat([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, 8)))
				defer uuid.SetRand(nil)

				jwtAuth := auth.NewJWTAuth(auth.Config{SecretKey: []byte("test-secret"), AccessTokenTTL: time.Hour, RefreshTokenTTL: time.Hour}, jwt.SigningMethodHS256)
				handler := HandleRefreshTokens(jwtAuth)

				body := tt.body
				if tt.genToken {
					_, refresh, _ := jwtAuth.GenerateTokens(auth.WithSubject("test-subject"))
					body = fmt.Sprintf(`{"refresh_token":%q}`, refresh)
				}

				w := httptest.NewRecorder()
				r := httptest.NewRequest(http.MethodPost, "/auth/refresh", strings.NewReader(body))
				handler.ServeHTTP(w, r)

				if w.Code != tt.expectedCode {
					t.Errorf("status code = %d, want %d", w.Code, tt.expectedCode)
				}
				if tt.expectedBody != "" {
					got := strings.TrimSpace(w.Body.String())
					if got != tt.expectedBody {
						t.Errorf("body = %s, want %s", got, tt.expectedBody)
					}
				}
			})
		})
	}
}
