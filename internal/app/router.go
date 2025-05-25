package app

import (
	"net/http"

	"github.com/Guizzs26/fintrack/internal/bootstrap"
	"github.com/Guizzs26/fintrack/internal/infra/db"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter registers global middlewares and mounts all module routes
func NewRouter(pg *db.Postgres) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	bootstrap.RegisterModules(r, pg)

	return r
}
