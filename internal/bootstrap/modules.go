package bootstrap

import (
	"github.com/Guizzs26/fintrack/internal/infra/db"
	"github.com/Guizzs26/fintrack/internal/modules/identity/auth/delivery/rest"
	"github.com/go-chi/chi/v5"
)

func RegisterModules(r chi.Router, pg *db.Postgres) {
	authHanlder := rest.NewAuthHandler()

	authHanlder.RegisterRoutes(r)
}
