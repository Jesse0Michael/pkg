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
