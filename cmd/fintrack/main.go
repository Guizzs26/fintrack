package main

import (
	"log"

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

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
