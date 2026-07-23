package oidc

import "context"

// Verifier validates a raw ID token and returns its normalized claims.
type Verifier interface {
	Verify(ctx context.Context, rawIDToken string) (*Claims, error)
}

// Claims are the normalized identity claims extracted from a verified ID token.
type Claims struct {
	// Provider is the login method these claims came from, e.g. "google".
	Provider string
	// Subject is the provider's stable, unique identifier for the user (sub).
	Subject string
	Email   string
	// EmailVerified reports whether the provider vouches for the email.
	EmailVerified bool
	Name          string
	Picture       string
}
