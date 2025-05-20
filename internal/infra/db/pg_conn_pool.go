package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/Guizzs26/fintrack/internal/config"
	"github.com/Guizzs26/fintrack/pkg/logger"
	_ "github.com/lib/pq"
)

type Postgres struct {
	DB *sql.DB
}

func NewPostgresConnection(cfg config.PostgresConfig) *Postgres {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		logger.L().Error("Failed to open PostgreSQL connection", "error", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logger.L().Error("Failed to ping PostgreSQL", "error", err)
		panic(err)
	}

	logger.L().Info("Connected to PostgreSQL successfully")
	return &Postgres{DB: db}
}

func (pg *Postgres) Close() error {
	logger.L().Info("Closing PostgreSQL connection")
	return pg.DB.Close()
}
