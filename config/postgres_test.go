package config

import (
	"reflect"
	"testing"
)

func TestPostgresConfig_ConnectionString(t *testing.T) {
	config := PostgresConfig{
		Database: "catalogs",
		Port:     5432,
		User:     "postgres",
		Host:     "localhost",
		Password: "password",
		SSLMode:  "disable",
	}

	want := "host=localhost port=5432 user=postgres password=password dbname=catalogs sslmode=disable"
	if got := config.ConnectionString(); !reflect.DeepEqual(got, want) {
		t.Errorf("PostgresConfig.ConnectionString() = %v, want %v", got, want)
	}
}
