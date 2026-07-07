package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/talhag3/go-api-boilerplate/internal/config"
	"github.com/talhag3/go-api-boilerplate/internal/logger"
)

func main() {

	conf, err := config.LoadConfig()

	log := logger.New(conf.AppEnv)

	if err != nil {
		log.Error("Config Error", "error", err.Error())
	}

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
