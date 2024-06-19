package service

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
	"strings"
	"sync"
	"time"

	"github.com/nfnt/resize"
)

type gridCell struct {
	img      image.Image
	fileName string
}

var (
	processedFiles = sync.Map{}
)

func OverlayImages(inputFile, gridCellFolder, outputFolder string, opacity float64) {
	inputImg, err := LoadImage(inputFile)
	if err != nil {
		fmt.Println(inputImg)
		fmt.Println("Error loading input image:", err)
		return
	}

	gridCells, err := LoadGridCells(gridCellFolder)
	println(gridCells)
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

	bounds := resizedInputImg.Bounds()
	outputImg := image.NewRGBA(bounds)

	draw.Draw(outputImg, bounds, resizedInputImg, image.Point{}, draw.Src)
	
	for i, gridCell := range gridCells {
		if i >= bounds.Dx()/gridCellSize.X*bounds.Dy()/gridCellSize.Y {
			break
		}

		alphaImg := gridCell.alpha(opacity)

		draw.DrawMask(outputImg, gridCell.img.Bounds(), gridCell.img, image.Point{}, alphaImg, image.Point{}, draw.Over)
		// Move the processed grid cell to the processed images folder
		err := MoveProcessedImage(gridCellFolder, processedImageFolder, gridCell.fileName)
		if err != nil {
			fmt.Println("Error moving processed image:", err)
			return
		}
   
		// Remove the processed grid cell from the slice
		gridCells = append(gridCells[:i], gridCells[i+1:]...)
		outputFilePath := filepath.Join(outputFolder, gridCell.fileName)
		err = SaveImage(outputImg, outputFilePath)

		if err != nil {
			fmt.Println("Error saving output image:", err)
		}
	}
}

func ResizeAndCrop(img image.Image, size image.Point) image.Image {
	resizedImg := resize.Resize(uint(size.X), uint(size.Y), img, resize.Lanczos3)

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

func HandleNewFile(filePath, gridCellFolder, outputFolder string, opacity float64) {
	processedFiles.Store(filePath, struct{}{})

	if isSupportedFormat(filePath) {
		fmt.Println("New file detected:", filePath)
		OverlayImages(filePath, gridCellFolder, outputFolder, opacity)
	}
}

// move the processed image to ProcessedImage folder
func MoveProcessedImage(gridCellFolder, processedImageFolder, fileName string) error {
    sourcePath := filepath.Join(gridCellFolder, fileName)
    destPath := filepath.Join(processedImageFolder, fileName)

    err := os.Rename(sourcePath, destPath)
    if err != nil {
        return err
    }

    return nil
}
