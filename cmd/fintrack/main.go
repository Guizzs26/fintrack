package main

import (
	"log"

	"github.com/Guizzs26/fintrack/internal/app"
)

func main() {
	router := app.NewRouter()

	srv := app.NewServer(app.ServerConfig{
		Router: router,
		Addr:   ":8080",
	})

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
