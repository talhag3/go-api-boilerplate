package main

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/talhag3/go-api-boilerplate/internal/config"
	"github.com/talhag3/go-api-boilerplate/internal/logger"
)

func main() {

	conf, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	log := logger.New(conf)

	// Override Go's default logger so third-party packages output structured logs instead of plain text.
	slog.SetDefault(log)

	app := fiber.New()

	app.Get("/", func(ctx fiber.Ctx) error {

		type Data struct {
			Message string
		}

		return ctx.JSON(Data{Message: "Hi"})
	})

	log.Debug("Application Started! ", "port", conf.AppPort)
	app.Listen(":" + conf.AppPort)
}
