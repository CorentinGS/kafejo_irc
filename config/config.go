package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/pkg/errors"
)

// Config holds the configuration for the application.
type (
	// Config -.
	Config struct {
		App  `yaml:"app"`
		HTTP `yaml:"http"`
		Log  `yaml:"logger"`
	}

	// App -.
	App struct {
		Name    string `env-required:"true" yaml:"name"    env:"APP_NAME"`
		Version string `env-required:"true" yaml:"version" env:"APP_VERSION"`
	}

	// HTTP -.
	HTTP struct {
		Port string `env-required:"true" env:"HTTP_PORT" yaml:"port" env-default:"8080"`
		Host string `env-required:"true" env:"HTTP_HOST" yaml:"host" env-default:"localhost"`
	}

	// Log -.
	Log struct {
		Level string `env-required:"true" env:"LOG_LEVEL" yaml:"level" env-default:"warning"`
	}
)

var (
	AccessToken string //nolint:gochecknoglobals //Very Secure
	WebhookURL  string //nolint:gochecknoglobals //Very Secure
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig("./config/config.yml", cfg)
	if err != nil {
		return nil, errors.Wrap(err, "error reading config")
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
