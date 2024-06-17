package main

import (
	"fmt"
	"mosaic/controllers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New(fiber.Config{
		BodyLimit: 1024 * 1024 * 1024,
	})

	app.Use(cors.New())

	app.Use(logger.New())

	app.Post("/mosaic", controllers.CutController)
	app.Post("/start-overlay", controllers.StartOverlayController)

	err := app.Listen(":3000")
	if err != nil {
		fmt.Println("Server failed to start:", err)
	}
}
