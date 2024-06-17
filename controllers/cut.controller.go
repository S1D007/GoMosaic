package controllers

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
)

func CutController(c *fiber.Ctx) error {
	numRows := c.FormValue("rows")
	numCols := c.FormValue("cols")
	outputFolderNameValue := c.FormValue("output")

	rows, err := strconv.Atoi(numRows)
	if err != nil || rows <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid number of rows"})
	}

	cols, err := strconv.Atoi(numCols)
	if err != nil || cols <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid number of columns"})
	}

	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to read image file"})
	}

	inputFile, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open image file"})
	}
	defer inputFile.Close()

	img, format, err := image.Decode(inputFile)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid image file"})
	}

	var outputDir string
	if filepath.IsAbs(outputFolderNameValue) {
		outputDir = outputFolderNameValue
	} else {
		outputDir = filepath.Join("output", outputFolderNameValue)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create output directory"})
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	cellWidth := width / cols
	cellHeight := height / rows

	var wg sync.WaitGroup
	errors := make(chan error, rows*cols)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			wg.Add(1)
			go func(row, col int) {
				defer wg.Done()

				x0 := col * cellWidth
				y0 := row * cellHeight
				x1 := x0 + cellWidth
				y1 := y0 + cellHeight
				if x1 > width {
					x1 = width
				}
				if y1 > height {
					y1 = height
				}

				subImg := img.(interface {
					SubImage(r image.Rectangle) image.Image
				}).SubImage(image.Rect(x0, y0, x1, y1))

				outputFile := filepath.Join(outputDir, fmt.Sprintf("R%dC%d.%s", row+1, col+1, format))
				outFile, err := os.Create(outputFile)
				if err != nil {
					errors <- fmt.Errorf("failed to create output file %s: %v", strings.ToLower(outputFile), err)
					return
				}
				defer outFile.Close()

				var encodeErr error
				switch format {
				case "jpeg":
					encodeErr = jpeg.Encode(outFile, subImg, nil)
				case "png":
					encodeErr = png.Encode(outFile, subImg)
				default:
					encodeErr = fmt.Errorf("unsupported image format: %s", strings.ToLower(format))
				}

				if encodeErr != nil {
					errors <- fmt.Errorf("failed to encode image %s: %v", strings.ToLower(outputFile), encodeErr)
				}
			}(row, col)
		}
	}

	go func() {
		wg.Wait()
		close(errors)
	}()

	for err := range errors {
		fmt.Println("Error:", err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":          "Image split into grid cells successfully",
		"output_directory": outputDir,
	})
}
