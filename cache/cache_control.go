package cache

import (
	"context"

	"github.com/marcw/cachecontrol"
)

type contextKey string

const CacheControlContextKey = contextKey("cacheControl")

// NoCache is a Cache Control helper method that will check context for cache control settings
// and retrieve the no-cache value
func NoCache(ctx context.Context) bool {
	if cc, ok := ctx.Value(CacheControlContextKey).(cachecontrol.CacheControl); ok {
		b, _ := cc.NoCache()
		return b
	}
	return false
}

// NoStore is a Cache Control helper method that will check context for cache control settings
// and retrieve the no-store value
func NoStore(ctx context.Context) bool {
	if cc, ok := ctx.Value(CacheControlContextKey).(cachecontrol.CacheControl); ok {
		return cc.NoStore()
	}
	return false
}
