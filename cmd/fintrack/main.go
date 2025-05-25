package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Guizzs26/fintrack/internal/app"
	"github.com/Guizzs26/fintrack/internal/config"
	"github.com/Guizzs26/fintrack/internal/infra/db"
	"github.com/Guizzs26/fintrack/pkg/logger"
)

func init() {
	if err := config.LoadEnv(); err != nil {
		panic("❌ Failed to load env: " + err.Error())
	}
}

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		panic("❌ Failed to initialize config: " + err.Error())
	}
	logger.Init(cfg.App.Env)

	pg := db.NewPostgresConnection(cfg.DB)
	defer func() {
		if err := pg.Close(); err != nil {
			logger.L().Error("Error closing DB connection", "error", err)
		}
	}()

	logger.L().Info("Starting application", "env", cfg.App.Env)
	router := app.NewRouter(pg)
	srv := app.NewServer(cfg.Server, router)

	// Start the HTTP server in a goroutine
	go func() {
		logger.L().Info("Server is running", "addr", cfg.Server.Addr)

		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.L().Error("Unexpected server error", "error", err)
			os.Exit(1)
		}

		logger.L().Info("Stopped serving new connections")
	}()

	// channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// wait for termination signal
	sig := <-stop
	logger.L().Info("Received signal. Shutting down...", "signal", sig)

	// create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.L().Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}
	logger.L().Info("Server shutdown completed gracefully")
}
