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
	ErrUnauthenticated  = status.Error(codes.Unauthenticated, "unauthenticated")
	ErrPermissionDenied = status.Error(codes.PermissionDenied, "permission denied")
)

// Authenticator verifies access tokens and returns claims.
type Authenticator interface {
	VerifyAccessToken(token string) (*auth.Claim, error)
}

// AuthUnaryServerInterceptor returns a gRPC unary server interceptor that
// authenticates and authorizes requests using the provided Authenticator.
// RPCs respect no_auth, admin_only, and reject_read_only method/service options.
func AuthUnaryServerInterceptor(a Authenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if HasNoAuth(info.FullMethod) {
			return handler(ctx, req)
		}

		ctx, claims, err := authenticate(ctx, a)
		if err != nil {
			return nil, err
		}

		if err := authorize(claims, info.FullMethod); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// AuthStreamServerInterceptor returns a gRPC stream server interceptor that
// authenticates and authorizes requests using the provided Authenticator.
// RPCs respect no_auth, admin_only, and reject_read_only method/service options.
func AuthStreamServerInterceptor(a Authenticator) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if HasNoAuth(info.FullMethod) {
			return handler(srv, ss)
		}

		ctx, claims, err := authenticate(ss.Context(), a)
		if err != nil {
			return err
		}

		if err := authorize(claims, info.FullMethod); err != nil {
			return err
		}

		return handler(srv, &wrappedServerStream{ServerStream: ss, ctx: ctx})
	}
}

// authenticate extracts the Bearer token from gRPC metadata, verifies it as an
// access token, and returns a context enriched with the claim values.
func authenticate(ctx context.Context, a Authenticator) (context.Context, *auth.Claim, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, nil, ErrUnauthenticated
	}

	vals := md.Get("authorization")
	if len(vals) == 0 {
		return ctx, nil, ErrUnauthenticated
	}

	token := strings.TrimPrefix(vals[0], "Bearer ")
	claims, err := a.VerifyAccessToken(token)
	if err != nil {
		slog.WarnContext(ctx, "JWT verification failed", "err", err)
		return ctx, nil, ErrUnauthenticated
	}

	ctx = context.WithValue(ctx, auth.AuthorizationContextKey, token)
	ctx = auth.WithClaims(ctx, claims)
	ctx = auth.WithSpan(ctx)

	return ctx, claims, nil
}

// authorize checks admin_only and reject_read_only constraints against the authenticated claims.
func authorize(claims *auth.Claim, fullMethod string) error {
	if HasAdminOnly(fullMethod) && !claims.Admin {
		return ErrPermissionDenied
	}
	if HasRejectReadOnly(fullMethod) && claims.ReadOnly {
		return ErrPermissionDenied
	}
	return nil
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context { return w.ctx }
