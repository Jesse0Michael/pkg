package cache

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sony/gobreaker"
)

// ResilientCache is a cache wrapper that implements the go-cache rediser interface.
// It supports the cache control settings passed in context.
// https://github.com/go-redis/cache/blob/8756f3baa759d22acfdc1dc67f9fbcc0e21e6332/cache.go#L32
type ResilientCache struct {
	cfg     Config
	redis   *redis.Client
	breaker *gobreaker.CircuitBreaker
}

func NewRedisBreaker(cfg Config, r *redis.Client) *ResilientCache {
	settings := gobreaker.Settings{
		Name:    "cache breaker",
		Timeout: cfg.BreakerTimeout,
		OnStateChange: func(name string, from, to gobreaker.State) {
			slog.With("from", from.String(), "to", to.String()).Warn(fmt.Sprintf("%s changing state", name))
		},
	}
	return &ResilientCache{
		cfg:     cfg,
		redis:   r,
		breaker: gobreaker.NewCircuitBreaker(settings),
	}
}

func (r *ResilientCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd {
	if NoStore(ctx) || !r.cfg.Enabled {
		return redis.NewStatusResult("", nil)
	}

	result, err := r.breaker.Execute(func() (interface{}, error) {
		result := r.redis.Set(ctx, key, value, ttl)
		return result, result.Err()
	})
	if cmd, ok := result.(*redis.StatusCmd); ok {
		return cmd
	}
	return redis.NewStatusResult("", err)
}

func (r *ResilientCache) SetXX(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.BoolCmd {
	if NoStore(ctx) || !r.cfg.Enabled {
		return redis.NewBoolResult(false, nil)
	}

	result, err := r.breaker.Execute(func() (interface{}, error) {
		result := r.redis.SetXX(ctx, key, value, ttl)
		return result, result.Err()
	})
	if cmd, ok := result.(*redis.BoolCmd); ok {
		return cmd
	}
	return redis.NewBoolResult(false, err)
}

func (r *ResilientCache) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.BoolCmd {
	if NoStore(ctx) || !r.cfg.Enabled {
		return redis.NewBoolResult(false, nil)
	}

	result, err := r.breaker.Execute(func() (interface{}, error) {
		result := r.redis.SetNX(ctx, key, value, ttl)
		return result, result.Err()
	})
	if cmd, ok := result.(*redis.BoolCmd); ok {
		return cmd
	}
	return redis.NewBoolResult(false, err)
}

func (r *ResilientCache) Get(ctx context.Context, key string) *redis.StringCmd {
	if NoCache(ctx) || !r.cfg.Enabled {
		return redis.NewStringResult("", nil)
	}

	result, err := r.breaker.Execute(func() (interface{}, error) {
		result := r.redis.Get(ctx, key)
		if result.Err() == redis.Nil {
			return result, nil
		}
		return result, result.Err()
	})
	if cmd, ok := result.(*redis.StringCmd); ok {
		return cmd
	}
	return redis.NewStringResult("", err)
}

func (r *ResilientCache) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	if !r.cfg.Enabled {
		return redis.NewIntResult(0, nil)
	}

	result, err := r.breaker.Execute(func() (interface{}, error) {
		result := r.redis.Del(ctx, keys...)
		return result, result.Err()
	})
	if cmd, ok := result.(*redis.IntCmd); ok {
		return cmd
	}
	return redis.NewIntResult(0, err)
}
