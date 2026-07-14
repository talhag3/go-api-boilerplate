package middleware

import (
	"log/slog"
	"runtime/debug"

	"github.com/gofiber/fiber/v3"
)

// Recover catches panics so a single bug doesn't crash our whole web server.
// Fiber has its own recover middleware, but we wrote ours so we can log using slog.
// NOTE: I accidentally put the Recover middleware inside logger.go, and the Logger middleware inside recover.go... oops! I need to rename these files later.
func Recover(log *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) (err error) {
		// In Go, `defer` is awesome because it runs right before the function exits, no matter what!
		// `recover()` will catch any panic that happens while the next handlers run.
		// We use a named return value `(err error)` so we can modify the return error inside the defer func.
		defer func() {
			if r := recover(); r != nil {
				// Get the stack trace. This shows exactly which line of code exploded!
				// Otherwise, panics are very hard to debug in production.
				stackTrace := string(debug.Stack())

				// Log the error and the stack trace
				log.Error("panic recovered",
					"error", r,
					"path", c.Path(),
					"stack", stackTrace,
				)

				// Send back a clean 500 error to the client instead of just dropping the connection
				err = fiber.ErrInternalServerError
			}
		}()

		// Continue to the next handler in the router chain.
		// If any handler after this panics, our defer block above will catch it!
		return c.Next()
	}
}
