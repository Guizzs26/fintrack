package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Guizzs26/fintrack/internal/platform/config"
	"github.com/Guizzs26/fintrack/internal/platform/postgres"
	"github.com/labstack/echo/v4"
)

const MockUserID = ""

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config to api: %s\n", err)
		os.Exit(1)
	}

	if err := run(ctx, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg *config.Config) error {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	e := echo.New()

	pg, err := postgres.NewPostgresConnection(ctx, *cfg)
	if err != nil {
		return err
	}
	defer pg.Close()

	e.Logger.Fatal(e.Start(":9999"))

	return nil
}
