package config

type AppConfig struct {
	Environment string `envconfig:"ENVIRONMENT"`
	Name        string `envconfig:"APP_NAME"`
	Version     string `envconfig:"VERSION"`
	LogLevel    string `envconfig:"LOG_LEVEL"`
}
