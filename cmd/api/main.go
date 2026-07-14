package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/talhag3/go-api-boilerplate/internal/config"
	"github.com/talhag3/go-api-boilerplate/internal/db"
	"github.com/talhag3/go-api-boilerplate/internal/logger"
	"github.com/talhag3/go-api-boilerplate/internal/server"
)

func main() {
	// Let's load the configuration. If this fails, we can't do anything, so we have to panic and stop!
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	// Initialize our structured logger. slog is built-in now in newer Go versions, pretty cool!
	log := logger.New(cfg)

	// Set this as the default logger for the whole application
	slog.SetDefault(log)

	// Let's create a background context. We need this to open the DB connection.
	ctx := context.Background()
	// Open the database connection pool. We pass the config and the logger we just made.
	pool, err := db.Open(ctx, cfg, log)
	if err != nil {
		// Log the error and exit the app with code 1 (which means something went wrong)
		log.Error("failed to open db", "error", err)
		os.Exit(1)
	}

	// Run database migrations before starting the server so the DB is up to date!
	if err := db.RunMigrations(pool, "internal/db/migrations", log); err != nil {
		log.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Create the HTTP server and pass all the dependencies.
	srv := server.New(cfg, log, pool)

	// Set up graceful shutdown. We wait for Ctrl+C (SIGINT) or SIGTERM from the OS.
	ctxStop, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop() // Always call stop to clean up resources!

	// Start the server in a goroutine because Start() blocks.
	// We need a channel to receive the error if it fails to start.
	errCh := make(chan error, 1)
	go func() {
		// This runs in the background
		if err := srv.Start(); err != nil {
			errCh <- err // Send error to the channel
		}
	}()

	// The select statement blocks until one of the cases receives a message!
	select {
	case err := <-errCh:
		// If the server failed to start, log it and exit
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	case <-ctxStop.Done():
		// We got a termination signal (Ctrl+C)
		log.Info("shutdown signal received")
	}

	// Wait up to 15 seconds for active requests to finish before shutting down
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel() // cancel must be called to release resources!

	// Shutdown the server gracefully
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
	}

	log.Info("application stopped cleanly")
}
