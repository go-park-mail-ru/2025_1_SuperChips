package image

import (
	"image"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	_ "image/png"
	_ "image/jpeg"
	_ "image/gif"
)

func IsImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))

	pattern := "*.jpg;*.jpeg;*.png;*.gif;*.bmp;*.tiff;*.webp"

	for _, p := range strings.Split(pattern, ";") {
		match, err := filepath.Match(p, ext)
		if err != nil {
			continue
		}
		if match {
			return true
		}
	}

	return false
}

func UploadImage(imageFilename, staticDir, imageDir, baseUrl string, file io.Reader) (string, string, error) {
	ext := filepath.Ext(imageFilename)
	filename := uuid.New().String() + ext
	filePath := filepath.Join(staticDir, imageDir, filename)
	fileDir := filepath.Join(staticDir, imageDir)

	if err := os.MkdirAll(filepath.Join(".", fileDir), os.ModePerm); err != nil {
		return "", "", err
	}

	dst, err := os.Create(filepath.Join(".", filePath))
	if err != nil {
		return "", "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", "", err
	}

	url := baseUrl + filePath

	return filename, url, nil
}

func GetImageDimensions(img image.Image) (int, int, error) {
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	return width, height, nil
}
