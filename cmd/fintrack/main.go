package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Guizzs26/fintrack/internal/app"
	"github.com/Guizzs26/fintrack/internal/config"
	"github.com/Guizzs26/fintrack/internal/infra/db"
)

func init() {
	if err := config.LoadEnv(); err != nil {
		log.Fatalf("❌ Failed to load env: %v", err)
	}
}

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("❌ Failed to initialize config: %v", err)
	}

	router := app.NewRouter()
	srv := app.NewServer(cfg.Server, router)

	pg := db.NewPostgresConnection(cfg.DB)
	defer func() {
		if err := pg.Close(); err != nil {
			log.Printf("⚠️ Error closing DB connection: %v", err)
		}
	}()

	// Start the HTTP server in a goroutine
	go func() {
		log.Printf("🚀 Server is running on %s", cfg.Server.Addr)

		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("❌ Unexpected server error: %v", err)
		}

		log.Println("🛑 Stopped serving new connections")
	}()

	// channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// wait for termination signal
	sig := <-stop
	log.Printf("Received signal: %s. Shutting down...", sig)

	// create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}
	log.Println("✅ Server shutdown completed gracefully")
}
