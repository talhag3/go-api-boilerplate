package middleware

import (
	"log/slog"
	"runtime/debug"

	"github.com/gofiber/fiber/v3"
)

// Recover catches panics so a single bug doesn't crash the entire server.
// Fiber has a built-in Recover, but writing our own lets us log using slog.
func Recover(log *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) (err error) {
		// In Go, `defer` runs after the function finishes.
		// `recover()` stops a panic from crashing the program.
		// We use a named return value `(err error)` so we can change what
		// gets returned to Fiber if a panic happens.
		defer func() {
			if r := recover(); r != nil {
				// Grab the stack trace so we know exactly which line of code exploded.
				// This is invaluable for debugging production issues.
				stackTrace := string(debug.Stack())

				log.Error("panic recovered",
					"error", r,
					"path", c.Path(),
					"stack", stackTrace,
				)

				// Return a standard 500 error to the client instead of crashing the process
				err = fiber.ErrInternalServerError
			}
		}()

		// Call the next handler. If it panics, the deferred function above catches it.
		return c.Next()
	}
}
