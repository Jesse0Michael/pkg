package interceptors

import (
	"context"
	"sync"

	"github.com/jesse0michael/pkg/auth"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

var ErrRateLimited = status.Error(codes.ResourceExhausted, "rate limit exceeded")

// RateLimitConfig configures the rate limiting interceptor.
type RateLimitConfig struct {
	// Rate is the token refill rate (per second) for authenticated users.
	// Default ~1.67/s = 100 requests per minute.
	Rate rate.Limit `envconfig:"RATE_LIMIT_RATE" default:"1.67"`
	// Burst is the maximum burst size for authenticated users.
	Burst int `envconfig:"RATE_LIMIT_BURST" default:"100"`
	// NoAuthRate is the token refill rate for unauthenticated requests,
	// keyed by peer address. Default ~0.17/s = 10 requests per minute.
	NoAuthRate rate.Limit `envconfig:"RATE_LIMIT_NO_AUTH_RATE" default:"0.17"`
	// NoAuthBurst is the maximum burst size for unauthenticated requests, keyed by peer address.
	NoAuthBurst int `envconfig:"RATE_LIMIT_NO_AUTH_BURST" default:"20"`
}

// RateLimiter provides in-memory rate limiting for gRPC RPCs.
// Authenticated users are keyed by auth.Subject; unauthenticated requests
// are keyed by peer address.
//
// TODO: The rates sync.Map will grow throughout the lifetime of the server.
// Eviction or a periodic fresh swap should be considered.
type RateLimiter struct {
	cfg   RateLimitConfig
	rates sync.Map // map[string]*rate.Limiter
}

// NewRateLimiter creates a RateLimiter with the given configuration.
func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		cfg: cfg,
	}
}

// UnaryServerInterceptor returns a gRPC unary server interceptor that rate
// limits requests. Authenticated users get per-subject limits; unauthenticated
// requests get per-peer limits.
func (rl *RateLimiter) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if !rl.allow(ctx) {
			return nil, ErrRateLimited
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a gRPC stream server interceptor that rate
// limits requests. Authenticated users get per-subject limits; unauthenticated
// requests get per-peer limits.
func (rl *RateLimiter) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !rl.allow(ss.Context()) {
			return ErrRateLimited
		}
		return handler(srv, ss)
	}
}

// allow checks whether the request should be allowed.
// Requests with an authenticated subject get per-subject limits;
// everything else gets per-peer limits with the NoAuth rates.
func (rl *RateLimiter) allow(ctx context.Context) bool {
	if subject, ok := auth.Subject(ctx); ok {
		v, _ := rl.rates.LoadOrStore(subject, rate.NewLimiter(rl.cfg.Rate, rl.cfg.Burst))
		return v.(*rate.Limiter).Allow()
	}

	key := peerAddr(ctx)
	v, _ := rl.rates.LoadOrStore(key, rate.NewLimiter(rl.cfg.NoAuthRate, rl.cfg.NoAuthBurst))
	return v.(*rate.Limiter).Allow()
}

// peerAddr extracts the peer address from the gRPC context.
func peerAddr(ctx context.Context) string {
	if p, ok := peer.FromContext(ctx); ok {
		return p.Addr.String()
	}
	return "unknown"
}
