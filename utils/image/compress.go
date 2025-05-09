package image

import (
	"fmt"
	"image"
	"image/color"

	"golang.org/x/image/draw"
)

// this function returns 4 main colors of the image
// in hex format
func GetImageMainColors(img image.Image) []string {
	// compresses image to 2x2
	compressed := compressImage(img)

	return get2x2Colors(compressed)
}

func compressImage(img image.Image) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, 2, 2))

	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	return dst
}

func сolorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02X%02X%02X", r>>8, g>>8, b>>8)
}

func get2x2Colors(img image.Image) []string {
	var colors [4]color.Color
	colors[0] = img.At(0, 0)
	colors[1] = img.At(1, 0)
	colors[2] = img.At(0, 1)
	colors[3] = img.At(1, 1)
	
	var colorsHex []string

	for i := range colors {
		colorsHex = append(colorsHex, сolorToHex(colors[i]))
	}

	return colorsHex
}

