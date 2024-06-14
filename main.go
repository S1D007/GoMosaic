package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/nfnt/resize"

	"github.com/fsnotify/fsnotify"
)

type gridCell struct {
	img image.Image
}

var (
	processedFiles = sync.Map{}
)

func main() {
	app := fiber.New()

	app.Use(logger.New())

	app.Post("/mosaic", uploadHandler)
	app.Post("/start-overlay", startOverlayHandler)

	err := app.Listen(":3000")
	if err != nil {
		fmt.Println("Server failed to start:", err)
	}
}

func uploadHandler(c *fiber.Ctx) error {
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

func startOverlayHandler(c *fiber.Ctx) error {
	gridCellFolder := c.FormValue("gridCellFolder")
	inputFolder := c.FormValue("inputFolder")
	outputFolder := c.FormValue("outputFolder")
	opacityStr := c.FormValue("opacity")

	if gridCellFolder == "" || inputFolder == "" || outputFolder == "" || opacityStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "All parameters are required"})
	}

	opacity, err := strconv.ParseFloat(opacityStr, 64)
	if err != nil || opacity < 0 || opacity > 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid opacity value"})
	}

	go monitorInputFolder(gridCellFolder, inputFolder, outputFolder, opacity)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Overlay process started"})
}

func monitorInputFolder(gridCellFolder, inputFolder, outputFolder string, opacity float64) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Error creating watcher:", err)
		return
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					handleNewFile(event.Name, gridCellFolder, outputFolder, opacity)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("Error:", err)
			}
		}
	}()

	err = watcher.Add(inputFolder)
	if err != nil {
		fmt.Println("Error adding watcher:", err)
		return
	}
	<-done
}

func overlayImages(inputFile, gridCellFolder, outputFolder string, opacity float64) {
	inputImg, err := loadImage(inputFile)
	if err != nil {
		fmt.Println("Error loading input image:", err)
		return
	}

	gridCells, err := loadGridCells(gridCellFolder)
	if err != nil {
		fmt.Println("Error loading grid cells:", err)
		return
	}

	if len(gridCells) == 0 {
		fmt.Println("No grid cells found")
		return
	}

	// Shuffle the grid cells
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(gridCells), func(i, j int) {
		gridCells[i], gridCells[j] = gridCells[j], gridCells[i]
	})

	// Resize the input image to match the dimensions of the first grid cell
	gridCellSize := gridCells[0].img.Bounds().Size()
	resizedInputImg := resizeAndCrop(inputImg, gridCellSize)

	bounds := resizedInputImg.Bounds()
	outputImg := image.NewRGBA(bounds)

	draw.Draw(outputImg, bounds, resizedInputImg, image.Point{}, draw.Src)

	for i, gridCell := range gridCells {
		if i >= bounds.Dx()/gridCellSize.X*bounds.Dy()/gridCellSize.Y {
			break
		}

		alphaImg := gridCell.alpha(opacity)
		draw.DrawMask(outputImg, gridCell.img.Bounds(), gridCell.img, image.Point{}, alphaImg, image.Point{}, draw.Over)
	}

	outputFilePath := filepath.Join(outputFolder, filepath.Base(inputFile))
	err = saveImage(outputImg, outputFilePath)
	if err != nil {
		fmt.Println("Error saving output image:", err)
	}
}

func resizeAndCrop(img image.Image, size image.Point) image.Image {
	resizedImg := resize.Resize(uint(size.X), uint(size.Y), img, resize.Lanczos3)

	// Crop from center
	dx, dy := resizedImg.Bounds().Dx(), resizedImg.Bounds().Dy()
	x0 := (dx - size.X) / 2
	y0 := (dy - size.Y) / 2
	x1 := x0 + size.X
	y1 := y0 + size.Y

	croppedImg := resizedImg.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(x0, y0, x1, y1))

	return croppedImg
}

func loadImage(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	return img, err
}

func loadGridCells(gridCellFolder string) ([]gridCell, error) {
	var gridCells []gridCell
	err := filepath.Walk(gridCellFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") || strings.HasSuffix(path, ".png")) {
			img, err := loadImage(path)
			if err != nil {
				return err
			}
			gridCells = append(gridCells, gridCell{img: img})
		}
		return nil
	})
	return gridCells, err
}

func (gc gridCell) alpha(opacity float64) *image.Alpha {
	bounds := gc.img.Bounds()
	alphaImg := image.NewAlpha(bounds)

	alphaValue := uint8(opacity * 255)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			alphaImg.Set(x, y, color.Alpha{A: alphaValue})
		}
	}

	return alphaImg
}

func saveImage(img image.Image, filePath string) error {
	outFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	switch filepath.Ext(filePath) {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(outFile, img, nil)
	case ".png":
		err = png.Encode(outFile, img)
	default:
		err = fmt.Errorf("unsupported file extension")
	}
	return err
}

func handleNewFile(filePath, gridCellFolder, outputFolder string, opacity float64) {
	if _, exists := processedFiles.Load(filePath); exists {
		return
	}
	processedFiles.Store(filePath, struct{}{})

	if strings.HasSuffix(filePath, ".jpg") || strings.HasSuffix(filePath, ".jpeg") || strings.HasSuffix(filePath, ".png") {
		fmt.Println("New file detected:", filePath)
		overlayImages(filePath, gridCellFolder, outputFolder, opacity)
	}
}
