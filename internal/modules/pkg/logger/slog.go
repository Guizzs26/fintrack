package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Level defines the logging level for the application
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Format defines the output format for the logger
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// SlogConfig holds all the configuration for the application logger (slog)
type SlogConfig struct {
	Level     Level     // Level is the minimum level of logs to be written
	Format    Format    // Format specifies the output format (e.g., "json" or "text")
	AddSource bool      // AddSource determines whether to include the source code file and line number in the log output
	Writer    io.Writer // Writer is the destination for the logs. Defaults to os.Stdout if nil
}

// NewSlogConfig creates a new slog.Logger based on the provided configuration
func NewSlogConfig(cfg SlogConfig) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(string(cfg.Level)) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Set the output writer, defaulting to standard output
	writer := cfg.Writer
	if writer == nil {
		writer = os.Stdout
	}

	// Configure handler options based on the config
	opts := slog.HandlerOptions{
		AddSource: cfg.AddSource,
		Level:     level,
	}

	// Create the appropriate handler based on the specified format
	var handler slog.Handler
	switch cfg.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(writer, &opts)
	case FormatText:
		handler = slog.NewTextHandler(writer, &opts)
	default:
		handler = slog.NewJSONHandler(writer, &opts)
	}

	return slog.New(handler)
}
