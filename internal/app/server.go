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
		ReadTimeout:       cfg.ReadTimeout,       // Max time the server waits to read the entire request (header + body)
		ReadHeaderTimeout: cfg.ReadHeaderTimeout, // Max time the server waits to read only the request headers
		WriteTimeout:      cfg.WriteTimeout,      // Max time the server has to write the entire response to the client
		IdleTimeout:       cfg.IdleTimeout,       // Max time the server waits to keep a connection inactive (keep-alive)
	}
}
