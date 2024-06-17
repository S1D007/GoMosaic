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

	app.Use(cors.New(
		cors.Config{
			AllowOrigins: "*",
			AllowHeaders: "Origin, Content-Type, Accept",
		},
	))

	app.Use(logger.New())

	app.Post("/mosaic", controllers.CutController)
	app.Post("/start-overlay", controllers.StartOverlayController)

	err := app.Listen(":8000")
	if err != nil {
		fmt.Println("Server failed to start:", err)
	}
}
