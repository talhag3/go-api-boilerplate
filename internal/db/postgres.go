package db

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/talhag3/go-api-boilerplate/internal/config"
)

// Open creates a pgx connection pool.
// A connection pool keeps a bunch of database connections open so we don't have to keep reconnecting.
func Open(ctx context.Context, cfg *config.Config, log *slog.Logger) (*pgxpool.Pool, error) {
	// Parse the DSN (connection string) that we built in config.go
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse pgx config: %w", err)
	}

	// Set pool limits so we don't crash Postgres by opening too many connections at once!
	poolCfg.MaxConns = cfg.DBMaxConns
	poolCfg.MaxConnLifetime = time.Hour
	poolCfg.MaxConnIdleTime = 30 * time.Minute

	// Actually create the pool using the context and the config we parsed
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	// Give the database a strict 5-second deadline to respond to our ping.
	// If it takes longer, we stop waiting and throw an error.
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel() // Release resources when done

	// Check if the connection actually works by pinging the DB
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close() // Close the pool since we can't use it!
		return nil, fmt.Errorf("ping db: %w", err)
	}

	// Log that we connected successfully!
	log.Info("postgres connection pool created",
		"host", cfg.DBHost,
		"max_conns", cfg.DBMaxConns,
	)

	return pool, nil
}

// RunMigrations applies all pending goose migrations.
// Goose needs a standard *sql.DB, so we wrap the pgxpool with pgx's stdlib adapter.
// This is because pgx uses its own connection types, but goose expects the standard Go sql package.
func RunMigrations(pool *pgxpool.Pool, migrationsDir string, log *slog.Logger) error {
	// Tell goose to use our slog logger for its internal output so it looks nice in our console
	goose.SetLogger(gooseLogger{log: log})

	// Convert pgxpool -> *sql.DB so goose works.
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	// Tell goose we are using postgres
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	// Run all the SQL migration files that are not already run
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	log.Info("migrations applied successfully")
	return nil
}

// gooseLogger is an adapter that lets us pass our *slog.Logger to goose.
// Goose expects a specific logger interface (which has Fatalf and Printf), so we implement those methods.
type gooseLogger struct{ log *slog.Logger }

// Fatalf logs the error at the Error level and exits the process.
func (g gooseLogger) Fatalf(format string, v ...any) {
	g.log.Error(fmt.Sprintf(format, v...))
	os.Exit(1) // Crash the app because Fatal means we can't continue
}

// Printf logs informational messages from goose (like which migration is running).
func (g gooseLogger) Printf(format string, v ...any) {
	g.log.Info(fmt.Sprintf(format, v...))
}
