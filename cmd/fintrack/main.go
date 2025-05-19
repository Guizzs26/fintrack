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

	go func() {
		log.Printf("🚀 Server is running on %s", cfg.Server.Addr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("❌ HTTP server error: %v", err)
		}
		log.Println("Stopped serving new connections")

	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc
	log.Println("Shutting down server...")

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("❌ HTTP shutdown error: %v", err)
	}
	log.Println("Graceful shutdown complete")
}
