package interceptors

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LogUnaryServerInterceptor returns a gRPC unary server interceptor that
// logs errors returned by handlers.
func LogUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			slog.ErrorContext(ctx, "rpc error",
				"method", info.FullMethod,
				"code", status.Code(err).String(),
				"request", req,
				"err", err,
			)
		}
		return resp, err
	}
}

// LogStreamServerInterceptor returns a gRPC stream server interceptor that
// logs errors returned by handlers.
func LogStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, ss)
		if err != nil {
			slog.ErrorContext(ss.Context(), "rpc error",
				"method", info.FullMethod,
				"code", status.Code(err).String(),
				"err", err,
			)
		}
		return err
	}
}
