package interceptors

import (
	"context"
	"net"
	"testing"

	"github.com/jesse0michael/pkg/auth"
	"google.golang.org/grpc"
	grpcpeer "google.golang.org/grpc/peer"
)

func TestRateLimiterUnaryServerInterceptor(t *testing.T) {
	tests := []struct {
		name    string
		cfg     RateLimitConfig
		ctx     context.Context
		wantErr bool
	}{
		{
			name: "authenticated request allowed",
			cfg:  RateLimitConfig{Rate: 10, Burst: 10, NoAuthRate: 1, NoAuthBurst: 1},
			ctx:  context.WithValue(t.Context(), auth.SubjectContextKey, "test-user"),
		},
		{
			name:    "authenticated request rate limited",
			cfg:     RateLimitConfig{Rate: 0, Burst: 1, NoAuthRate: 0, NoAuthBurst: 10},
			ctx:     context.WithValue(t.Context(), auth.SubjectContextKey, "test-user"),
			wantErr: true,
		},
		{
			name: "unauthenticated request rate limited by peer",
			cfg:  RateLimitConfig{Rate: 10, Burst: 10, NoAuthRate: 0, NoAuthBurst: 1},
			ctx: grpcpeer.NewContext(t.Context(), &grpcpeer.Peer{
				Addr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 1234},
			}),
			wantErr: true,
		},
		{
			name: "unauthenticated request allowed within limit",
			cfg:  RateLimitConfig{Rate: 10, Burst: 10, NoAuthRate: 10, NoAuthBurst: 10},
			ctx: grpcpeer.NewContext(t.Context(), &grpcpeer.Peer{
				Addr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 1234},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := NewRateLimiter(tt.cfg).UnaryServerInterceptor()
			handler := func(_ context.Context, _ any) (any, error) { return "ok", nil }
			info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

			var err error
			for range 3 {
				_, err = interceptor(tt.ctx, nil, info, handler)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("got err=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}

	// Verify independent subject limits: a second subject should still be allowed
	// even after the first subject exhausted their limit.
	t.Run("independent subjects verified", func(t *testing.T) {
		interceptor := NewRateLimiter(RateLimitConfig{Rate: 0, Burst: 1, NoAuthRate: 0, NoAuthBurst: 1}).UnaryServerInterceptor()
		handler := func(_ context.Context, _ any) (any, error) { return "ok", nil }
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

		ctx1 := context.WithValue(t.Context(), auth.SubjectContextKey, "test-user-1")
		if _, err := interceptor(ctx1, nil, info, handler); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := interceptor(ctx1, nil, info, handler); err == nil {
			t.Fatal("expected error on second request for test-user-1")
		}

		ctx2 := context.WithValue(t.Context(), auth.SubjectContextKey, "test-user-2")
		if _, err := interceptor(ctx2, nil, info, handler); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// Verify independent peer limits: a second peer should still be allowed
	// even after the first peer exhausted their limit.
	t.Run("independent peers verified", func(t *testing.T) {
		interceptor := NewRateLimiter(RateLimitConfig{Rate: 0, Burst: 1, NoAuthRate: 0, NoAuthBurst: 1}).UnaryServerInterceptor()
		handler := func(_ context.Context, _ any) (any, error) { return "ok", nil }
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

		peer1 := grpcpeer.NewContext(t.Context(), &grpcpeer.Peer{
			Addr: &net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1000},
		})
		if _, err := interceptor(peer1, nil, info, handler); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := interceptor(peer1, nil, info, handler); err == nil {
			t.Fatal("expected error on second request for peer1")
		}

		peer2 := grpcpeer.NewContext(t.Context(), &grpcpeer.Peer{
			Addr: &net.TCPAddr{IP: net.ParseIP("10.0.0.2"), Port: 2000},
		})
		if _, err := interceptor(peer2, nil, info, handler); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRateLimiterStreamServerInterceptor(t *testing.T) {
	tests := []struct {
		name    string
		cfg     RateLimitConfig
		ctx     context.Context
		wantErr bool
	}{
		{
			name: "stream allowed",
			cfg:  RateLimitConfig{Rate: 10, Burst: 10, NoAuthRate: 1, NoAuthBurst: 1},
			ctx:  context.WithValue(t.Context(), auth.SubjectContextKey, "test-user"),
		},
		{
			name:    "stream rate limited",
			cfg:     RateLimitConfig{Rate: 0, Burst: 1, NoAuthRate: 0, NoAuthBurst: 10},
			ctx:     context.WithValue(t.Context(), auth.SubjectContextKey, "test-user"),
			wantErr: true,
		},
		{
			name: "stream unauthenticated rate limited",
			cfg:  RateLimitConfig{Rate: 10, Burst: 10, NoAuthRate: 0, NoAuthBurst: 1},
			ctx: grpcpeer.NewContext(t.Context(), &grpcpeer.Peer{
				Addr: &net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1000},
			}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := NewRateLimiter(tt.cfg).StreamServerInterceptor()
			ss := &fakeServerStream{ctx: tt.ctx}
			handler := func(_ any, _ grpc.ServerStream) error { return nil }
			info := &grpc.StreamServerInfo{FullMethod: "/test.Service/Method"}

			var err error
			for range 3 {
				err = interceptor(nil, ss, info, handler)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("got err=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}
