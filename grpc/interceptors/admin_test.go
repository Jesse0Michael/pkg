package interceptors

import (
	"context"
	"testing"

	"github.com/jesse0michael/pkg/auth"
	_ "github.com/jesse0michael/pkg/grpc/proto/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAdminUnaryServerInterceptor(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		fullMethod string
		wantErr    bool
		wantCode   codes.Code
	}{
		{
			name: "admin claim allowed",
			ctx:  context.WithValue(t.Context(), auth.AdminContextKey, true),
		},
		{
			name:     "non-admin claim denied",
			ctx:      context.WithValue(t.Context(), auth.AdminContextKey, false),
			wantErr:  true,
			wantCode: codes.PermissionDenied,
		},
		{
			name:     "no claim denied",
			ctx:      t.Context(),
			wantErr:  true,
			wantCode: codes.PermissionDenied,
		},
		{
			name:       "no_auth method skips admin check",
			ctx:        t.Context(),
			fullMethod: "/testproto.TestService/Public",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := AdminUnaryServerInterceptor()

			handler := func(_ context.Context, _ any) (any, error) {
				return "ok", nil
			}

			got, err := interceptor(tt.ctx, nil, &grpc.UnaryServerInfo{FullMethod: tt.fullMethod}, handler)
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

func TestAdminStreamServerInterceptor(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		fullMethod string
		wantErr    bool
		wantCode   codes.Code
	}{
		{
			name: "admin claim allowed",
			ctx:  context.WithValue(t.Context(), auth.AdminContextKey, true),
		},
		{
			name:     "non-admin claim denied",
			ctx:      context.WithValue(t.Context(), auth.AdminContextKey, false),
			wantErr:  true,
			wantCode: codes.PermissionDenied,
		},
		{
			name:     "no claim denied",
			ctx:      t.Context(),
			wantErr:  true,
			wantCode: codes.PermissionDenied,
		},
		{
			name:       "no_auth method skips admin check",
			ctx:        t.Context(),
			fullMethod: "/testproto.TestService/Public",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := AdminStreamServerInterceptor()

			ss := &fakeServerStream{ctx: tt.ctx}
			handler := func(_ any, _ grpc.ServerStream) error {
				return nil
			}

			err := interceptor(nil, ss, &grpc.StreamServerInfo{FullMethod: tt.fullMethod}, handler)
			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.wantCode, status.Code(err))
				return
			}

			require.NoError(t, err)
		})
	}
}
