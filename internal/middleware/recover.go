package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Logger returns a Fiber middleware that logs every request.
// We use slog so our access logs match the rest of our app's JSON format.
func Logger(log *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Record the start time so we can calculate how long the request took
		start := time.Now()

		// Call the next handler in the chain (our actual route logic)
		err := c.Next()

		// Build a list of key-value pairs for our structured log
		// slog accepts a slice of 'any' for this, which is really flexible
		fields := []any{
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"latency_ms", time.Since(start).Milliseconds(), // easier to read in JSON than seconds
			"ip", c.IP(),
		}

		// If the route handler returned an error, log it as an Error level
		if err != nil {
			fields = append(fields, "error", err.Error())
			log.Error("request failed", fields...)
		} else {
			// Otherwise, just a normal Info level access log
			log.Info("request completed", fields...)
		}

		// We must return the error so Fiber's built-in error handler can still process it
		return err
	}
}
