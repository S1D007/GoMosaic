package controllers

import (
	"mosaic/modules"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func StartOverlayController(c *fiber.Ctx) error {
	gridCellFolder := c.FormValue("gridCellFolder")
	inputFolder := c.FormValue("inputFolder")
	outputFolder := c.FormValue("outputFolder")
	// processedImageFolder:= c.FormValue("processed")
	opacityStr := c.FormValue("opacity")

	if gridCellFolder == "" || inputFolder == "" || outputFolder == "" || opacityStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "All parameters are required"})
	}

	opacity, err := strconv.ParseFloat(opacityStr, 64)
	if err != nil || opacity < 0 || opacity > 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid opacity value"})
	}

	go modules.MonitorInputFolder(gridCellFolder, inputFolder, outputFolder, opacity)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Overlay process started"})
}
