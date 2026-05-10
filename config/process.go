package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type options struct {
	prefix string
	files  []string
}

// Option configures the New function.
type Option func(*options)

// WithPrefix sets the envconfig prefix for env var resolution.
func WithPrefix(prefix string) Option {
	return func(o *options) {
		o.prefix = prefix
	}
}

// WithFile adds a config file to be loaded. Files are applied in order,
// so later files override earlier ones. Supports JSON and YAML.
// Files that don't exist are silently skipped.
func WithFile(path string) Option {
	return func(o *options) {
		o.files = append(o.files, path)
	}
}

// Initializer can be implemented by config structs to set dynamic initial values
// before other sources (env vars, files, CLI args) are applied.
type Initializer interface {
	Init()
}

// New creates a config of type T by layering sources in order:
// init < struct tag defaults + env vars < files (in order) < CLI args.
func New[T any](opts ...Option) (T, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	var cfg T
	if d, ok := any(&cfg).(Initializer); ok {
		d.Init()
	}

	if err := envconfig.Process(o.prefix, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to process env config: %w", err)
	}

	for _, path := range o.files {
		if err := loadFile(path, &cfg); err != nil {
			return cfg, fmt.Errorf("failed to load config file %s: %w", path, err)
		}
	}

	if err := parseArgs(os.Args[1:], &cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse CLI args: %w", err)
	}

	return cfg, nil
}

func parseArgs(args []string, cfg any) error {
	p, err := arg.NewParser(arg.Config{IgnoreEnv: true, IgnoreDefault: true}, cfg)
	if err != nil {
		return err
	}
	if err := p.Parse(args); err != nil && !strings.HasPrefix(err.Error(), "unknown argument") {
		return err
	}
	return nil
}

func loadFile(path string, cfg any) error {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		return json.Unmarshal(data, cfg)
	case ".yaml", ".yml":
		return yaml.Unmarshal(data, cfg)
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}
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
