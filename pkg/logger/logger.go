package logger

import (
	"log/slog"
	"os"
)

var log *slog.Logger

func Init(env string) {
	var handler slog.Handler

	switch env {
	case "production":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})

	default:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	log = slog.New(handler)
	slog.SetDefault(log)
}

// L returns the global logger
func L() *slog.Logger {
	if log == nil {
		panic("logger not initialized")
	}
	return log
}
