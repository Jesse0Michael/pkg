package interceptors

import (
	"context"
	"errors"
	"testing"

	"github.com/jesse0michael/pkg/auth"
	// Register the test proto so the global registry has services with options.
	_ "github.com/jesse0michael/pkg/grpc/proto/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type mockAuthenticator struct {
	claim *auth.Claim
	err   error
}

func (m *mockAuthenticator) VerifyAccessToken(_ string) (*auth.Claim, error) {
	return m.claim, m.err
}

func TestAuthUnaryServerInterceptor(t *testing.T) {
	tests := []struct {
		name       string
		auth       Authenticator
		token      string
		fullMethod string
		wantErr    bool
		wantCode   codes.Code
	}{
		{
			name:       "authorized",
			auth:       &mockAuthenticator{claim: &auth.Claim{}},
			token:      "test-token",
			fullMethod: "/testproto.TestService/Authed",
		},
		{
			name:       "no metadata",
			auth:       &mockAuthenticator{},
			fullMethod: "/testproto.TestService/Authed",
			wantErr:    true,
			wantCode:   codes.Unauthenticated,
		},
		{
			name:       "verify error",
			auth:       &mockAuthenticator{err: errors.New("test-error")},
			token:      "test-token",
			fullMethod: "/testproto.TestService/Authed",
			wantErr:    true,
			wantCode:   codes.Unauthenticated,
		},
		{
			name:       "no_auth method option skips authentication",
			auth:       &mockAuthenticator{err: errors.New("test-error")},
			fullMethod: "/testproto.TestService/Public",
		},
		{
			name:       "no_auth service option skips authentication",
			auth:       &mockAuthenticator{err: errors.New("test-error")},
			fullMethod: "/testproto.PublicService/DoPublic",
		},
		{
			name:       "admin_only method allows admin",
			auth:       &mockAuthenticator{claim: &auth.Claim{Admin: true}},
			token:      "test-token",
			fullMethod: "/testproto.TestService/AdminMethod",
		},
		{
			name:       "admin_only method denies non-admin",
			auth:       &mockAuthenticator{claim: &auth.Claim{}},
			token:      "test-token",
			fullMethod: "/testproto.TestService/AdminMethod",
			wantErr:    true,
			wantCode:   codes.PermissionDenied,
		},
		{
			name:       "admin_only service allows admin",
			auth:       &mockAuthenticator{claim: &auth.Claim{Admin: true}},
			token:      "test-token",
			fullMethod: "/testproto.AdminService/DoAdmin",
		},
		{
			name:       "admin_only service denies non-admin",
			auth:       &mockAuthenticator{claim: &auth.Claim{}},
			token:      "test-token",
			fullMethod: "/testproto.AdminService/DoAdmin",
			wantErr:    true,
			wantCode:   codes.PermissionDenied,
		},
		{
			name:       "reject_read_only method allows non-read-only user",
			auth:       &mockAuthenticator{claim: &auth.Claim{}},
			token:      "test-token",
			fullMethod: "/testproto.TestService/WriteMethod",
		},
		{
			name:       "reject_read_only method denies read-only user",
			auth:       &mockAuthenticator{claim: &auth.Claim{ReadOnly: true}},
			token:      "test-token",
			fullMethod: "/testproto.TestService/WriteMethod",
			wantErr:    true,
			wantCode:   codes.PermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := AuthUnaryServerInterceptor(tt.auth)

			ctx := t.Context()
			if tt.token != "" {
				ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(
					"authorization", "Bearer "+tt.token,
				))
			}

			handler := func(_ context.Context, _ any) (any, error) {
				return "ok", nil
			}

			info := &grpc.UnaryServerInfo{FullMethod: tt.fullMethod}
			got, err := interceptor(ctx, nil, info, handler)
			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.wantCode, status.Code(err))
				return
			}

			require.NoError(t, err)
			require.Equal(t, "ok", got)
		})
	}
}

func TestAuthStreamServerInterceptor(t *testing.T) {
	tests := []struct {
		name       string
		auth       Authenticator
		token      string
		fullMethod string
		wantErr    bool
		wantCode   codes.Code
	}{
		{
			name:       "authorized",
			auth:       &mockAuthenticator{claim: &auth.Claim{}},
			token:      "test-token",
			fullMethod: "/testproto.TestService/Authed",
		},
		{
			name:       "no metadata",
			auth:       &mockAuthenticator{},
			fullMethod: "/testproto.TestService/Authed",
			wantErr:    true,
			wantCode:   codes.Unauthenticated,
		},
		{
			name:       "verify error",
			auth:       &mockAuthenticator{err: errors.New("test-error")},
			token:      "test-token",
			fullMethod: "/testproto.TestService/Authed",
			wantErr:    true,
			wantCode:   codes.Unauthenticated,
		},
		{
			name:       "no_auth method option skips authentication",
			auth:       &mockAuthenticator{err: errors.New("test-error")},
			fullMethod: "/testproto.TestService/Public",
		},
		{
			name:       "admin_only method denies non-admin",
			auth:       &mockAuthenticator{claim: &auth.Claim{}},
			token:      "test-token",
			fullMethod: "/testproto.TestService/AdminMethod",
			wantErr:    true,
			wantCode:   codes.PermissionDenied,
		},
		{
			name:       "reject_read_only method denies read-only user",
			auth:       &mockAuthenticator{claim: &auth.Claim{ReadOnly: true}},
			token:      "test-token",
			fullMethod: "/testproto.TestService/WriteMethod",
			wantErr:    true,
			wantCode:   codes.PermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := AuthStreamServerInterceptor(tt.auth)

			ctx := t.Context()
			if tt.token != "" {
				ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(
					"authorization", "Bearer "+tt.token,
				))
			}

			ss := &fakeServerStream{ctx: ctx}
			handler := func(_ any, _ grpc.ServerStream) error {
				return nil
			}

			info := &grpc.StreamServerInfo{FullMethod: tt.fullMethod}
			err := interceptor(nil, ss, info, handler)
			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.wantCode, status.Code(err))
				return
			}

			require.NoError(t, err)
		})
	}
}

type fakeServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (f *fakeServerStream) Context() context.Context { return f.ctx }

func TestHasNoAuth(t *testing.T) {
	tests := []struct {
		name       string
		fullMethod string
		want       bool
	}{
		{
			name:       "method with no_auth option",
			fullMethod: "/testproto.TestService/Public",
			want:       true,
		},
		{
			name:       "method without no_auth option",
			fullMethod: "/testproto.TestService/Authed",
			want:       false,
		},
		{
			name:       "service with no_auth option",
			fullMethod: "/testproto.PublicService/DoPublic",
			want:       true,
		},
		{
			name:       "unknown service",
			fullMethod: "/unknown.Service/Method",
			want:       false,
		},
		{
			name:       "malformed method",
			fullMethod: "garbage",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, hasNoAuth(tt.fullMethod))
		})
	}
}

func TestHasAdminOnly(t *testing.T) {
	tests := []struct {
		name       string
		fullMethod string
		want       bool
	}{
		{
			name:       "method with admin_only option",
			fullMethod: "/testproto.TestService/AdminMethod",
			want:       true,
		},
		{
			name:       "method without admin_only option",
			fullMethod: "/testproto.TestService/Authed",
			want:       false,
		},
		{
			name:       "service with admin_only option",
			fullMethod: "/testproto.AdminService/DoAdmin",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, hasAdminOnly(tt.fullMethod))
		})
	}
}

func TestHasRejectReadOnly(t *testing.T) {
	tests := []struct {
		name       string
		fullMethod string
		want       bool
	}{
		{
			name:       "method with reject_read_only option",
			fullMethod: "/testproto.TestService/WriteMethod",
			want:       true,
		},
		{
			name:       "method without reject_read_only option",
			fullMethod: "/testproto.TestService/Authed",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, hasRejectReadOnly(tt.fullMethod))
		})
	}
}
