package interceptors

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestLogUnaryServerInterceptor(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantLog bool
	}{
		{
			name: "no error no log",
		},
		{
			name:    "error logged",
			err:     errors.New("test-error"),
			wantLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))

			interceptor := LogUnaryServerInterceptor()
			handler := func(_ context.Context, _ any) (any, error) { return nil, tt.err }
			info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

			_, _ = interceptor(t.Context(), nil, info, handler)

			require.Equal(t, tt.wantLog, buf.Len() > 0)
		})
	}
}

func TestLogStreamServerInterceptor(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantLog bool
	}{
		{
			name: "no error no log",
		},
		{
			name:    "error logged",
			err:     errors.New("test-error"),
			wantLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))

			interceptor := LogStreamServerInterceptor()
			ss := &fakeServerStream{ctx: t.Context()}
			handler := func(_ any, _ grpc.ServerStream) error { return tt.err }
			info := &grpc.StreamServerInfo{FullMethod: "/test.Service/Method"}

			_ = interceptor(nil, ss, info, handler)

			require.Equal(t, tt.wantLog, buf.Len() > 0)
		})
	}
}
