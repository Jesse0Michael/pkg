package cache

import (
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

type Config struct {
	Enabled        bool          `envconfig:"CACHE_ENABLED" default:"true"`
	Size           int           `envconfig:"CACHE_SIZE" default:"10000"`
	LocalTTL       time.Duration `envconfig:"CACHE_LOCAL_TTL" default:"5m"`
	TTL            time.Duration `envconfig:"CACHE_TTL" default:"1h"`
	Timeout        time.Duration `envconfig:"CACHE_TIMEOUT" default:"500ms"`
	BreakerTimeout time.Duration `envconfig:"CACHE_BREAKER_TIMEOUT" default:"60s"`
}

// NewCache returns a go-redis/cache using the Config object with a breaker wrapped redis client
func NewCache(cfg Config, r *redis.Client) *cache.Cache {
	return cache.New(&cache.Options{
		Redis:      NewRedisBreaker(cfg, r),
		LocalCache: cache.NewTinyLFU(cfg.Size, cfg.LocalTTL),
	})
}
