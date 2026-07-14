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

// Open creates a pgx connection pool

func Open(ctx context.Context, cfg *config.Config, log *slog.Logger) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse pgx config: %w", err)
	}

	// Set pool limits to prevent exhausting database connections
	poolCfg.MaxConns = cfg.DBMaxConns
	poolCfg.MaxConnLifetime = time.Hour
	poolCfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	// Give the database a strict 5-second deadline to respond to the ping.
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	log.Info("postgres connection pool created",
		"host", cfg.DBHost,
		"max_conns", cfg.DBMaxConns,
	)

	return pool, nil
}

// RunMigrations applies all pending goose migrations.
// Goose needs a standard *sql.DB, so we wrap the pgxpool with pgx's stdlib adapter.
func RunMigrations(pool *pgxpool.Pool, migrationsDir string, log *slog.Logger) error {
	// Tell goose to use our slog logger for its internal output
	goose.SetLogger(gooseLogger{log: log})

	// Convert pgxpool -> *sql.DB so goose (which uses database/sql) works.
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	log.Info("migrations applied successfully")
	return nil
}

// gooseLogger is an adapter that lets us pass our *slog.Logger to goose.
// Goose expects a specific logger interface, so we implement it here.
type gooseLogger struct{ log *slog.Logger }

// Fatalf logs the error at the Error level and exits the process.
// slog doesn't have a built-in Fatal that exits, so we do it manually.
func (g gooseLogger) Fatalf(format string, v ...any) {
	g.log.Error(fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Printf logs informational messages from goose (like which migration is running).
func (g gooseLogger) Printf(format string, v ...any) {
	g.log.Info(fmt.Sprintf(format, v...))
}
