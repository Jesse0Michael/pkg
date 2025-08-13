package config

import (
	"reflect"
	"testing"
)

func TestMysqlConfig_ConnectionString(t *testing.T) {
	config := MysqlConfig{
		Database: "vehicle",
		Port:     3306,
		User:     "mysql",
		Host:     "mysql",
		Password: "password",
	}

	want := "mysql:password@(mysql:3306)/vehicle"
	if got := config.ConnectionString(); !reflect.DeepEqual(got, want) {
		t.Errorf("New() = %v, want %v", got, want)
	}
}

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
