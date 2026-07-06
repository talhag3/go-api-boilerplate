package main

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/talhag3/go-api-boilerplate/internal/config"
)

func main() {

	conf, err := config.LoadConfig()

	if err != nil {
		fmt.Println("Config Error")
	}

	app := fiber.New()

	app.Get("/", func(ctx fiber.Ctx) error {

		type Data struct {
			Message string
		}

		return ctx.JSON(Data{Message: "Hi"})
	})

	app.Listen(":" + conf.AppPort)
}
