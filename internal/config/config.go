package config

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
)

const (
	EnvDev  = "development"
	EnvProd = "production"

	defaultReadTimeout       = 10 * time.Second
	defaultReadHeaderTimeout = 5 * time.Second
	defaultWriteTimeout      = 10 * time.Second
	defaultIdleTimeout       = 60 * time.Second
)

type Config struct {
	App          AppConfig
	ServerConfig ServerConfig
}

type AppConfig struct {
	Env string `validate:"required,oneof=development production staging"`
}

func (c *Config) IsProduction() bool {
	return c.App.Env == EnvProd
}

/*
- ServerConfig holds the global dependencies and configurations needed to start the server.
*/
type ServerConfig struct {
	Addr              string        `validate:"required,hostname_port"`
	ReadTimeout       time.Duration `validate:"gt=0"`
	ReadHeaderTimeout time.Duration `validate:"gt=0"`
	WriteTimeout      time.Duration `validate:"gt=0"`
	IdleTimeout       time.Duration `validate:"gt=0"`
}

func InitConfig() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Env: mustGetString("ENV", "development"),
		},
		ServerConfig: ServerConfig{
			Addr:              mustGetString("ADDR", ":3333"),
			ReadTimeout:       mustGetDuration("READ_TIMEOUT", defaultReadTimeout),
			ReadHeaderTimeout: mustGetDuration("READ_HEADER_TIMEOUT", defaultReadHeaderTimeout),
			WriteTimeout:      mustGetDuration("WRITE_TIMEOUT", defaultWriteTimeout),
			IdleTimeout:       mustGetDuration("IDLE_TIMEOUT", defaultIdleTimeout),
		},
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		var vErrs validator.ValidationErrors
		if errors.As(err, &vErrs) {
			for _, vErr := range vErrs {
				log.Printf("⚠️ Config validation: Field '%s' failed '%s' tag", vErr.Field(), vErr.Tag())
			}
		}
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}
