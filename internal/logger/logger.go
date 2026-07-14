package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/talhag3/go-api-boilerplate/internal/config"
)

// New creates and configures our logger.
// We use slog because it is the standard structured logger in Go.
func New(conf *config.Config) *slog.Logger {
	var level slog.Level
	// Map the string log level from config to slog.Level types.
	// We convert it to lowercase first so "DEBUG", "Debug", "debug" all match!
	switch strings.ToLower(conf.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		// Default level is info if none or invalid was provided
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	// If we are in development mode, we want easy-to-read text logs in our console.
	// In production, we want JSON logs because tools like Datadog, ELK, or AWS CloudWatch can parse them easily!
	if conf.AppEnv == "development" {
		handler = slog.NewTextHandler(os.Stdout, opts) // Prints clean lines like "time=... level=INFO msg=..."
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts) // Prints JSON like {"time": "...", "level": "INFO", "msg": "..."}
	}

	// Create and return the logger with the chosen handler
	return slog.New(handler)
}
