package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestNewAuthService(t *testing.T) {
	cfg := Config{
		SecretKey:       []byte("test-secret"),
		Issuer:          "test-issuer",
		AccessTokenTTL:  time.Minute * 15,
		RefreshTokenTTL: time.Hour * 24 * 7,
	}

	svc := NewJWTAuth(cfg, jwt.SigningMethodHS256)

	if svc == nil {
		t.Fatal("service should not be nil")
	}
	if string(svc.cfg.SecretKey) != string(cfg.SecretKey) {
		t.Errorf("secret key = %q, want %q", svc.cfg.SecretKey, cfg.SecretKey)
	}
	if svc.cfg.Issuer != cfg.Issuer {
		t.Errorf("issuer = %q, want %q", svc.cfg.Issuer, cfg.Issuer)
	}
	if svc.cfg.AccessTokenTTL != cfg.AccessTokenTTL {
		t.Errorf("access token TTL = %v, want %v", svc.cfg.AccessTokenTTL, cfg.AccessTokenTTL)
	}
	if svc.cfg.RefreshTokenTTL != cfg.RefreshTokenTTL {
		t.Errorf("refresh token TTL = %v, want %v", svc.cfg.RefreshTokenTTL, cfg.RefreshTokenTTL)
	}
}

func TestGenerateTokens(t *testing.T) {
	tests := []struct {
		name    string
		options []TokenOption
	}{
		{
			name: "empty options",
		},
		{
			name:    "with subject",
			options: []TokenOption{WithSubject("test-subject")},
		},
		{
			name:    "with audience",
			options: []TokenOption{WithAudience("test-audience-1", "test-audience-2")},
		},
		{
			name:    "with admin",
			options: []TokenOption{WithAdmin()},
		},
		{
			name:    "with read only",
			options: []TokenOption{WithReadOnly()},
		},
		{
			name: "with all options",
			options: []TokenOption{
				WithSubject("test-subject"),
				WithAudience("test-audience"),
				WithAdmin(),
				WithReadOnly(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				SecretKey:       []byte("test-secret"),
				Issuer:          "test-issuer",
				AccessTokenTTL:  time.Minute * 15,
				RefreshTokenTTL: time.Hour * 24 * 7,
			}

			svc := NewJWTAuth(cfg, jwt.SigningMethodHS256)
			accessToken, refreshToken, err := svc.GenerateTokens(tt.options...)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if accessToken == "" {
				t.Fatal("access token should not be empty")
			}
			if refreshToken == "" {
				t.Fatal("refresh token should not be empty")
			}

			// Validate token structure (should be 3 parts separated by dots)
			if parts := strings.Split(accessToken, "."); len(parts) != 3 {
				t.Errorf("access token parts = %d, want 3", len(parts))
			}
			if parts := strings.Split(refreshToken, "."); len(parts) != 3 {
				t.Errorf("refresh token parts = %d, want 3", len(parts))
			}
		})
	}
}

func TestVerifyAccessToken(t *testing.T) {
	userID := uuid.NewString()
	signingKey := []byte("test-secret")

	cfg := Config{
		SecretKey:       signingKey,
		Issuer:          "test-issuer",
		AccessTokenTTL:  time.Minute * 15,
		RefreshTokenTTL: time.Hour * 24 * 7,
	}

	svc := NewJWTAuth(cfg, jwt.SigningMethodHS256)
	accessToken, _, err := svc.GenerateTokens(WithSubject(userID))
	if err != nil {
		t.Fatalf("GenerateTokens: %v", err)
	}

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		checkUser bool
	}{
		{
			name:      "valid token",
			token:     accessToken,
			wantErr:   false,
			checkUser: true,
		},
		{
			name:      "invalid token format",
			token:     "invalid.token.format",
			wantErr:   true,
			checkUser: false,
		},
		{
			name:      "empty token",
			token:     "",
			wantErr:   true,
			checkUser: false,
		},
		{
			name:      "tampered token",
			token:     accessToken + "tampered",
			wantErr:   true,
			checkUser: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := svc.VerifyToken(tt.token, AccessTokenType)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if claims == nil {
				t.Fatal("claims should not be nil")
			}

			if tt.checkUser && claims.Subject != userID {
				t.Errorf("subject = %q, want %q", claims.Subject, userID)
			}
		})
	}
}

func TestVerifyRefreshToken(t *testing.T) {
	userID := uuid.NewString()
	signingKey := []byte("test-secret")

	cfg := Config{
		SecretKey:       signingKey,
		Issuer:          "test-issuer",
		AccessTokenTTL:  time.Minute * 15,
		RefreshTokenTTL: time.Hour * 24 * 7,
	}

	svc := NewJWTAuth(cfg, jwt.SigningMethodHS256)
	_, refreshToken, err := svc.GenerateTokens(WithSubject(userID))
	if err != nil {
		t.Fatalf("GenerateTokens: %v", err)
	}

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		checkUser bool
	}{
		{
			name:      "valid token",
			token:     refreshToken,
			wantErr:   false,
			checkUser: true,
		},
		{
			name:      "invalid token format",
			token:     "invalid.token.format",
			wantErr:   true,
			checkUser: false,
		},
		{
			name:      "empty token",
			token:     "",
			wantErr:   true,
			checkUser: false,
		},
		{
			name:      "tampered token",
			token:     refreshToken + "tampered",
			wantErr:   true,
			checkUser: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := svc.VerifyToken(tt.token, RefreshTokenType)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if claims == nil {
				t.Fatal("claims should not be nil")
			}

			if tt.checkUser && claims.Subject != userID {
				t.Errorf("subject = %q, want %q", claims.Subject, userID)
			}
		})
	}
}

func TestRefreshTokens(t *testing.T) {
	userID := uuid.NewString()
	signingKey := []byte("test-secret")

	cfg := Config{
		SecretKey:       signingKey,
		Issuer:          "test-issuer",
		AccessTokenTTL:  time.Minute * 15,
		RefreshTokenTTL: time.Hour * 24 * 7,
	}

	svc := NewJWTAuth(cfg, jwt.SigningMethodHS256)
	_, refreshToken, err := svc.GenerateTokens(WithSubject(userID))
	if err != nil {
		t.Fatalf("GenerateTokens: %v", err)
	}

	tests := []struct {
		name            string
		refreshToken    string
		wantErr         bool
		verifyNewTokens bool
	}{
		{
			name:            "valid refresh",
			refreshToken:    refreshToken,
			wantErr:         false,
			verifyNewTokens: true,
		},
		{
			name:            "invalid token",
			refreshToken:    "invalid.token.format",
			wantErr:         true,
			verifyNewTokens: false,
		},
		{
			name:            "empty token",
			refreshToken:    "",
			wantErr:         true,
			verifyNewTokens: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newAccessToken, newRefreshToken, err := svc.RefreshTokens(tt.refreshToken)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if newAccessToken == "" {
				t.Fatal("new access token should not be empty")
			}
			if newRefreshToken == "" {
				t.Fatal("new refresh token should not be empty")
			}
			if newRefreshToken == refreshToken {
				t.Error("refresh token should be different")
			}

			if tt.verifyNewTokens {
				accessClaims, err := svc.VerifyToken(newAccessToken, AccessTokenType)
				if err != nil {
					t.Fatalf("VerifyToken access: %v", err)
				}
				if accessClaims.Subject != userID {
					t.Errorf("access subject = %q, want %q", accessClaims.Subject, userID)
				}

				refreshClaims, err := svc.VerifyToken(newRefreshToken, RefreshTokenType)
				if err != nil {
					t.Fatalf("VerifyToken refresh: %v", err)
				}
				if refreshClaims.Subject != userID {
					t.Errorf("refresh subject = %q, want %q", refreshClaims.Subject, userID)
				}
			}
		})
	}
}

func TestExpiredToken(t *testing.T) {
	userID := uuid.NewString()

	// Very short TTL so the token expires within the test.
	cfg := Config{
		SecretKey:       []byte("test-secret"),
		Issuer:          "test-issuer",
		AccessTokenTTL:  time.Millisecond * 10,
		RefreshTokenTTL: time.Hour * 24,
	}

	svc := NewJWTAuth(cfg, jwt.SigningMethodHS256)
	accessToken, _, err := svc.GenerateTokens(WithSubject(userID))
	if err != nil {
		t.Fatalf("GenerateTokens: %v", err)
	}

	time.Sleep(time.Millisecond * 15)

	_, err = svc.VerifyToken(accessToken, AccessTokenType)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "expired")
	}
}

func TestTokenTypeValidation(t *testing.T) {
	userID := uuid.NewString()
	cfg := Config{
		SecretKey:       []byte("test-secret"),
		Issuer:          "test-issuer",
		AccessTokenTTL:  time.Minute * 15,
		RefreshTokenTTL: time.Hour * 24 * 7,
	}

	svc := NewJWTAuth(cfg, jwt.SigningMethodHS256)
	accessToken, refreshToken, err := svc.GenerateTokens(WithSubject(userID))
	if err != nil {
		t.Fatalf("GenerateTokens: %v", err)
	}

	_, err = svc.VerifyToken(accessToken, RefreshTokenType)
	if err == nil {
		t.Fatal("expected error verifying access token as refresh, got nil")
	}
	if !strings.Contains(err.Error(), "expected refresh token") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "expected refresh token")
	}

	_, err = svc.VerifyToken(refreshToken, AccessTokenType)
	if err == nil {
		t.Fatal("expected error verifying refresh token as access, got nil")
	}
	if !strings.Contains(err.Error(), "expected access token") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "expected access token")
	}
}

func TestExtractUserIDFromToken(t *testing.T) {
	userID := uuid.NewString()
	cfg := Config{
		SecretKey:       []byte("test-secret"),
		Issuer:          "test-issuer",
		AccessTokenTTL:  time.Minute * 15,
		RefreshTokenTTL: time.Hour * 24 * 7,
	}

	svc := NewJWTAuth(cfg, jwt.SigningMethodHS256)
	accessToken, refreshToken, err := svc.GenerateTokens(WithSubject(userID))
	if err != nil {
		t.Fatalf("GenerateTokens: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid access token",
			token:   accessToken,
			want:    userID,
			wantErr: false,
		},
		{
			name:    "valid refresh token",
			token:   refreshToken,
			want:    userID,
			wantErr: false,
		},
		{
			name:    "invalid token",
			token:   "invalid.token.format",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var subject string
			var err error

			// Try as access token first, then as refresh token.
			if claims, verifyErr := svc.VerifyToken(tt.token, AccessTokenType); verifyErr == nil {
				subject = claims.Subject
			} else if claims, verifyErr := svc.VerifyToken(tt.token, RefreshTokenType); verifyErr == nil {
				subject = claims.Subject
			} else {
				err = verifyErr
			}

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if subject != tt.want {
				t.Errorf("subject = %q, want %q", subject, tt.want)
			}
		})
	}
}

func TestVerifyTokenWithInvalidSigningMethod(t *testing.T) {
	userID := uuid.NewString()
	cfg := Config{
		SecretKey:       []byte("test-secret"),
		Issuer:          "test-issuer",
		AccessTokenTTL:  time.Minute * 15,
		RefreshTokenTTL: time.Hour * 24 * 7,
	}

	svc := NewJWTAuth(cfg, jwt.SigningMethodHS256)
	accessToken, _, err := svc.GenerateTokens(WithSubject(userID))
	if err != nil {
		t.Fatalf("GenerateTokens: %v", err)
	}

	claims, err := svc.VerifyToken(accessToken, AccessTokenType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims == nil {
		t.Fatal("claims should not be nil")
	}
	if claims.TokenType != AccessTokenType {
		t.Errorf("token type = %q, want %q", claims.TokenType, AccessTokenType)
	}
}

func TestWithInvalidConfig(t *testing.T) {
	tests := []struct {
		name       string
		cfg        Config
		shouldWork bool
	}{
		{
			name: "empty signing secret",
			cfg: Config{
				SecretKey:       nil,
				Issuer:          "test-issuer",
				AccessTokenTTL:  time.Minute * 15,
				RefreshTokenTTL: time.Hour * 24,
			},
			shouldWork: false,
		},
		{
			name: "empty issuer",
			cfg: Config{
				SecretKey:       []byte("test-secret"),
				Issuer:          "",
				AccessTokenTTL:  time.Minute * 15,
				RefreshTokenTTL: time.Hour * 24,
			},
			shouldWork: true,
		},
		{
			name: "zero TTLs",
			cfg: Config{
				SecretKey:       []byte("test-secret"),
				Issuer:          "test-issuer",
				AccessTokenTTL:  0,
				RefreshTokenTTL: 0,
			},
			shouldWork: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewJWTAuth(tt.cfg, jwt.SigningMethodHS256)
			userID := uuid.NewString()
			accessToken, refreshToken, err := svc.GenerateTokens(WithSubject(userID))

			if !tt.shouldWork {
				if err == nil {
					if _, err = svc.VerifyToken(accessToken, AccessTokenType); err != nil {
						return
					}
					if _, err = svc.VerifyToken(refreshToken, RefreshTokenType); err != nil {
						return
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if accessToken == "" {
				t.Fatal("access token should not be empty")
			}
			if refreshToken == "" {
				t.Fatal("refresh token should not be empty")
			}

			accessClaims, err := svc.VerifyToken(accessToken, AccessTokenType)
			if err != nil {
				t.Fatalf("VerifyToken access: %v", err)
			}
			if accessClaims.Subject != userID {
				t.Errorf("access subject = %q, want %q", accessClaims.Subject, userID)
			}

			refreshClaims, err := svc.VerifyToken(refreshToken, RefreshTokenType)
			if err != nil {
				t.Fatalf("VerifyToken refresh: %v", err)
			}
			if refreshClaims.Subject != userID {
				t.Errorf("refresh subject = %q, want %q", refreshClaims.Subject, userID)
			}
		})
	}
}

func TestWithClaims(t *testing.T) {
	tests := []struct {
		name  string
		claim *Claim
	}{
		{
			name:  "empty claim",
			claim: &Claim{},
		},
		{
			name: "filled claim",
			claim: &Claim{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "test-subject",
					ID:      "test-jti",
				},
				Admin:    true,
				ReadOnly: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithClaims(t.Context(), tt.claim)

			subject, subjectOK := Subject(ctx)
			if subject != tt.claim.Subject {
				t.Errorf("subject = %q, want %q", subject, tt.claim.Subject)
			}
			if subjectOK != (tt.claim.Subject != "") {
				t.Errorf("subjectOK = %v, want %v", subjectOK, tt.claim.Subject != "")
			}

			admin, adminOK := Admin(ctx)
			if admin != tt.claim.Admin {
				t.Errorf("admin = %v, want %v", admin, tt.claim.Admin)
			}
			if adminOK != tt.claim.Admin {
				t.Errorf("adminOK = %v, want %v", adminOK, tt.claim.Admin)
			}

			readOnly, readOnlyOK := ReadOnly(ctx)
			if readOnly != tt.claim.ReadOnly {
				t.Errorf("readOnly = %v, want %v", readOnly, tt.claim.ReadOnly)
			}
			if readOnlyOK != tt.claim.ReadOnly {
				t.Errorf("readOnlyOK = %v, want %v", readOnlyOK, tt.claim.ReadOnly)
			}

			jti, jtiOK := JTI(ctx)
			if jti != tt.claim.ID {
				t.Errorf("jti = %q, want %q", jti, tt.claim.ID)
			}
			if jtiOK != (tt.claim.ID != "") {
				t.Errorf("jtiOK = %v, want %v", jtiOK, tt.claim.ID != "")
			}
		})
	}
}
