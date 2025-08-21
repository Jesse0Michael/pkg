package auth

import (
	"net/http"
	"testing"
)

func TestNewBasicAuthenticator(t *testing.T) {
	username := "test"
	password := "test"
	basic := NewBasicAuthenticator(username, password)

	correct, _ := http.NewRequest(http.MethodGet, "/", nil)
	correct.SetBasicAuth("test", "test")

	incorrect, _ := http.NewRequest(http.MethodGet, "/", nil)
	incorrect.SetBasicAuth("bad", "password")

	_, auth := basic.Authenticate(correct)
	if !auth {
		t.Errorf("auth.Authenticate != expected")
	}

	_, auth = basic.Authenticate(incorrect)
	if auth {
		t.Errorf("auth.Authenticate != expected")
	}
}
