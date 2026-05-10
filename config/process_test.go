package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestProcess(t *testing.T) {
	wd, _ := os.Getwd()
	tmpFile := filepath.Join(wd, "testdata/bad_dir/invalid_file.env")
	f, _ := os.Create(tmpFile)
	if err := f.Chmod(0333); err != nil {
		t.Fatal(err)
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(tmpFile)

	type testSubConfig struct {
		Value      int     `envconfig:"TEST_VALUE" default:"1"`
		Multiplier float64 `envconfig:"TEST_MULTIPLIER"`
	}
	type testConfig struct {
		Environment string        `envconfig:"TEST_ENVIRONMENT" default:"development"`
		Host        string        `envconfig:"TEST_HOST" default:"localhost"`
		User        string        `envconfig:"TEST_USER"`
		Password    string        `envconfig:"TEST_PASSWORD"`
		Timeout     time.Duration `envconfig:"TEST_TIMEOUT" default:"30s"`
		Sub         testSubConfig
	}
	tests := []struct {
		name    string
		dir     string
		want    testConfig
		wantErr bool
	}{
		{
			name: "loaded",
			dir:  "testdata/good_dir",
			want: testConfig{
				Environment: "test",
				Host:        "localhost",
				User:        "test-user",
				Password:    "",
				Timeout:     30 * time.Second,
				Sub: testSubConfig{
					Value:      1,
					Multiplier: 7.77,
				},
			},
		},
		{
			name:    "bad",
			dir:     "testdata/bad_dir",
			wantErr: true,
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			got := testConfig{}
			if err := Process(tt.dir, &got); (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Process() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadEnv(t *testing.T) {
	wd, _ := os.Getwd()
	tmpFile := filepath.Join(wd, "testdata/bad_dir/invalid_file.env")
	f, _ := os.Create(tmpFile)
	if err := f.Chmod(0333); err != nil {
		t.Fatal(err)
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(tmpFile)

	tests := []struct {
		name    string
		dir     string
		want    map[string]string
		wantErr bool
	}{
		{
			name: "good",
			dir:  filepath.Join(wd, "testdata/good_dir"),
			want: map[string]string{
				"TEST_ENVIRONMENT": "test",
				"TEST_USER":        "test-user",
				"TEST_MULTIPLIER":  "7.77",
			},
			wantErr: false,
		},
		{
			name: "goodWith/",
			dir:  filepath.Join(wd, "testdata/good_dir/"),
			want: map[string]string{
				"TEST_ENVIRONMENT": "test",
				"TEST_USER":        "test-user",
				"TEST_MULTIPLIER":  "7.77",
			},
		},
		{
			name:    "empty",
			dir:     "",
			wantErr: false,
		},
		{
			name:    "invalid",
			dir:     "testdata/invalid",
			wantErr: true,
		},
		{
			name:    "bad",
			dir:     "testdata/bad_dir",
			wantErr: true,
		},
		{
			name:    "badWith/",
			dir:     "testdata/bad_dir/",
			wantErr: true,
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			if err := LoadEnv(tt.dir); (err != nil) != tt.wantErr {
				t.Errorf("LoadEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for k, v := range tt.want {
				if !reflect.DeepEqual(os.Getenv(k), v) {
					t.Errorf("LoadEnv().%s = %v, want %v", k, os.Getenv(k), v)
				}
			}
		})
	}
}

func TestNew(t *testing.T) {
	type testNewConfig struct {
		Environment string            `envconfig:"ENVIRONMENT" default:"development" help:"deployment environment"`
		Host        string            `envconfig:"HOST" default:"localhost" help:"server host"`
		Port        int               `envconfig:"PORT" default:"8080" help:"server port"`
		Debug       bool              `envconfig:"DEBUG" help:"enable debug mode"`
		Rate        float64           `envconfig:"RATE" default:"1.5" help:"rate limit"`
		Timeout     time.Duration     `envconfig:"TIMEOUT" default:"30s" help:"request timeout"`
		Secret      string            `envconfig:"SECRET" default:"shh" arg:"-"`
		Items       []string          `envconfig:"ITEMS"`
		Tags        map[string]string `envconfig:"TAGS"`
	}
	tests := []struct {
		name     string
		envSetup func(t *testing.T)
		args     []string
		opts     []Option
		want     testNewConfig
		wantErr  bool
	}{
		{
			name:     "defaults only",
			envSetup: func(t *testing.T) {},
			want: testNewConfig{
				Environment: "development",
				Host:        "localhost",
				Port:        8080,
				Rate:        1.5,
				Timeout:     30 * time.Second,
				Secret:      "shh",
			},
		},
		{
			name: "env vars only",
			envSetup: func(t *testing.T) {
				t.Setenv("ENVIRONMENT", "test")
				t.Setenv("HOST", "test-host")
				t.Setenv("PORT", "3000")
				t.Setenv("DEBUG", "true")
				t.Setenv("SECRET", "test-secret")
				t.Setenv("ITEMS", "env-item-1,env-item-2")
				t.Setenv("TAGS", "region:us")
			},
			want: testNewConfig{
				Environment: "test",
				Host:        "test-host",
				Port:        3000,
				Debug:       true,
				Rate:        1.5,
				Timeout:     30 * time.Second,
				Secret:      "test-secret",
				Items:       []string{"env-item-1", "env-item-2"},
				Tags:        map[string]string{"region": "us"},
			},
		},
		{
			name: "file overrides env var",
			envSetup: func(t *testing.T) {
				t.Setenv("ENVIRONMENT", "test")
				t.Setenv("HOST", "test-host")
				t.Setenv("PORT", "3000")
			},
			opts: []Option{WithFile("testdata/global.json")},
			want: testNewConfig{
				Environment: "global",
				Host:        "global-host",
				Port:        9090,
				Rate:        2.5,
				Timeout:     30 * time.Second,
				Secret:      "shh",
				Items:       []string{"global-item-1", "global-item-2"},
				Tags:        map[string]string{"source": "global", "tier": "free"},
			},
		},
		{
			name: "file replaces slice from env",
			envSetup: func(t *testing.T) {
				t.Setenv("ITEMS", "env-item-1,env-item-2")
			},
			opts: []Option{WithFile("testdata/global.json")},
			want: testNewConfig{
				Environment: "global",
				Host:        "global-host",
				Port:        9090,
				Rate:        2.5,
				Timeout:     30 * time.Second,
				Secret:      "shh",
				Items:       []string{"global-item-1", "global-item-2"},
				Tags:        map[string]string{"source": "global", "tier": "free"},
			},
		},
		{
			name: "file merges map keys from env",
			envSetup: func(t *testing.T) {
				t.Setenv("TAGS", "region:us")
			},
			opts: []Option{WithFile("testdata/global.json")},
			want: testNewConfig{
				Environment: "global",
				Host:        "global-host",
				Port:        9090,
				Rate:        2.5,
				Timeout:     30 * time.Second,
				Secret:      "shh",
				Items:       []string{"global-item-1", "global-item-2"},
				Tags:        map[string]string{"region": "us", "source": "global", "tier": "free"},
			},
		},
		{
			name:     "later file overrides earlier file",
			envSetup: func(t *testing.T) {},
			opts: []Option{
				WithFile("testdata/global.json"),
				WithFile("testdata/local.yaml"),
			},
			want: testNewConfig{
				Environment: "local",
				Host:        "local-host",
				Port:        9090,
				Rate:        2.5,
				Timeout:     30 * time.Second,
				Secret:      "shh",
				Items:       []string{"local-item-1"},
				Tags:        map[string]string{"source": "local", "tier": "free"},
			},
		},
		{
			name: "env var used when no file overrides it",
			envSetup: func(t *testing.T) {
				t.Setenv("TIMEOUT", "1m")
			},
			opts: []Option{WithFile("testdata/local.yaml")},
			want: testNewConfig{
				Environment: "local",
				Host:        "local-host",
				Port:        8080,
				Rate:        1.5,
				Timeout:     time.Minute,
				Secret:      "shh",
				Items:       []string{"local-item-1"},
				Tags:        map[string]string{"source": "local"},
			},
		},
		{
			name:     "cli args only",
			envSetup: func(t *testing.T) {},
			args:     []string{"--environment", "cli", "--port", "4000"},
			want: testNewConfig{
				Environment: "cli",
				Host:        "localhost",
				Port:        4000,
				Rate:        1.5,
				Timeout:     30 * time.Second,
				Secret:      "shh",
			},
		},
		{
			name: "cli args override env vars",
			envSetup: func(t *testing.T) {
				t.Setenv("ENVIRONMENT", "test")
				t.Setenv("HOST", "test-host")
				t.Setenv("SECRET", "test-secret")
			},
			args: []string{"--environment", "cli"},
			want: testNewConfig{
				Environment: "cli",
				Host:        "test-host",
				Port:        8080,
				Rate:        1.5,
				Timeout:     30 * time.Second,
				Secret:      "test-secret",
			},
		},
		{
			name: "cli args override file values",
			envSetup: func(t *testing.T) {
				t.Setenv("PORT", "3000")
			},
			args: []string{"--environment", "cli"},
			opts: []Option{
				WithFile("testdata/global.json"),
			},
			want: testNewConfig{
				Environment: "cli",
				Host:        "global-host",
				Port:        9090,
				Rate:        2.5,
				Timeout:     30 * time.Second,
				Secret:      "shh",
				Items:       []string{"global-item-1", "global-item-2"},
				Tags:        map[string]string{"source": "global", "tier": "free"},
			},
		},
		{
			name: "cli args do not reset unset fields",
			envSetup: func(t *testing.T) {
				t.Setenv("ENVIRONMENT", "test")
				t.Setenv("HOST", "test-host")
				t.Setenv("PORT", "3000")
				t.Setenv("DEBUG", "true")
				t.Setenv("RATE", "5.0")
				t.Setenv("TIMEOUT", "1m")
				t.Setenv("SECRET", "test-secret")
			},
			args: []string{"--port", "4000"},
			want: testNewConfig{
				Environment: "test",
				Host:        "test-host",
				Port:        4000,
				Debug:       true,
				Rate:        5.0,
				Timeout:     time.Minute,
				Secret:      "test-secret",
			},
		},
		{
			name: "cli default tags do not override env values",
			envSetup: func(t *testing.T) {
				t.Setenv("ENVIRONMENT", "production")
				t.Setenv("HOST", "prodhost")
				t.Setenv("PORT", "443")
				t.Setenv("RATE", "10.0")
				t.Setenv("TIMEOUT", "2m")
			},
			args: []string{},
			want: testNewConfig{
				Environment: "production",
				Host:        "prodhost",
				Port:        443,
				Rate:        10.0,
				Timeout:     2 * time.Minute,
				Secret:      "shh",
			},
		},
		{
			name: "cli args replace slice",
			envSetup: func(t *testing.T) {
				t.Setenv("ITEMS", "env-item-1,env-item-2")
			},
			args: []string{"--items", "cli-item-1", "cli-item-2"},
			opts: []Option{
				WithFile("testdata/global.json"),
			},
			want: testNewConfig{
				Environment: "global",
				Host:        "global-host",
				Port:        9090,
				Rate:        2.5,
				Timeout:     30 * time.Second,
				Secret:      "shh",
				Items:       []string{"cli-item-1", "cli-item-2"},
				Tags:        map[string]string{"source": "global", "tier": "free"},
			},
		},
		{
			name:     "bool flag without value",
			envSetup: func(t *testing.T) {},
			args:     []string{"--debug"},
			want: testNewConfig{
				Environment: "development",
				Host:        "localhost",
				Port:        8080,
				Debug:       true,
				Rate:        1.5,
				Timeout:     30 * time.Second,
				Secret:      "shh",
			},
		},
		{
			name:     "duration flag",
			envSetup: func(t *testing.T) {},
			args:     []string{"--timeout", "5m"},
			want: testNewConfig{
				Environment: "development",
				Host:        "localhost",
				Port:        8080,
				Rate:        1.5,
				Timeout:     5 * time.Minute,
				Secret:      "shh",
			},
		},
		{
			name:     "arg excluded field is ignored",
			envSetup: func(t *testing.T) {},
			args:     []string{"--secret", "hacked"},
			want: testNewConfig{
				Environment: "development",
				Host:        "localhost",
				Port:        8080,
				Rate:        1.5,
				Timeout:     30 * time.Second,
				Secret:      "shh",
			},
		},
		{
			name:     "missing file skipped",
			envSetup: func(t *testing.T) {},
			opts:     []Option{WithFile("testdata/nonexistent.yaml")},
			want: testNewConfig{
				Environment: "development",
				Host:        "localhost",
				Port:        8080,
				Rate:        1.5,
				Timeout:     30 * time.Second,
				Secret:      "shh",
			},
		},
		{
			name:     "bad json file",
			envSetup: func(t *testing.T) {},
			opts:     []Option{WithFile("testdata/bad.json")},
			wantErr:  true,
		},
		{
			name:     "unsupported file format",
			envSetup: func(t *testing.T) {},
			opts:     []Option{WithFile("testdata/ingored_file")},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.envSetup(t)

			orig := os.Args
			os.Args = append([]string{"test"}, tt.args...)
			t.Cleanup(func() { os.Args = orig })

			got, err := New[testNewConfig](tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("invalid type", func(t *testing.T) {
		_, err := New[map[string]string]()
		if err == nil {
			t.Error("New[map[string]string]() should have failed")
		}
	})
}
