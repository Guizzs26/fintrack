package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Guizzs26/fintrack/internal/modules/ledger"
	"github.com/Guizzs26/fintrack/internal/modules/pkg/clock"
	"github.com/Guizzs26/fintrack/internal/modules/pkg/validatorx"
	"github.com/Guizzs26/fintrack/internal/platform/config"
	"github.com/Guizzs26/fintrack/internal/platform/postgres"
	"github.com/labstack/echo/v4"
)

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
	e.Validator = validatorx.NewValidator()

	pgConn, err := postgres.NewPostgresConnection(ctx, *cfg)
	if err != nil {
		return err
	}
	defer pgConn.Close()

	clock := clock.SystemClock{}

	// ----- Ledger module dependencies ----- //
	accountRepo := ledger.NewPostgresAccountRepository(pgConn.Pool)
	ledgerSvc := ledger.NewLedgerService(accountRepo, clock)
	ledgerHandler := ledger.NewLedgerHandler(ledgerSvc)

	apiRouteGroup := e.Group("/api/v1")
	ledgerHandler.RegisterRoutes(apiRouteGroup)

	e.Logger.Fatal(e.Start(":9999"))
	return nil
}
