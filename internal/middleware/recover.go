package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Logger returns a Fiber middleware that logs details of every HTTP request.
// We use slog here too so these logs are structured just like the other logs.
// NOTE: Yes, this Logger middleware is in the recover.go file. I got the filenames mixed up. I will fix it later!
func Logger(log *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Save the start time so we can see how long the request took to process!
		start := time.Now()

		// Call the next handler in the chain
		err := c.Next()

		// Build a slice of key-value pairs for structured logging.
		// "any" is the new name for empty interface (interface{}), which is cool.
		fields := []any{
			"method", c.Method(),                         // GET, POST, etc.
			"path", c.Path(),                             // like /api/v1/users
			"status", c.Response().StatusCode(),         // 200, 404, 500, etc.
			"latency_ms", time.Since(start).Milliseconds(), // Calculate how long it took in milliseconds
			"ip", c.IP(),                                 // Client's IP address
		}

		// If there was an error in the handler, log it as an Error level
		if err != nil {
			fields = append(fields, "error", err.Error())
			log.Error("request failed", fields...)
		} else {
			// Otherwise, log it as a normal Info request
			log.Info("request completed", fields...)
		}

		// Important: we must return the error so Fiber's error handler can run!
		return err
	}
}
