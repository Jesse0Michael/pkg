package config

import (
	"fmt"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/jmoiron/sqlx/splunksqlx"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

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

func NewPostgresClient(cfg PostgresConfig) (*sqlx.DB, error) {
	postgresAttrs := []attribute.KeyValue{attribute.String("service.name", "postgres")}
	db, err := splunksqlx.Connect("postgres", cfg.ConnectionString(),
		splunksql.WithAttributes(postgresAttrs),
	)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxIdleTime(cfg.MaxIdleDuration)
	db.SetMaxOpenConns(cfg.MaxConns)
	db.SetConnMaxLifetime(cfg.MaxConnDuration)

	if err := otelsql.RegisterDBStatsMetrics(db.DB, otelsql.WithAttributes(semconv.DBSystemPostgreSQL)); err != nil {
		return nil, err
	}

	return db, nil
}
