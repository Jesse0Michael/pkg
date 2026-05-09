package boot

import "github.com/jesse0michael/pkg/config"

type options struct {
	configOpts []config.Option
}

// Option configures NewApp.
type Option func(*options)

// WithConfigPrefix sets the envconfig prefix for env var resolution.
func WithConfigPrefix(prefix string) Option {
	return func(o *options) {
		o.configOpts = append(o.configOpts, config.WithPrefix(prefix))
	}
}

// WithConfigFile adds a config file to be loaded. Files are applied in order,
// so later files override earlier ones. Supports JSON and YAML.
func WithConfigFile(path string) Option {
	return func(o *options) {
		o.configOpts = append(o.configOpts, config.WithFile(path))
	}
}
