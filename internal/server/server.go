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

// Server holds all the components that our application needs to run.
// Putting them in a struct makes it really clean to start/stop the server.
type Server struct {
	app  *fiber.App     // The Fiber app instance (web framework)
	cfg  *config.Config // App configuration
	log  *slog.Logger   // Structured logger
	pool *pgxpool.Pool  // Postgres connection pool
}

// New sets up our Fiber web application, wires all our layers together, and returns a Server.
// This is where the magic happens - we connect database -> repository -> service -> handlers!
func New(cfg *config.Config, log *slog.Logger, pool *pgxpool.Pool) *Server {
	// 1. Create the Fiber app. We set some custom timeouts so requests don't hang forever!
	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Second, // Max time to read the incoming request body
		WriteTimeout: 10 * time.Second, // Max time to write the response back to the client
		IdleTimeout:  60 * time.Second, // Keep-alive connection timeout
		// Custom error handler so that Fiber returns structured JSON error pages
		// instead of default plain HTML error pages when something goes wrong!
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			// If it's a Fiber-specific error (like 404 Route Not Found), get its code
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

	// 2. Register global middlewares that run on every single request.
	app.Use(requestid.New())         // Adds a unique Request ID header to help track requests
	app.Use(middleware.Recover(log)) // Catches code panics so the server doesn't crash
	app.Use(middleware.Logger(log))  // Logs information about requests (method, status, latency)

	// 3. Simple health check route. Very useful for Docker/Kubernetes health checks!
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// 4. Wire up all our project layers.
	// Repo depends on DB pool -> Service depends on Repo -> Handler depends on Service.
	userRepo := repository.NewUserRepository(pool)
	userSvc := service.NewUserService(userRepo, log)
	userHandler := handler.NewUserHandler(userSvc)

	// 5. Register the user routes under the "/api/v1" prefix.
	api := app.Group("/api/v1")
	userHandler.Register(api)

	return &Server{
		app:  app,
		cfg:  cfg,
		log:  log,
		pool: pool,
	}
}

// Start makes the Fiber app listen for HTTP requests on the port from our config.
func (s *Server) Start() error {
	addr := ":" + s.cfg.AppPort
	s.log.Info("http server starting", "addr", addr)
	// app.Listen is blocking, meaning the code pauses here while the server is running.
	return s.app.Listen(addr)
}

// Shutdown stops the server gracefully.
// It stops accepting new requests, waits for active requests to finish, and closes the database connection.
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("http server shutting down")

	// Shutdown the Fiber web server
	if err := s.app.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("fiber shutdown failed: %w", err)
	}

	// Close the DB connection pool safely
	s.pool.Close()

	return nil
}
