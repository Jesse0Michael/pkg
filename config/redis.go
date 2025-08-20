package config

import (
	"crypto/tls"
	"time"

	"github.com/go-redis/redis/extra/redisotel/v8"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/attribute"
)

type RedisConfig struct {
	Addr         string        `envconfig:"REDIS_ADDR" default:"localhost:6379"`
	Password     string        `envconfig:"REDIS_PASSWORD"`
	DB           int           `envconfig:"REDIS_DB"`
	TLS          bool          `envconfig:"REDIS_TLS"`
	ReadTimeout  time.Duration `envconfig:"REDIS_READ_TIMEOUT" default:"1s"`
	WriteTimeout time.Duration `envconfig:"REDIS_WRITE_TIMEOUT" default:"5s"`
}

func NewRedisClient(cfg RedisConfig) *redis.Client {
	opts := &redis.Options{
		Addr:         cfg.Addr,
		DB:           cfg.DB,
		Password:     cfg.Password,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
	if cfg.TLS {
		opts.TLSConfig = &tls.Config{
			InsecureSkipVerify: true, // nolint:gosec
		}
	}
	rc := redis.NewClient(opts)
	rc.AddHook(redisotel.NewTracingHook(redisotel.WithAttributes(attribute.String("service.name", "redis"))))
	return rc
}
