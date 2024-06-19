package service

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nfnt/resize"
	// "github.com/nfnt/resize"
)

type gridCell struct {
	img      image.Image
	fileName string
}


func OverlayImages(inputFile, gridCellFolder, outputFolder string, opacity float64) {
	fmt.Print(inputFile)
	inputImg, err := LoadImage(inputFile)
	if err != nil {
		fmt.Println(inputImg)
		fmt.Println("Error loading input image:", err)
		return
	}

	gridCells, err := LoadGridCells(gridCellFolder)
	// println(gridCells)
	if err != nil {
		fmt.Println("gridCell")
		fmt.Println("Error loading grid cells:", err)
		return
	}

	if len(gridCells) == 0 {
		fmt.Println("No grid cells found")
		return
	}
	processedImageFolder := filepath.Join(gridCellFolder, "processed")
	err = os.MkdirAll(processedImageFolder, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating processed images folder:", err)
		return
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(gridCells), func(i, j int) {
		gridCells[i], gridCells[j] = gridCells[j], gridCells[i]
	})

	gridCellSize := gridCells[0].img.Bounds().Size()
	resizedInputImg := ResizeAndCrop(inputImg, gridCellSize)
	// fmt.Println(resizedInputImg);
	bounds := resizedInputImg.Bounds()
	outputImg := image.NewRGBA(bounds)

	draw.Draw(outputImg, bounds, resizedInputImg, image.Point{}, draw.Src)

	for i, gridCell := range gridCells {
		fmt.Println("entry")
		if i >= bounds.Dx()/gridCellSize.X*bounds.Dy()/gridCellSize.Y {
			fmt.Println("break")
			break
		}

		alphaImg := gridCell.alpha(opacity)
		// fmt.Println(alphaImg)
		draw.DrawMask(outputImg, gridCell.img.Bounds(), gridCell.img, image.Point{}, alphaImg, image.Point{}, draw.Over)

		err := MoveProcessedImage(gridCellFolder, processedImageFolder, gridCell.fileName)
		if err != nil {
			fmt.Println("Error moving processed image:", err)
			return
		}
		// Remove the processed grid cell from the slice
		gridCells = append(gridCells[:i], gridCells[i+1:]...)
		outputFilePath := filepath.Join(outputFolder, gridCell.fileName)
		err = SaveImage(outputImg, outputFilePath)
		fmt.Println("exit")
		if err != nil {
			fmt.Println("Error saving output image:", err)
		}
	}
}

func ResizeAndCrop(img image.Image, size image.Point) image.Image {

	originalBounds := img.Bounds()
	originalWidth := originalBounds.Dx()
	originalHeight := originalBounds.Dy()

	minDim := math.Min(float64(originalWidth), float64(originalHeight))

	var xOffset, yOffset int
	if originalWidth > originalHeight {
		xOffset = (originalWidth - int(minDim)) / 2
	} else {
		yOffset = (originalHeight - int(minDim)) / 2
	}

	cropRect := image.Rect(xOffset, yOffset, xOffset+int(minDim), yOffset+int(minDim))

	croppedImg := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(cropRect)

	resizedImg := resize.Resize(uint(size.X), uint(size.Y), croppedImg, resize.Lanczos3)

	return resizedImg
}

func LoadImage(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	return img, err
}

func LoadGridCells(gridCellFolder string) ([]gridCell, error) {
	var gridCells []gridCell
	err := filepath.Walk(gridCellFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && isSupportedFormat(path) {
			img, err := LoadImage(path)
			if err != nil {
				return err
			}
			gridCells = append(gridCells, gridCell{img: img, fileName: info.Name()})
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

func SaveImage(img image.Image, filePath string) error {
	outFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(outFile, img, nil)
	case ".png":
		err = png.Encode(outFile, img)
	default:
		err = fmt.Errorf("unsupported file extension")
	}
	return err
}

func isSupportedFormat(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".jpg", ".jpeg", ".png":
		return true
	default:
		return false
	}
}



func MoveProcessedImage(gridCellFolder, processedImageFolder, fileName string) error {
    sourcePath := filepath.Join(gridCellFolder, fileName)
    destPath := filepath.Join(processedImageFolder, fileName)

    // Ensure the destination directory exists
    err := os.MkdirAll(processedImageFolder, os.ModePerm)
    if err != nil {
        return fmt.Errorf("error creating processed images folder: %w", err)
    }

    // Check if the source file exists
    if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
        return fmt.Errorf("source file does not exist: %s", sourcePath)
    }

    // Attempt to move the file
    err = os.Rename(sourcePath, destPath)
    if err != nil {
        return fmt.Errorf("error moving processed image from %s to %s: %w", sourcePath, destPath, err)
    }

    return nil
}