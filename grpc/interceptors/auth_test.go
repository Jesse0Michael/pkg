package interceptors

import (
	"context"
	"errors"
	"testing"

	"github.com/jesse0michael/pkg/auth"
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
		name     string
		auth     Authenticator
		token    string
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name:  "authorized",
			auth:  &mockAuthenticator{claim: &auth.Claim{}},
			token: "valid-token",
		},
		{
			name:     "no metadata",
			auth:     &mockAuthenticator{},
			wantErr:  true,
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "verify error",
			auth:     &mockAuthenticator{err: errors.New("bad token")},
			token:    "bad-token",
			wantErr:  true,
			wantCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := AuthUnaryServerInterceptor(tt.auth, nil)

			ctx := t.Context()
			if tt.token != "" {
				ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(
					"authorization", "Bearer "+tt.token,
				))
			}

			handler := func(_ context.Context, _ any) (any, error) {
				return "ok", nil
			}

			info := &grpc.UnaryServerInfo{FullMethod: "/test/Method"}
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
		name     string
		auth     Authenticator
		token    string
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name:  "authorized",
			auth:  &mockAuthenticator{claim: &auth.Claim{}},
			token: "valid-token",
		},
		{
			name:     "no metadata",
			auth:     &mockAuthenticator{},
			wantErr:  true,
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "verify error",
			auth:     &mockAuthenticator{err: errors.New("bad token")},
			token:    "bad-token",
			wantErr:  true,
			wantCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := AuthStreamServerInterceptor(tt.auth, nil)

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

			info := &grpc.StreamServerInfo{FullMethod: "/test/StreamMethod"}
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
