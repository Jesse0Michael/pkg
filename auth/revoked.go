package auth

import "context"

// RevokedTokenChecker abstracts the storage lookup for revoked tokens.
// Implement this against your database and pass it to the revoked-token middleware.
type RevokedTokenChecker interface {
	IsRevoked(ctx context.Context, jti string) (bool, error)
}
