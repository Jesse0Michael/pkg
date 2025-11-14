package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// New creates and processes a generic type T with envconfig environment variables
func New[T any]() (T, error) {
	var cfg T
	err := envconfig.Process("", &cfg)
	return cfg, err
}

// Process loads environment variables from .env files in the specified directory
// and then processes the environment variables into the provided configuration struct
func Process(envDir string, cfg interface{}) error {
	if err := LoadEnv(envDir); err != nil {
		return err
	}

	return envconfig.Process("", cfg)
}

// LoadEnv loads environment variables from .env files in the specified directory
func LoadEnv(envDir string) error {
	if envDir != "" {
		files, err := os.ReadDir(envDir)
		if err != nil {
			return err
		}
		for i := range files {
			f := files[i]

			fullFilePath := filepath.Join(envDir, f.Name())
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".env") {
				continue
			}

			if err := godotenv.Load(fullFilePath); err != nil {
				return err
			}
		}
	}
	return nil
}
