package config

import (
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Server struct {
		Port         string        `envconfig:"SERVER_PORT" default:"3333"`
		ReadTimeout  time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"5s"`
		WriteTimeout time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"10s"`
		IdleTimeout  time.Duration `envconfig:"SERVER_IDLE_TIMEOUT" default:"120s"`
	}
	Postgres struct {
		MaxConns          int32         `envconfig:"PGX_MAX_CONNS" default:"20"`
		MinConns          int32         `envconfig:"PGX_MIN_CONNS" default:"5"`
		MaxConnLifetime   time.Duration `envconfig:"PGX_MAX_CONN_LIFETIME" default:"30m"`
		MaxConnIdleTime   time.Duration `envconfig:"PGX_MAX_CONN_IDLE_TIME" default:"5m"`
		HealthCheckPeriod time.Duration `envconfig:"PGX_HEALTH_CHECK_PERIOD" default:"1m"`
		ConnectTimeout    time.Duration `envconfig:"PGX_CONNECT_TIMEOUT" default:"5s"`
	}
	Database struct {
		Host     string `envconfig:"DB_HOST" required:"true"`
		Port     int    `envconfig:"DB_PORT" required:"true"`
		User     string `envconfig:"DB_USER" required:"true"`
		Password string `envconfig:"DB_PASSWORD" required:"true"`
		Name     string `envconfig:"DB_NAME" required:"true"`
		SSLMode  string `envconfig:"DB_SSL_MODE" default:"disable"`
	}
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf("error loading .env file: %s", err)
		return nil, err
	}

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process config from environment: %w", err)
	}
	log.Println("✔️ Configuration loaded successfully")
	return &cfg, nil
}
