package config

import (
	"path/filepath"
	"testing"
)

func TestNewAWSConfig(t *testing.T) {
	type result struct {
		Region string
		Key    string
		Secret string
	}
	tests := []struct {
		name      string
		cfg       AWSConfig
		envSetup  func(t *testing.T)
		expected  result
		wantError bool
	}{
		{
			name: "overrides",
			cfg: AWSConfig{
				Region:    "us-west-2",
				AccessKey: "test-key",
				SecretKey: "test-secret",
			},
			envSetup: func(t *testing.T) {},
			expected: result{Region: "us-west-2", Key: "test-key", Secret: "test-secret"},
		},
		{
			name: "falls back to sdk environment",
			cfg:  AWSConfig{},
			envSetup: func(t *testing.T) {
				t.Setenv("AWS_REGION", "eu-west-1")
				t.Setenv("AWS_ACCESS_KEY_ID", "test-env-key")
				t.Setenv("AWS_SECRET_ACCESS_KEY", "test-env-secret")
			},
			expected: result{Region: "eu-west-1", Key: "test-env-key", Secret: "test-env-secret"},
		},
		{
			name: "static credentials need both keys",
			cfg:  AWSConfig{AccessKey: "test-key"},
			envSetup: func(t *testing.T) {
				t.Setenv("AWS_ACCESS_KEY_ID", "test-env-key")
				t.Setenv("AWS_SECRET_ACCESS_KEY", "test-env-secret")
			},
			expected: result{Key: "test-env-key", Secret: "test-env-secret"},
		},
		{
			name:      "unknown profile",
			cfg:       AWSConfig{Profile: "test-profile"},
			envSetup:  func(t *testing.T) {},
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("AWS_CONFIG_FILE", filepath.Join(dir, "config"))
			t.Setenv("AWS_SHARED_CREDENTIALS_FILE", filepath.Join(dir, "credentials"))
			t.Setenv("AWS_REGION", "")
			t.Setenv("AWS_DEFAULT_REGION", "")
			t.Setenv("AWS_ACCESS_KEY_ID", "")
			t.Setenv("AWS_SECRET_ACCESS_KEY", "")
			t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
			tt.envSetup(t)

			awsCfg, err := NewAWSConfig(t.Context(), tt.cfg)

			if (err != nil) != tt.wantError {
				t.Fatalf("NewAWSConfig() error = %v, wantError %v", err, tt.wantError)
			}
			if tt.wantError {
				return
			}
			creds, err := awsCfg.Credentials.Retrieve(t.Context())
			if err != nil {
				t.Fatalf("Credentials.Retrieve() error = %v", err)
			}
			got := result{
				Region: awsCfg.Region,
				Key:    creds.AccessKeyID,
				Secret: creds.SecretAccessKey,
			}
			if got != tt.expected {
				t.Errorf("NewAWSConfig() = %+v, want %+v", got, tt.expected)
			}
		})
	}
}
