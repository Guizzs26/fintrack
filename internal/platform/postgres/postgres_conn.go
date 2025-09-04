package postgres

import (
	"context"
	"fmt"
	"log"

	"github.com/Guizzs26/fintrack/internal/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func NewPostgresConnection(ctx context.Context, cfg config.Config) (*Postgres, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	parsedCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx config: %w", err)
	}

	parsedCfg.MaxConns = cfg.Postgres.MaxConns
	parsedCfg.MinConns = cfg.Postgres.MinConns
	parsedCfg.MaxConnLifetime = cfg.Postgres.MaxConnLifetime
	parsedCfg.MaxConnIdleTime = cfg.Postgres.MaxConnIdleTime
	parsedCfg.HealthCheckPeriod = cfg.Postgres.HealthCheckPeriod
	parsedCfg.ConnConfig.ConnectTimeout = cfg.Postgres.ConnectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, parsedCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	log.Printf("✅ Postgres connection pool established successfully")
	return &Postgres{Pool: pool}, nil
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
		log.Printf("Postgres connection pool closed\n")
	}
}
