package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/marcw/cachecontrol"
	"github.com/sony/gobreaker"
)

func TestRedisBreaker_Set(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		cfg          Config
		redisSetup   func(*redis.Client)
		breakerSetup func(*gobreaker.CircuitBreaker)
		wantValue    string
		wantErr      bool
	}{
		{
			name:         "cache control: no store",
			ctx:          context.WithValue(t.Context(), CacheControlContextKey, cachecontrol.Parse("no-store")),
			cfg:          Config{Enabled: true},
			redisSetup:   func(rc *redis.Client) {},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    "",
			wantErr:      false,
		},
		{
			name:         "redis disabled",
			ctx:          t.Context(),
			cfg:          Config{Enabled: false},
			redisSetup:   func(rc *redis.Client) {},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    "",
			wantErr:      false,
		},
		{
			name:         "redis operation successful",
			ctx:          t.Context(),
			cfg:          Config{Enabled: true},
			redisSetup:   func(rc *redis.Client) {},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    "OK",
			wantErr:      false,
		},
		{
			name:         "redis closed",
			ctx:          t.Context(),
			cfg:          Config{Enabled: true},
			redisSetup:   func(rc *redis.Client) { _ = rc.Close() },
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    "",
			wantErr:      true,
		},
		{
			name:       "breaker open",
			ctx:        t.Context(),
			cfg:        Config{Enabled: true},
			redisSetup: func(rc *redis.Client) {},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {
				for i := 0; i < 10; i++ {
					_, _ = b.Execute(func() (interface{}, error) { return nil, errors.New("test-error") })
				}
			},
			wantValue: "",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := miniredis.Run()
			rc := redis.NewClient(&redis.Options{Addr: s.Addr()})
			tt.redisSetup(rc)
			r := NewRedisBreaker(tt.cfg, rc)
			tt.breakerSetup(r.breaker)
			value, err := r.Set(tt.ctx, "test-key", "test-value", time.Hour).Result()

			if value != tt.wantValue {
				t.Errorf("RedisBreaker.Set().Result() value = %v, want %v", value, tt.wantValue)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("RedisBreaker.Set().Result() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRedisBreaker_SetXX(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		cfg          Config
		redisSetup   func(*redis.Client)
		breakerSetup func(*gobreaker.CircuitBreaker)
		wantValue    bool
		wantErr      bool
	}{
		{
			name:         "cache control: no store",
			ctx:          context.WithValue(t.Context(), CacheControlContextKey, cachecontrol.Parse("no-store")),
			cfg:          Config{Enabled: true},
			redisSetup:   func(rc *redis.Client) {},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    false,
			wantErr:      false,
		},
		{
			name:         "redis disabled",
			ctx:          t.Context(),
			cfg:          Config{Enabled: false},
			redisSetup:   func(rc *redis.Client) {},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    false,
			wantErr:      false,
		},
		{
			name:         "redis operation successful",
			ctx:          t.Context(),
			cfg:          Config{Enabled: true},
			redisSetup:   func(rc *redis.Client) {},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    false,
			wantErr:      false,
		},
		{
			name:         "redis closed",
			ctx:          t.Context(),
			cfg:          Config{Enabled: true},
			redisSetup:   func(rc *redis.Client) { _ = rc.Close() },
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    false,
			wantErr:      true,
		},
		{
			name:       "breaker open",
			ctx:        t.Context(),
			cfg:        Config{Enabled: true},
			redisSetup: func(rc *redis.Client) {},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {
				for i := 0; i < 10; i++ {
					_, _ = b.Execute(func() (interface{}, error) { return nil, errors.New("test-error") })
				}
			},
			wantValue: false,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := miniredis.Run()
			rc := redis.NewClient(&redis.Options{Addr: s.Addr()})
			tt.redisSetup(rc)
			r := NewRedisBreaker(tt.cfg, rc)
			tt.breakerSetup(r.breaker)
			value, err := r.SetXX(tt.ctx, "test-key", "test-value", time.Hour).Result()

			if value != tt.wantValue {
				t.Errorf("RedisBreaker.SetXX().Result() value = %v, want %v", value, tt.wantValue)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("RedisBreaker.SetXX().Result() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRedisBreaker_SetNX(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		cfg          Config
		redisSetup   func(*redis.Client)
		breakerSetup func(*gobreaker.CircuitBreaker)
		wantValue    bool
		wantErr      bool
	}{
		{
			name:         "cache control: no store",
			ctx:          context.WithValue(t.Context(), CacheControlContextKey, cachecontrol.Parse("no-store")),
			cfg:          Config{Enabled: true},
			redisSetup:   func(rc *redis.Client) {},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    false,
			wantErr:      false,
		},
		{
			name:         "redis disabled",
			ctx:          t.Context(),
			cfg:          Config{Enabled: false},
			redisSetup:   func(rc *redis.Client) {},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    false,
			wantErr:      false,
		},
		{
			name:         "redis operation successful",
			ctx:          t.Context(),
			cfg:          Config{Enabled: true},
			redisSetup:   func(rc *redis.Client) {},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    true,
			wantErr:      false,
		},
		{
			name:         "redis closed",
			ctx:          t.Context(),
			cfg:          Config{Enabled: true},
			redisSetup:   func(rc *redis.Client) { _ = rc.Close() },
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    false,
			wantErr:      true,
		},
		{
			name:       "breaker open",
			ctx:        t.Context(),
			cfg:        Config{Enabled: true},
			redisSetup: func(rc *redis.Client) {},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {
				for i := 0; i < 10; i++ {
					_, _ = b.Execute(func() (interface{}, error) { return nil, errors.New("test-error") })
				}
			},
			wantValue: false,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := miniredis.Run()
			rc := redis.NewClient(&redis.Options{Addr: s.Addr()})
			tt.redisSetup(rc)
			r := NewRedisBreaker(tt.cfg, rc)
			tt.breakerSetup(r.breaker)
			value, err := r.SetNX(tt.ctx, "test-key", "test-value", time.Hour).Result()

			if value != tt.wantValue {
				t.Errorf("RedisBreaker.SetNX().Result() value = %v, want %v", value, tt.wantValue)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("RedisBreaker.SetNX().Result() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRedisBreaker_Get(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		cfg          Config
		redisSetup   func(*redis.Client)
		breakerSetup func(*gobreaker.CircuitBreaker)
		wantValue    string
		wantErr      bool
	}{
		{
			name:         "cache control: no cache",
			ctx:          context.WithValue(t.Context(), CacheControlContextKey, cachecontrol.Parse("no-cache")),
			cfg:          Config{Enabled: true},
			redisSetup:   func(rc *redis.Client) {},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    "",
			wantErr:      false,
		},
		{
			name:         "redis disabled",
			ctx:          t.Context(),
			cfg:          Config{Enabled: false},
			redisSetup:   func(rc *redis.Client) {},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    "",
			wantErr:      false,
		},
		{
			name: "redis operation successful",
			ctx:  t.Context(),
			cfg:  Config{Enabled: true},
			redisSetup: func(rc *redis.Client) {
				_ = rc.Set(t.Context(), "test-key", "test-value", time.Hour)
			},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    "test-value",
			wantErr:      false,
		},
		{
			name: "redis operation - missed",
			ctx:  t.Context(),
			cfg:  Config{Enabled: true},
			redisSetup: func(rc *redis.Client) {
			},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    "",
			wantErr:      true,
		},
		{
			name:         "redis closed",
			ctx:          t.Context(),
			cfg:          Config{Enabled: true},
			redisSetup:   func(rc *redis.Client) { _ = rc.Close() },
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    "",
			wantErr:      true,
		},
		{
			name:       "breaker open",
			ctx:        t.Context(),
			cfg:        Config{Enabled: true},
			redisSetup: func(rc *redis.Client) {},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {
				for i := 0; i < 10; i++ {
					_, _ = b.Execute(func() (interface{}, error) { return nil, errors.New("test-error") })
				}
			},
			wantValue: "",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := miniredis.Run()
			rc := redis.NewClient(&redis.Options{Addr: s.Addr()})
			tt.redisSetup(rc)
			r := NewRedisBreaker(tt.cfg, rc)
			tt.breakerSetup(r.breaker)
			for i := 0; i < 10; i++ {
				_, _ = r.Get(tt.ctx, "test-key").Result()
			}
			value, err := r.Get(tt.ctx, "test-key").Result()

			if value != tt.wantValue {
				t.Errorf("RedisBreaker.Get().Result() value = %v, want %v", value, tt.wantValue)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("RedisBreaker.Get().Result() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRedisBreaker_Del(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		cfg          Config
		redisSetup   func(*redis.Client)
		breakerSetup func(*gobreaker.CircuitBreaker)
		wantValue    int64
		wantErr      bool
	}{
		{
			name:         "redis disabled",
			ctx:          t.Context(),
			cfg:          Config{Enabled: false},
			redisSetup:   func(rc *redis.Client) {},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    0,
			wantErr:      false,
		},
		{
			name: "redis operation successful",
			ctx:  t.Context(),
			cfg:  Config{Enabled: true},
			redisSetup: func(rc *redis.Client) {
				_ = rc.Set(t.Context(), "test-key", "test-value", time.Hour)
			},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    1,
			wantErr:      false,
		},
		{
			name:         "redis operation successful - missing value",
			ctx:          t.Context(),
			cfg:          Config{Enabled: true},
			redisSetup:   func(rc *redis.Client) {},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    0,
			wantErr:      false,
		},
		{
			name:         "redis closed",
			ctx:          t.Context(),
			cfg:          Config{Enabled: true},
			redisSetup:   func(rc *redis.Client) { _ = rc.Close() },
			breakerSetup: func(b *gobreaker.CircuitBreaker) {},
			wantValue:    0,
			wantErr:      true,
		},
		{
			name:       "breaker open",
			ctx:        t.Context(),
			cfg:        Config{Enabled: true},
			redisSetup: func(rc *redis.Client) {},
			breakerSetup: func(b *gobreaker.CircuitBreaker) {
				for i := 0; i < 10; i++ {
					_, _ = b.Execute(func() (interface{}, error) { return nil, errors.New("test-error") })
				}
			},
			wantValue: 0,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := miniredis.Run()
			rc := redis.NewClient(&redis.Options{Addr: s.Addr()})
			tt.redisSetup(rc)
			r := NewRedisBreaker(tt.cfg, rc)
			tt.breakerSetup(r.breaker)
			value, err := r.Del(tt.ctx, "test-key").Result()

			if value != tt.wantValue {
				t.Errorf("RedisBreaker.Del().Result() value = %v, want %v", value, tt.wantValue)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("RedisBreaker.Del().Result() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
