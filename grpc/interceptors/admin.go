package interceptors

import (
	"context"

	"github.com/jesse0michael/pkg/auth"
	"google.golang.org/grpc"
)

// AdminUnaryServerInterceptor returns a gRPC unary server interceptor that
// requires the caller to have Admin claims. It expects the auth interceptor
// to have already populated the context with claims.
func AdminUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if admin, _ := auth.Admin(ctx); !admin {
			return nil, ErrPermissionDenied
		}
		return handler(ctx, req)
	}
}

// AdminStreamServerInterceptor returns a gRPC stream server interceptor that
// requires the caller to have Admin claims. It expects the auth interceptor
// to have already populated the context with claims.
func AdminStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if admin, _ := auth.Admin(ss.Context()); !admin {
			return ErrPermissionDenied
		}
		return handler(srv, ss)
	}
}
