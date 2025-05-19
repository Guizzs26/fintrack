package app

import (
	"net/http"

	"github.com/Guizzs26/fintrack/internal/config"
)

// NewServer builds and returns a configured *http.Server
func NewServer(cfg config.ServerConfig, router http.Handler) *http.Server {
	return &http.Server{
		Handler:           router,
		Addr:              cfg.Addr,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
}
