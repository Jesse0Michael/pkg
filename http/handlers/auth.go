package handlers

import (
	"context"
	"net/http"

	"github.com/jesse0michael/pkg/auth"
	"github.com/jesse0michael/pkg/auth/oidc"
	httperrors "github.com/jesse0michael/pkg/http/errors"
	"github.com/jesse0michael/pkg/http/server"
)

// OIDCResolver maps verified OIDC identity claims to token options. This is
// where the caller does app-specific work — DB lookups, user creation, linking
// accounts — and returns the options that configure the issued tokens.
type OIDCResolver func(ctx context.Context, claims *oidc.Claims) ([]auth.TokenOption, error)

// IdentityResolver maps a username/email and password to token options. The
// caller handles credential verification, user lookup, and returns the options
// that configure the issued tokens.
type IdentityResolver func(ctx context.Context, identity, password string) ([]auth.TokenOption, error)

type OIDCRequest struct {
	IDToken string `json:"id_token"`
}

type LoginRequest struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse is the standard token pair response for auth endpoints.
type AuthResponse struct {
	Subject      string `json:"subject"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// HandleOIDCLogin returns a handler that verifies an OIDC ID token, resolves
// the user via the provided resolver, and issues access/refresh tokens.
func HandleOIDCLogin(verifier oidc.Verifier, jwtAuth *auth.JWTAuth, resolve OIDCResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req OIDCRequest
		if err := server.Decode(r.Body, &req); err != nil {
			httperrors.WriteError(ctx, w, httperrors.NewError(http.StatusBadRequest, "invalid request body", err.Error()))
			return
		}
		if req.IDToken == "" {
			httperrors.WriteError(ctx, w, httperrors.NewError(http.StatusBadRequest, "id_token is required", ""))
			return
		}

		claims, err := verifier.Verify(ctx, req.IDToken)
		if err != nil {
			httperrors.WriteError(ctx, w, httperrors.NewError(http.StatusUnauthorized, "invalid id token", ""))
			return
		}

		opts, err := resolve(ctx, claims)
		if err != nil {
			httperrors.WriteError(ctx, w, err)
			return
		}

		issueTokens(ctx, w, jwtAuth, opts)
	})
}

// HandleLogin returns a handler that authenticates via identity/password,
// resolves the user via the provided resolver, and issues access/refresh tokens.
func HandleLogin(jwtAuth *auth.JWTAuth, resolve IdentityResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req LoginRequest
		if err := server.Decode(r.Body, &req); err != nil {
			httperrors.WriteError(ctx, w, httperrors.NewError(http.StatusBadRequest, "invalid request body", err.Error()))
			return
		}
		if req.Identity == "" || req.Password == "" {
			httperrors.WriteError(ctx, w, httperrors.NewError(http.StatusBadRequest, "identity and password are required", ""))
			return
		}

		opts, err := resolve(ctx, req.Identity, req.Password)
		if err != nil {
			httperrors.WriteError(ctx, w, err)
			return
		}

		issueTokens(ctx, w, jwtAuth, opts)
	})
}

// HandleRefreshTokens returns a handler that validates a refresh token and
// issues a new access/refresh token pair.
func HandleRefreshTokens(jwtAuth *auth.JWTAuth) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req RefreshRequest
		if err := server.Decode(r.Body, &req); err != nil {
			httperrors.WriteError(ctx, w, httperrors.NewError(http.StatusBadRequest, "invalid request body", err.Error()))
			return
		}
		if req.RefreshToken == "" {
			httperrors.WriteError(ctx, w, httperrors.NewError(http.StatusBadRequest, "refresh_token is required", ""))
			return
		}

		access, refresh, err := jwtAuth.RefreshTokens(req.RefreshToken)
		if err != nil {
			httperrors.WriteError(ctx, w, httperrors.NewError(http.StatusUnauthorized, "invalid refresh token", ""))
			return
		}

		// Extract subject from the new access token for the response.
		claims, _ := jwtAuth.VerifyAccessToken(access)
		subject := ""
		if claims != nil {
			subject = claims.Subject
		}

		//TODO validate the JTI and/or SUB

		_ = server.Encode(w, http.StatusOK, AuthResponse{
			Subject:      subject,
			AccessToken:  access,
			RefreshToken: refresh,
		})
	})
}

func issueTokens(ctx context.Context, w http.ResponseWriter, jwtAuth *auth.JWTAuth, opts []auth.TokenOption) {
	access, refresh, err := jwtAuth.GenerateTokens(opts...)
	if err != nil {
		httperrors.WriteError(ctx, w, err)
		return
	}

	// Extract subject from the new access token for the response.
	claims, _ := jwtAuth.VerifyAccessToken(access)
	subject := ""
	if claims != nil {
		subject = claims.Subject
	}

	_ = server.Encode(w, http.StatusOK, AuthResponse{
		Subject:      subject,
		AccessToken:  access,
		RefreshToken: refresh,
	})
}
