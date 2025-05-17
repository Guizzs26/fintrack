package app

import (
	"net/http"
	"time"
)

/*
- ServerConfig holds the global dependencies and configurations needed to start the server.

- This struct in injected with things like DB connections, loggers, config, etc...
*/
type ServerConfig struct {
	Router            http.Handler
	Addr              string
	ReadTimeout       int
	ReadHeaderTimeout int
	WriteTimeout      int
	IdleTimeout       int
}

// NewServer builds and returns a configured *http.Server
func NewServer(cfg ServerConfig) *http.Server {
	return &http.Server{
		Handler:           cfg.Router,
		Addr:              cfg.Addr,
		ReadTimeout:       time.Second * 10, // Max time the server waits to read the entire request (header + body)
		ReadHeaderTimeout: time.Second * 5,  // Max time the server waits to read only the request headers
		WriteTimeout:      time.Second * 10, // Max time the server has to write the entire response to the client
		IdleTimeout:       time.Second * 60, // Max time the server waits to keep a connection inactive (keep-alive)
	}
}
