package interceptors

import (
	"context"
	"errors"
	"testing"

	"github.com/jesse0michael/pkg/auth"
	// Register the test proto so the global registry has services with options.
	_ "github.com/jesse0michael/pkg/grpc/proto/test"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockRevokedTokenChecker struct {
	revoked bool
	err     error
}

func (m *mockRevokedTokenChecker) IsRevoked(_ context.Context, _ string) (bool, error) {
	return m.revoked, m.err
}

func TestRevokedTokenUnaryServerInterceptor(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		checker    auth.RevokedTokenChecker
		fullMethod string
		wantErr    bool
		wantCode   codes.Code
	}{
		{
			name:       "not revoked",
			ctx:        context.WithValue(t.Context(), auth.JTIContextKey, "test-jti"),
			checker:    &mockRevokedTokenChecker{revoked: false},
			fullMethod: "/testproto.TestService/Authed",
		},
		{
			name:       "revoked token",
			ctx:        context.WithValue(t.Context(), auth.JTIContextKey, "test-jti"),
			checker:    &mockRevokedTokenChecker{revoked: true},
			fullMethod: "/testproto.TestService/Authed",
			wantErr:    true,
			wantCode:   codes.Unauthenticated,
		},
		{
			name:       "checker error",
			ctx:        context.WithValue(t.Context(), auth.JTIContextKey, "test-jti"),
			checker:    &mockRevokedTokenChecker{err: errors.New("test-error")},
			fullMethod: "/testproto.TestService/Authed",
			wantErr:    true,
			wantCode:   codes.Internal,
		},
		{
			name:       "no jti in context",
			ctx:        t.Context(),
			checker:    &mockRevokedTokenChecker{},
			fullMethod: "/testproto.TestService/Authed",
			wantErr:    true,
			wantCode:   codes.Unauthenticated,
		},
		{
			name:       "no_auth method skips check",
			ctx:        t.Context(),
			checker:    &mockRevokedTokenChecker{revoked: true},
			fullMethod: "/testproto.TestService/Public",
		},
		{
			name:       "no_auth service skips check",
			ctx:        t.Context(),
			checker:    &mockRevokedTokenChecker{revoked: true},
			fullMethod: "/testproto.PublicService/DoPublic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := RevokedTokenUnaryServerInterceptor(tt.checker)

			handler := func(_ context.Context, _ any) (any, error) {
				return "ok", nil
			}

			info := &grpc.UnaryServerInfo{FullMethod: tt.fullMethod}
			got, err := interceptor(tt.ctx, nil, info, handler)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if code := status.Code(err); code != tt.wantCode {
					t.Errorf("got code %v, want %v", code, tt.wantCode)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != "ok" {
				t.Errorf("got %v, want ok", got)
			}
		})
	}
}

func TestRevokedTokenStreamServerInterceptor(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		checker    auth.RevokedTokenChecker
		fullMethod string
		wantErr    bool
		wantCode   codes.Code
	}{
		{
			name:       "not revoked",
			ctx:        context.WithValue(t.Context(), auth.JTIContextKey, "test-jti"),
			checker:    &mockRevokedTokenChecker{revoked: false},
			fullMethod: "/testproto.TestService/Authed",
		},
		{
			name:       "revoked token",
			ctx:        context.WithValue(t.Context(), auth.JTIContextKey, "test-jti"),
			checker:    &mockRevokedTokenChecker{revoked: true},
			fullMethod: "/testproto.TestService/Authed",
			wantErr:    true,
			wantCode:   codes.Unauthenticated,
		},
		{
			name:       "checker error",
			ctx:        context.WithValue(t.Context(), auth.JTIContextKey, "test-jti"),
			checker:    &mockRevokedTokenChecker{err: errors.New("test-error")},
			fullMethod: "/testproto.TestService/Authed",
			wantErr:    true,
			wantCode:   codes.Internal,
		},
		{
			name:       "no_auth method skips check",
			ctx:        t.Context(),
			checker:    &mockRevokedTokenChecker{revoked: true},
			fullMethod: "/testproto.TestService/Public",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := RevokedTokenStreamServerInterceptor(tt.checker)

			ss := &fakeServerStream{ctx: tt.ctx}
			handler := func(_ any, _ grpc.ServerStream) error {
				return nil
			}

			info := &grpc.StreamServerInfo{FullMethod: tt.fullMethod}
			err := interceptor(nil, ss, info, handler)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if code := status.Code(err); code != tt.wantCode {
					t.Errorf("got code %v, want %v", code, tt.wantCode)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
