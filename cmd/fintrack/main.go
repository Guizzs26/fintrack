package main

import (
	"log"

	"github.com/Guizzs26/fintrack/internal/app"
	"github.com/Guizzs26/fintrack/internal/config"
)

func main() {
	cfg := config.InitConfig()

	router := app.NewRouter()

	srv := app.NewServer(cfg.ServerConfig, router)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
