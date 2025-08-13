package config

import (
	"fmt"
	"time"
)

type AppConfig struct {
	Environment string `envconfig:"ENVIRONMENT"`
	Name        string `envconfig:"APP_NAME"`
	Version     string `envconfig:"VERSION"`
	LogLevel    string `envconfig:"LOG_LEVEL"`
}

type MysqlConfig struct {
	Password string `envconfig:"MYSQL_PASSWORD"`
	User     string `envconfig:"MYSQL_USER"     default:"mysql"`
	Port     int    `envconfig:"MYSQL_PORT"     default:"3306"`
	Database string `envconfig:"MYSQL_DB"`
	Host     string `envconfig:"MYSQL_HOST"     default:"localhost"`
}

func (m MysqlConfig) ConnectionString() string {
	return fmt.Sprintf(
		"%s:%s@(%s:%d)/%s",
		m.User,
		m.Password,
		m.Host,
		m.Port,
		m.Database,
	)
}

type PostgresConfig struct {
	Password        string        `envconfig:"POSTGRES_PASSWORD"`
	User            string        `envconfig:"POSTGRES_USER"     default:"postgres"`
	Port            int           `envconfig:"POSTGRES_PORT"     default:"5432"`
	Database        string        `envconfig:"POSTGRES_DB"       default:"postgres"`
	Host            string        `envconfig:"POSTGRES_HOST"     default:"localhost"`
	SSLMode         string        `envconfig:"POSTGRES_SSLMODE"  default:"require"`
	MaxConns        int           `envconfig:"POSTGRES_MAX_CONNS"`
	MaxConnDuration time.Duration `envconfig:"POSTGRES_MAX_CONN_DURATION"`
	MaxIdleConns    int           `envconfig:"POSTGRES_MAX_IDLE_CONNS"`
	MaxIdleDuration time.Duration `envconfig:"POSTGRES_MAX_IDLE_DURATION"`
}

func (p PostgresConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host,
		p.Port,
		p.User,
		p.Password,
		p.Database,
		p.SSLMode)
}

type RedisConfig struct {
	Addr         string        `envconfig:"REDIS_ADDR" default:"localhost:6379"`
	Password     string        `envconfig:"REDIS_PASSWORD"`
	DB           int           `envconfig:"REDIS_DB"`
	TLS          bool          `envconfig:"REDIS_TLS"`
	ReadTimeout  time.Duration `envconfig:"REDIS_READ_TIMEOUT" default:"1s"`
	WriteTimeout time.Duration `envconfig:"REDIS_WRITE_TIMEOUT" default:"5s"`
}
