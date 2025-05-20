package db

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/Guizzs26/fintrack/internal/config"
	_ "github.com/lib/pq"
)

type Postgres struct {
	DB *sql.DB
}

func NewPostgresConnection(cfg config.PostgresConfig) *Postgres {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		log.Fatalf("❌ Failed to open PostgreSQL connection: %v", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("❌ Failed to ping PostgreSQL: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL successfully")
	return &Postgres{DB: db}
}

func (pg *Postgres) Close() error {
	log.Println("🧯 Closing PostgreSQL connection")
	return pg.DB.Close()
}
