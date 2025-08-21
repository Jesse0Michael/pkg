package auth

import (
	"context"
	"net/http"
)

type BasicAuthenticator struct {
	Username string
	Password string
}

func NewBasicAuthenticator(username string, password string) *BasicAuthenticator {
	return &BasicAuthenticator{
		Username: username,
		Password: password,
	}
}

func (b *BasicAuthenticator) Authenticate(r *http.Request) (context.Context, bool) {
	user, pass, _ := r.BasicAuth()
	if user == b.Username && pass == b.Password {
		return r.Context(), true
	}
	return r.Context(), false
}
