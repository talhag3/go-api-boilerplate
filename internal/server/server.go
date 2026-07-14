package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/talhag3/go-api-boilerplate/internal/config"
	"github.com/talhag3/go-api-boilerplate/internal/handler"
	"github.com/talhag3/go-api-boilerplate/internal/middleware"
	"github.com/talhag3/go-api-boilerplate/internal/repository"
	"github.com/talhag3/go-api-boilerplate/internal/service"
)

// Server holds everything our app needs to run.
// Keeping it in a struct like this makes it easy to start and stop the app cleanly.
type Server struct {
	app  *fiber.App
	cfg  *config.Config
	log  *slog.Logger
	pool *pgxpool.Pool
}

// New sets up the Fiber app, wires up all the dependencies, and returns a Server.
// This is the "composition root" - the one place where we connect the database,
// repository, service, and handlers together.
func New(cfg *config.Config, log *slog.Logger, pool *pgxpool.Pool) *Server {
	// 1. Create the Fiber app with custom timeouts and an error handler
	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Second, // Max time to read the request body
		WriteTimeout: 10 * time.Second, // Max time to write the response
		IdleTimeout:  60 * time.Second, // Max time for an idle keep-alive connection
		// Custom error handler so we always return JSON, even when Fiber throws an internal error
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			// If it's a Fiber error (like fiber.ErrNotFound), grab its status code
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "FIBER_ERROR",
					"message": err.Error(),
				},
			})
		},
	})

	// 2. Register global middlewares (these run on EVERY request)
	app.Use(requestid.New())         // Adds an X-Request-ID to every request for tracking
	app.Use(middleware.Recover(log)) // Catches panics so the server doesn't crash
	app.Use(middleware.Logger(log))  // Logs the request method, path, and latency

	// 3. Simple health check route for Kubernetes/Docker health probes
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// 4. Wire up dependencies layer by layer
	// The arrows show the flow: Repo depends on DB -> Service depends on Repo -> Handler depends on Service
	userRepo := repository.NewUserRepository(pool)
	userSvc := service.NewUserService(userRepo, log)
	userHandler := handler.NewUserHandler(userSvc)

	// 5. Set up the API routes
	api := app.Group("/api/v1") // All routes will start with /api/v1
	userHandler.Register(api)

	return &Server{
		app:  app,
		cfg:  cfg,
		log:  log,
		pool: pool,
	}
}

// Start tells Fiber to begin listening for HTTP requests on the configured port.
func (s *Server) Start() error {
	addr := ":" + s.cfg.AppPort
	s.log.Info("http server starting", "addr", addr)
	// app.Listen is a blocking call - it will run until the server shuts down
	return s.app.Listen(addr)
}

// Shutdown gracefully stops the server.
// It stops accepting new requests, waits for active requests to finish, then closes the DB.
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("http server shutting down")

	// Tell Fiber to stop accepting new connections and wait for active ones to finish
	if err := s.app.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("fiber shutdown failed: %w", err)
	}

	// Once the HTTP server is down, safely close the database connection pool
	s.pool.Close()

	return nil
}
