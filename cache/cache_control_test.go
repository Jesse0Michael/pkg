package cache

import (
	"context"
	"testing"

	"github.com/marcw/cachecontrol"
)

func TestNoCache(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want bool
	}{
		{
			name: "cache control not found",
			ctx:  context.TODO(),
			want: false,
		},
		{
			name: "cache control wrong type",
			ctx:  context.WithValue(context.TODO(), CacheControlContextKey, false),
			want: false,
		},
		{
			name: "cache control without no cache",
			ctx:  context.WithValue(context.TODO(), CacheControlContextKey, cachecontrol.CacheControl{}),
			want: false,
		},
		{
			name: "cache control with no cache",
			ctx:  context.WithValue(context.TODO(), CacheControlContextKey, cachecontrol.Parse("no-cache")),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NoCache(tt.ctx)
			if got != tt.want {
				t.Errorf("NoCache() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNoStore(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want bool
	}{
		{
			name: "cache control not found",
			ctx:  context.TODO(),
			want: false,
		},
		{
			name: "cache control wrong type",
			ctx:  context.WithValue(context.TODO(), CacheControlContextKey, false),
			want: false,
		},
		{
			name: "cache control without no store",
			ctx:  context.WithValue(context.TODO(), CacheControlContextKey, cachecontrol.CacheControl{}),
			want: false,
		},
		{
			name: "cache control with no store",
			ctx:  context.WithValue(context.TODO(), CacheControlContextKey, cachecontrol.Parse("no-store")),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NoStore(tt.ctx)
			if got != tt.want {
				t.Errorf("NoStore() got = %v, want %v", got, tt.want)
			}
		})
	}
}
