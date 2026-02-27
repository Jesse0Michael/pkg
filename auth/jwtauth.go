package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AccessTokenType  = "access"
	RefreshTokenType = "refresh"
)

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token expired")
	ErrInvalidSigningKey = errors.New("invalid signing key")
)

type Config struct {
	// SecretKey used for signing tokens (required)
	SecretKey []byte `envconfig:"AUTH_SECRET_KEY" required:"true"`

	// Issuer claim to include in tokens
	Issuer string `envconfig:"AUTH_ISSUER"`

	// Time-to-live for access tokens
	AccessTokenTTL time.Duration `envconfig:"AUTH_ACCESS_TOKEN_TTL" default:"168h"` // 7 days

	// Time-to-live for refresh tokens
	RefreshTokenTTL time.Duration `envconfig:"AUTH_REFRESH_TOKEN_TTL" default:"720h"` // 30 days
}

type Claim struct {
	// Admin indicates if the user has admin privileges
	Admin bool `json:"admin,omitempty"`

	// ReadOnly indicates if the user has read-only access
	ReadOnly bool `json:"readOnly,omitempty"`

	// TokenType indicates the type of the token (access or refresh)
	TokenType string `json:"type,omitempty"`

	jwt.RegisteredClaims
}

type JWTAuth struct {
	cfg           Config
	signingMethod jwt.SigningMethod
	Options       []jwt.ParserOption
}

func NewJWTAuth(cfg Config, signingMethod jwt.SigningMethod, opts ...jwt.ParserOption) *JWTAuth {
	return &JWTAuth{
		cfg:           cfg,
		signingMethod: signingMethod,
		Options:       opts,
	}
}

// GenerateTokens creates both access and refresh tokens for a user in one call
func (a *JWTAuth) GenerateTokens(opts ...TokenOption) (string, string, error) {
	p := applyTokenOptions(opts)

	accessToken, err := a.generateToken(p, AccessTokenType, a.cfg.AccessTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := a.generateToken(p, RefreshTokenType, a.cfg.RefreshTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// generateToken generates a token with the supplied parameters
func (a *JWTAuth) generateToken(p tokenParams, tokenType string, ttl time.Duration) (string, error) {
	tokenID := uuid.New().String()
	now := time.Now()
	expiresAt := now.Add(ttl)
	claims := Claim{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.cfg.Issuer,
			Subject:   p.subject,
			Audience:  p.audience,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        tokenID,
		},
		TokenType: tokenType,
		Admin:     p.admin,
		ReadOnly:  p.readOnly,
	}

	token := jwt.NewWithClaims(a.signingMethod, claims)

	return token.SignedString(a.cfg.SecretKey)
}

// VerifyToken validates a token and returns the claims
func (a *JWTAuth) VerifyToken(tokenString, expectedType string) (*Claim, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claim{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != a.signingMethod {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return a.cfg.SecretKey, nil
		},
		jwt.WithIssuer(a.cfg.Issuer),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claim)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("%w: expected %s token, got %s", ErrInvalidToken, expectedType, claims.TokenType)
	}
	if claims.Subject == "" {
		return nil, fmt.Errorf("%w: missing subject", ErrInvalidToken)
	}

	return claims, nil
}

// VerifyAccessToken specifically validates access tokens
func (s *JWTAuth) VerifyAccessToken(token string) (*Claim, error) {
	return s.VerifyToken(token, AccessTokenType)
}

// VerifyRefreshToken specifically validates refresh tokens
func (s *JWTAuth) VerifyRefreshToken(token string) (*Claim, error) {
	return s.VerifyToken(token, RefreshTokenType)
}

// RefreshTokens validates a refresh token and issues new access and refresh tokens
func (s *JWTAuth) RefreshTokens(token string) (string, string, error) {
	claims, err := s.VerifyRefreshToken(token)
	if err != nil {
		return "", "", err
	}

	opts := []TokenOption{WithSubject(claims.Subject)}
	if claims.Admin {
		opts = append(opts, WithAdmin())
	}
	if claims.ReadOnly {
		opts = append(opts, WithReadOnly())
	}
	if len(claims.Audience) > 0 {
		opts = append(opts, WithAudience(claims.Audience...))
	}

	return s.GenerateTokens(opts...)
}

// WithClaims sets identifying information from the claims into the context
func WithClaims(ctx context.Context, claim *Claim) context.Context {
	if claim.Subject != "" {
		ctx = context.WithValue(ctx, SubjectContextKey, claim.Subject)
	}
	if claim.Admin {
		ctx = context.WithValue(ctx, AdminContextKey, claim.Admin)
	}
	if claim.ReadOnly {
		ctx = context.WithValue(ctx, ReadOnlyContextKey, claim.ReadOnly)
	}
	if claim.ID != "" {
		ctx = context.WithValue(ctx, JTIContextKey, claim.ID)
	}
	return ctx
}
