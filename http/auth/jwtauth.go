package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaim struct {
	Admin    bool `json:"admin"`
	ReadOnly bool `json:"readOnly"`
	jwt.RegisteredClaims
}

type JWTAuth struct {
	AuthKey any
	Options []jwt.ParserOption
}

// Authenticate parses and validates a JWT using a key provided in the JWTAuth.
// If valid, it will set identifying information from the claims into the request context.
func (a *JWTAuth) Authenticate(r *http.Request) (context.Context, bool) {
	auth := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	var claim JWTClaim

	token, err := jwt.ParseWithClaims(auth, &claim, func(token *jwt.Token) (any, error) {
		return a.AuthKey, nil
	}, a.Options...)
	if err != nil || !token.Valid {
		return r.Context(), false
	}

	ctx := context.WithValue(r.Context(), AuthorizationContextKey, auth)
	ctx = context.WithValue(ctx, SubjectContextKey, claim.Subject)
	ctx = context.WithValue(ctx, AdminContextKey, claim.Admin)
	ctx = context.WithValue(ctx, ReadOnlyContextKey, claim.ReadOnly)
	return ctx, true
}
