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
	// 1. Load configuration first. If this fails, we can't do anything, so we panic.
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	// 2. Initialize our slog logger.
	log := logger.New(cfg)

	// Override Go's default logger so third-party packages output structured logs too
	slog.SetDefault(log)

	// 3. Open the database connection pool.
	// We use context.Background() because this is the root of the app,
	// we don't want the DB connection to be cancelled by a timeout yet.
	ctx := context.Background()
	pool, err := db.Open(ctx, cfg, log)
	if err != nil {
		log.Error("failed to open db", "error", err)
		os.Exit(1) // Exit with code 1 means "error"
	}

	// 4. Run database migrations before starting the server.
	// If migrations fail, the app won't match the code, so we exit.
	if err := db.RunMigrations(pool, "internal/db/migrations", log); err != nil {
		log.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// 5. Create the HTTP server (wires up all our handlers, services, repos).
	srv := server.New(cfg, log, pool)

	// 6. Set up graceful shutdown.
	// This tells the OS: "Hey, if someone presses Ctrl+C or sends a kill signal,
	// send it to this context instead of just murdering the process instantly."
	ctxStop, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop() // Clean up the signal listener when main() ends

	// 7. Start the server in a Go routine (a background thread).
	// We do this because srv.Start() is a blocking call. If we ran it normally,
	// the code below it (the shutdown logic) would never execute.
	errCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errCh <- err
		}
	}()

	// 8. The `select` block pauses here and waits for one of two things to happen.
	select {
	// Case A: The server crashed on its own (e.g., port already in use)
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	// Case B: The OS sent a shutdown signal (Ctrl+C)
	case <-ctxStop.Done():
		log.Info("shutdown signal received")
	}

	// 9. If we reach here, it means we are shutting down.
	// Give the server 15 seconds to finish processing any active requests
	// before forcefully killing the connections.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
	}

	log.Info("application stopped cleanly")
}
