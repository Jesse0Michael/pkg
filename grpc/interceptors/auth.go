package interceptors

import (
	"context"
	"log/slog"
	"strings"

	"github.com/jesse0michael/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	ErrUnauthenticated = status.Error(codes.Unauthenticated, "unauthenticated")
)

// Authenticator verifies access tokens and returns claims.
type Authenticator interface {
	VerifyAccessToken(token string) (*auth.Claim, error)
}

// authenticate extracts the Bearer token from gRPC metadata, verifies it as an
// access token, and returns a context enriched with the claim values.
func authenticate(ctx context.Context, a Authenticator) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, ErrUnauthenticated
	}

	vals := md.Get("authorization")
	if len(vals) == 0 {
		return ctx, ErrUnauthenticated
	}

	token := strings.TrimPrefix(vals[0], "Bearer ")
	claims, err := a.VerifyAccessToken(token)
	if err != nil {
		slog.WarnContext(ctx, "JWT verification failed", "err", err)
		return ctx, ErrUnauthenticated
	}

	ctx = context.WithValue(ctx, auth.AuthorizationContextKey, token)
	ctx = auth.WithClaims(ctx, claims)
	ctx = auth.WithSpan(ctx)

	return ctx, nil
}

// AuthUnaryServerInterceptor returns a gRPC unary server interceptor that
// authenticates requests using the provided JWTAuth. Methods in the skip set
// bypass authentication (e.g. Login, RefreshToken).
func AuthUnaryServerInterceptor(a Authenticator, skip map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if skip[info.FullMethod] {
			return handler(ctx, req)
		}

		ctx, err := authenticate(ctx, a)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// AuthStreamServerInterceptor returns a gRPC stream server interceptor that
// authenticates requests using the provided JWTAuth. Methods in the skip set
// bypass authentication.
func AuthStreamServerInterceptor(a Authenticator, skip map[string]bool) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if skip[info.FullMethod] {
			return handler(srv, ss)
		}

		ctx, err := authenticate(ss.Context(), a)
		if err != nil {
			return err
		}

		return handler(srv, &wrappedServerStream{ServerStream: ss, ctx: ctx})
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context { return w.ctx }
