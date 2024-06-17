package main

import (
	"fmt"
	"image"
	"mosaic/controllers"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type GridCell struct {
	img image.Image
}

var (
	ProcessedFiles = sync.Map{}
)

func main() {
	app := fiber.New()

	app.Use(logger.New())

	app.Post("/cut", controllers.CutController)
	app.Post("/start-overlay", controllers.StartOverlayController)

	err := app.Listen(":3000")
	if err != nil {
		fmt.Println("Server failed to start:", err)
	}
}
