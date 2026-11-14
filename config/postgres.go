package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/jmoiron/sqlx/splunksqlx"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type PostgresConfig struct {
	// URL is a full connection URL (postgresql://user:password@host/db?sslmode=require).
	// NewPostgresClient copies the parameters the URL specifies over the discrete fields.
	URL             string        `envconfig:"POSTGRES_URL"`
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

// resolve populates the discrete connection fields from URL when it is set. Parameters the
// URL omits keep their current values.
func (p *PostgresConfig) resolve() error {
	if p.URL == "" {
		return nil
	}
	u, err := url.Parse(p.URL)
	if err != nil {
		// url.Error embeds the raw URL, which carries the password.
		return fmt.Errorf("failed to parse postgres url: %w", errors.Unwrap(err))
	}
	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return fmt.Errorf("unsupported postgres url scheme %q", u.Scheme)
	}

	if host := u.Hostname(); host != "" {
		p.Host = host
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil {
			return fmt.Errorf("invalid postgres url port %q: %w", port, err)
		}
		p.Port = n
	}
	if u.User != nil {
		p.User = u.User.Username()
		if password, ok := u.User.Password(); ok {
			p.Password = password
		}
	}
	if db := strings.TrimPrefix(u.Path, "/"); db != "" {
		p.Database = db
	}
	if mode := u.Query().Get("sslmode"); mode != "" {
		p.SSLMode = mode
	}
	return nil
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

// ConnectionURL is the URL form of the connection string, the only form the
// Splunk instrumentation can parse a peer host and port from.
func (p PostgresConfig) ConnectionURL() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(p.User, p.Password),
		Host:     net.JoinHostPort(p.Host, strconv.Itoa(p.Port)),
		Path:     p.Database,
		RawQuery: url.Values{"sslmode": {p.SSLMode}}.Encode(),
	}
	return u.String()
}

func NewPostgresClient(cfg PostgresConfig) (*sqlx.DB, error) {
	if err := cfg.resolve(); err != nil {
		return nil, err
	}
	postgresAttrs := []attribute.KeyValue{attribute.String("service.name", "postgres")}
	db, err := splunksqlx.Connect("postgres", cfg.ConnectionURL(),
		splunksql.WithAttributes(postgresAttrs),
	)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxIdleTime(cfg.MaxIdleDuration)
	db.SetMaxOpenConns(cfg.MaxConns)
	db.SetConnMaxLifetime(cfg.MaxConnDuration)

	if _, err := otelsql.RegisterDBStatsMetrics(db.DB, otelsql.WithAttributes(semconv.DBSystemPostgreSQL)); err != nil {
		return nil, err
	}

	return db, nil
}
