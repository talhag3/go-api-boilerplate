package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/talhag3/go-api-boilerplate/internal/config"
)

func New(conf *config.Config) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(conf.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	// In dev, use TextHandler for readable console logs.
	// In prod, use JSONHandler for machine-parseable logs.
	if conf.AppEnv == "development" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
