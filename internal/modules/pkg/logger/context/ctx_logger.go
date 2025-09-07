package ctxlogger

import (
	"context"
	"log/slog"
)

// key is an unexported type used as the key for the logger in the context
// Using an unexported type prevents key collisions with other packages
type key string

// loggerKey is the specific key value used to store the logger in the context
const loggerKey key = "logger"

// SetLogger returns a new context that carries the provided slog.Logger
func SetLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// GetLogger retrieves the slog.Logger from the provided context
// If no logger is found in the context, it safely returns the global
// default logger, ensuring that a valid logger is always returned
func GetLogger(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}

	// Fallback return the default logger if none was found
	// This prevents nil pointer panics in case a context without a logger is passed
	return slog.Default()
}
