package oidc

import (
	"context"
	"fmt"
	"slices"

	"github.com/coreos/go-oidc/v3/oidc"
)

// GoogleAuthConfig holds the configuration for Google OIDC verification.
type GoogleAuthConfig struct {
	Issuer    string   `envconfig:"GOOGLE_ISSUER" default:"https://accounts.google.com"`
	ClientIDs []string `envconfig:"GOOGLE_CLIENT_ID"`
}

// GoogleAuth verifies Google-issued ID tokens against Google's JWKS.
type GoogleAuth struct {
	verifier  *oidc.IDTokenVerifier
	clientIDs []string
}

// NewGoogleAuth builds a GoogleAuth for Google ID tokens. The config's ClientIDs
// are the OAuth clients whose tokens are accepted — a token is valid if its
// audience matches any of them (e.g. the Desktop client for the CLI and the Web
// client for the site). Audience is checked manually because go-oidc only
// supports a single ClientID. It performs OIDC discovery against Google, so it
// needs network at construction time.
func NewGoogleAuth(ctx context.Context, cfg GoogleAuthConfig) (*GoogleAuth, error) {
	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to discover google oidc provider: %w", err)
	}
	return &GoogleAuth{
		verifier: provider.Verifier(&oidc.Config{
			SkipClientIDCheck: true,
		}),
		clientIDs: cfg.ClientIDs,
	}, nil
}

func (g *GoogleAuth) Verify(ctx context.Context, rawIDToken string) (*Claims, error) {
	tok, err := g.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify id token: %w", err)
	}

	// ensure the token's audience matches one of the accepted client IDs.
	if !slices.ContainsFunc(tok.Audience, func(a string) bool { return slices.Contains(g.clientIDs, a) }) {
		return nil, fmt.Errorf("token audience %v is not an accepted client", tok.Audience)
	}

	var raw struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := tok.Claims(&raw); err != nil {
		return nil, fmt.Errorf("failed to parse id token claims: %w", err)
	}

	return &Claims{
		Provider:      "google",
		Subject:       tok.Subject,
		Email:         raw.Email,
		EmailVerified: raw.EmailVerified,
		Name:          raw.Name,
		Picture:       raw.Picture,
	}, nil
}
