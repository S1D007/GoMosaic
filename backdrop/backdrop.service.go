package backdrop

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"
	"math"
	"mosaic/aws"

	"github.com/fogleman/gg"
	"github.com/gofiber/fiber/v2"
)

type Backdrop struct {
	ID      string `json:"eventId"`
	Rows    int    `json:"rows"`
	Columns int    `json:"cols"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
}

func calculateFontSize(boxSize int, text string) float64 {
	const tolerance = 0.8
	fontSize := float64(boxSize) / 1.5
	dc := gg.NewContext(1, 1)
	defer dc.Clear()

	for {
		dc.LoadFontFace("./fonts/roboto.ttf", fontSize)
		textWidth, _ := dc.MeasureString(text)

		if textWidth < float64(boxSize) || fontSize <= tolerance {
			break
		}
		fontSize *= 0.3
	}

	lowerBound := fontSize / 2
	upperBound := fontSize

	for lowerBound < upperBound-tolerance {
		mid := (lowerBound + upperBound) / 2
		dc.LoadFontFace("./fonts/roboto.ttf", mid)
		textWidth, _ := dc.MeasureString(text)

		if textWidth < float64(boxSize) {
			lowerBound += mid
		} else {
			upperBound = mid
		}
	}

	return lowerBound
}

func BackdropHandler(c *fiber.Ctx) error {
	data := new(Backdrop)
	err := c.BodyParser(data)
	if data.ID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "eventId is required"})
	}
	if data.Rows <= 0 || data.Columns <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Rows and Columns must be positive integers"})
	}
	if data.Width <= 0 || data.Height <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Width and Height must be positive integers"})
	}

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	dc := gg.NewContext(data.Width, data.Height)
	dc.SetColor(color.Black)
	dc.Clear()

	GridCellWidth := data.Width / data.Columns
	GridCellHeight := data.Height / data.Rows

	fontSize := calculateFontSize(int(math.Min(float64(GridCellWidth), float64(GridCellHeight))), fmt.Sprintf("R%dC%d", data.Rows, data.Columns))
	fmt.Println("Font size: ", fontSize)
	dc.SetColor(color.White)
	dc.LoadFontFace("./fonts/roboto.ttf", fontSize)

	rowSpacing := float64(data.Height) / float64(data.Rows)
	columnSpacing := float64(data.Width) / float64(data.Columns)

	for i := 0; i < data.Rows; i++ {
		y := float64(i)*rowSpacing + rowSpacing/2
		dc.DrawLine(0, float64(i)*rowSpacing+0.5, float64(data.Width), float64(i)*rowSpacing+0.5)
		dc.Stroke()

		for j := 0; j < data.Columns; j++ {
			x := float64(j)*columnSpacing + columnSpacing/2
			label := fmt.Sprintf("R%dC%d", i+1, j+1)
			textWidth, textHeight := dc.MeasureString(label)
			dc.DrawString(label, x-textWidth/2, y+textHeight/4)
		}
	}

	for i := 0; i < data.Columns; i++ {
		dc.DrawLine(float64(i)*columnSpacing+0.5, 0, float64(i)*columnSpacing+0.5, float64(data.Height))
		dc.Stroke()
	}

	buffer := new(bytes.Buffer)
	err = png.Encode(buffer, dc.Image())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to encode image"})
	}

	imageData, err := aws.UploadImageFromBuffer(buffer, data.ID, "gridImage")

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to upload image to S3"})

	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": imageData})
}
