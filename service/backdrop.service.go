package service

import (
	"github.com/fogleman/gg"
)

func CalculateFontSize(boxSize int, text string) float64 {
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
