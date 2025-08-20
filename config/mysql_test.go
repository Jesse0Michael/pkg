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
