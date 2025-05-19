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

// Config holds the top-level configuration for the application
type Config struct {
	App    AppConfig
	Server ServerConfig
	DB     PostgresConfig
}

// AppConfig holds general configuration for the app behavior
type AppConfig struct {
	Env string `validate:"required,oneof=development production staging"`
}

// IsProduction returns true if the app is running in production mode
func (c *Config) IsProduction() bool {
	return c.App.Env == EnvProd
}

// ServerConfig holds the global dependencies and configurations needed to start and manage the HTTP server
type ServerConfig struct {
	Addr              string        `validate:"required,hostname_port"`
	ReadTimeout       time.Duration `validate:"gt=0"` // Max time the server waits to read the entire request (header + body)
	ReadHeaderTimeout time.Duration `validate:"gt=0"` // Max time the server waits to read only the request headers
	WriteTimeout      time.Duration `validate:"gt=0"` // Max time the server has to write the entire response to the client
	IdleTimeout       time.Duration `validate:"gt=0"` // Max time the server waits to keep a connection inactive (keep-alive)
}

// PostgresConfig holds configuration for PostgreSQL connections. Useful to tune performance and connection pool behavior.
type PostgresConfig struct {
	DSN             string        `validate:"required"`
	MaxOpenConns    int           `validate:"gte=1"`
	MaxIdleConns    int           `validate:"gte=0"`
	ConnMaxLifetime time.Duration `validate:"gte=0"`
}

// InitConfig builds the full application configuration by reading environment variables.
// It returns a validated Config struct or an error if any field fails validation.
func InitConfig() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Env: mustGetString("ENV", "development"),
		},
		Server: ServerConfig{
			Addr:              mustGetString("ADDR", ":3333"),
			ReadTimeout:       mustGetDuration("READ_TIMEOUT", defaultReadTimeout),
			ReadHeaderTimeout: mustGetDuration("READ_HEADER_TIMEOUT", defaultReadHeaderTimeout),
			WriteTimeout:      mustGetDuration("WRITE_TIMEOUT", defaultWriteTimeout),
			IdleTimeout:       mustGetDuration("IDLE_TIMEOUT", defaultIdleTimeout),
		},
		DB: PostgresConfig{
			DSN:             mustEnvOrPanic("DB_URL"),
			MaxOpenConns:    mustGetInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    mustGetInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: mustGetDuration("DB_CONNS_MAX_LIFETIME", time.Hour),
		},
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		var vErrs validator.ValidationErrors
		if errors.As(err, &vErrs) {
			for _, vErr := range vErrs {
				log.Printf("⚠️ Invalid config: field '%s' failed '%s' constraint (value: '%v')",
					vErr.StructNamespace(), vErr.Tag(), vErr.Value())
			}
		}
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}
