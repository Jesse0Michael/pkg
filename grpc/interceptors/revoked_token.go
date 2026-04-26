package interceptors

import (
	"context"
	"fmt"

	"github.com/jesse0michael/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RevokedTokenUnaryServerInterceptor returns a unary interceptor that rejects
// requests whose JWT (by JTI) has been revoked.
// RPCs annotated with the no_auth option are skipped.
func RevokedTokenUnaryServerInterceptor(checker auth.RevokedTokenChecker) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if HasNoAuth(info.FullMethod) {
			return handler(ctx, req)
		}
		if err := checkRevoked(ctx, checker); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// RevokedTokenStreamServerInterceptor returns a stream interceptor that rejects
// requests whose JWT (by JTI) has been revoked.
// RPCs annotated with the no_auth option are skipped.
func RevokedTokenStreamServerInterceptor(checker auth.RevokedTokenChecker) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if HasNoAuth(info.FullMethod) {
			return handler(srv, ss)
		}
		if err := checkRevoked(ss.Context(), checker); err != nil {
			return err
		}
		return handler(srv, ss)
	}
}

func checkRevoked(ctx context.Context, checker auth.RevokedTokenChecker) error {
	jti, ok := auth.JTI(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing authentication context")
	}

	revoked, err := checker.IsRevoked(ctx, jti)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to check token revocation: %s", fmt.Sprintf("%v", err))
	}
	if revoked {
		return status.Error(codes.Unauthenticated, "token has been revoked")
	}
	return nil
}
