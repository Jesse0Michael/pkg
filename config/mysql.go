package config

import (
	"database/sql"
	"fmt"

	"github.com/XSAM/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

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

func NewMysqlClient(cfg MysqlConfig) (*sql.DB, error) {
	db, err := otelsql.Open("mysql", cfg.ConnectionString(), otelsql.WithAttributes(
		semconv.DBSystemMySQL,
	))
	if err != nil {
		return nil, err
	}

	if err = otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(
		semconv.DBSystemMySQL,
	)); err != nil {
		return nil, err
	}
	return db, nil
}
