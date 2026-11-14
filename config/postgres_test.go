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

func TestPostgresConfig_connectionURL(t *testing.T) {
	config := PostgresConfig{
		Database: "catalogs",
		Port:     5432,
		User:     "postgres",
		Host:     "localhost",
		Password: "p@ss word",
		SSLMode:  "disable",
	}

	want := "postgres://postgres:p%40ss%20word@localhost:5432/catalogs?sslmode=disable"
	if got := config.ConnectionURL(); got != want {
		t.Errorf("PostgresConfig.connectionURL() = %v, want %v", got, want)
	}
}

func TestPostgresConfig_resolve(t *testing.T) {
	defaults := PostgresConfig{
		User:     "postgres",
		Port:     5432,
		Database: "postgres",
		Host:     "localhost",
		SSLMode:  "require",
	}
	tests := []struct {
		name      string
		config    PostgresConfig
		expected  PostgresConfig
		wantError bool
	}{
		{
			name:     "no url leaves fields",
			config:   defaults,
			expected: defaults,
		},
		{
			name: "url overrides every field",
			config: PostgresConfig{
				URL:      "postgresql://test-user:p%40ss@db.example.com:6543/app?sslmode=disable",
				User:     "postgres",
				Port:     5432,
				Database: "postgres",
				Host:     "localhost",
				SSLMode:  "require",
			},
			expected: PostgresConfig{
				URL:      "postgresql://test-user:p%40ss@db.example.com:6543/app?sslmode=disable",
				User:     "test-user",
				Password: "p@ss",
				Port:     6543,
				Database: "app",
				Host:     "db.example.com",
				SSLMode:  "disable",
			},
		},
		{
			name: "omitted url parameters keep current values",
			config: PostgresConfig{
				URL:      "postgres://db.example.com",
				User:     "postgres",
				Port:     5432,
				Database: "postgres",
				Host:     "localhost",
				SSLMode:  "require",
				Password: "test-password",
			},
			expected: PostgresConfig{
				URL:      "postgres://db.example.com",
				User:     "postgres",
				Port:     5432,
				Database: "postgres",
				Host:     "db.example.com",
				SSLMode:  "require",
				Password: "test-password",
			},
		},
		{
			name:      "unsupported scheme",
			config:    PostgresConfig{URL: "mysql://db.example.com/app"},
			expected:  PostgresConfig{URL: "mysql://db.example.com/app"},
			wantError: true,
		},
		{
			name:      "invalid port",
			config:    PostgresConfig{URL: "postgres://db.example.com:abc/app"},
			expected:  PostgresConfig{URL: "postgres://db.example.com:abc/app"},
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.config
			err := cfg.resolve()

			if !reflect.DeepEqual(cfg, tt.expected) {
				t.Errorf("resolve() config = %+v, want %+v", cfg, tt.expected)
			}
			if (err != nil) != tt.wantError {
				t.Errorf("resolve() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
