package image

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
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
