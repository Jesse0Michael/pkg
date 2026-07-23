package oidc

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"testing/synctest"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGoogleAuth_Verify(t *testing.T) {
	tests := []struct {
		name      string
		clientIDs []string
		token     func(issuer string, key *rsa.PrivateKey) string
		wantErr   bool
		want      *Claims
	}{
		{
			name:      "valid token",
			clientIDs: []string{"test-client"},
			token: func(issuer string, key *rsa.PrivateKey) string {
				return signTestToken(t, key, testTokenClaims{
					Issuer:   issuer,
					Subject:  "test-subject",
					Audience: []string{"test-client"},
					Email:    "test@example.com",
				})
			},
			want: &Claims{
				Provider:      "google",
				Subject:       "test-subject",
				Email:         "test@example.com",
				EmailVerified: true,
				Name:          "test-name",
				Picture:       "test-picture",
			},
		},
		{
			name:      "wrong audience",
			clientIDs: []string{"test-client"},
			token: func(issuer string, key *rsa.PrivateKey) string {
				return signTestToken(t, key, testTokenClaims{
					Issuer:   issuer,
					Subject:  "test-subject",
					Audience: []string{"test-wrong-client"},
					Email:    "test@example.com",
				})
			},
			wantErr: true,
		},
		{
			name:      "multiple accepted clients",
			clientIDs: []string{"test-client-1", "test-client-2"},
			token: func(issuer string, key *rsa.PrivateKey) string {
				return signTestToken(t, key, testTokenClaims{
					Issuer:   issuer,
					Subject:  "test-subject",
					Audience: []string{"test-client-2"},
					Email:    "test@example.com",
				})
			},
			want: &Claims{
				Provider:      "google",
				Subject:       "test-subject",
				Email:         "test@example.com",
				EmailVerified: true,
				Name:          "test-name",
				Picture:       "test-picture",
			},
		},
		{
			name:      "invalid token",
			clientIDs: []string{"test-client"},
			token: func(_ string, _ *rsa.PrivateKey) string {
				return "test-invalid-token"
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				key, err := rsa.GenerateKey(rand.Reader, 2048)
				if err != nil {
					t.Fatalf("failed to generate RSA key: %v", err)
				}

				srv := newTestOIDCServer(t, key)
				defer srv.Close()

				ga, err := NewGoogleAuth(t.Context(), GoogleAuthConfig{
					Issuer:    srv.URL,
					ClientIDs: tt.clientIDs,
				})
				if err != nil {
					t.Fatalf("failed to create GoogleAuth: %v", err)
				}

				claims, err := ga.Verify(t.Context(), tt.token(srv.URL, key))
				if (err != nil) != tt.wantErr {
					t.Fatalf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				}
				if !reflect.DeepEqual(claims, tt.want) {
					t.Errorf("Verify() = %+v, want %+v", claims, tt.want)
				}
			})
		})
	}
}

type testTokenClaims struct {
	Issuer   string
	Subject  string
	Audience []string
	Email    string
}

func signTestToken(t *testing.T, key *rsa.PrivateKey, claims testTokenClaims) string {
	t.Helper()
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":            claims.Issuer,
		"sub":            claims.Subject,
		"aud":            claims.Audience,
		"iat":            now.Unix(),
		"exp":            now.Add(time.Hour).Unix(),
		"email":          claims.Email,
		"email_verified": true,
		"name":           "test-name",
		"picture":        "test-picture",
	})
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

func newTestOIDCServer(t *testing.T, key *rsa.PrivateKey) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	var srv *httptest.Server
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"issuer":   srv.URL,
			"jwks_uri": fmt.Sprintf("%s/jwks", srv.URL),
		})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rsaPublicKeyJWKS(&key.PublicKey))
	})

	srv = httptest.NewServer(mux)
	return srv
}

func rsaPublicKeyJWKS(pub *rsa.PublicKey) map[string]any {
	return map[string]any{
		"keys": []map[string]string{{
			"kty": "RSA",
			"alg": "RS256",
			"use": "sig",
			"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
			"e":   base64.RawURLEncoding.EncodeToString(new(big.Int).SetInt64(int64(pub.E)).Bytes()),
		}},
	}
}
