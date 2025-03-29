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

func UploadImage(imageFilename, staticDir, imageDir string, file io.Reader) error {
	ext := filepath.Ext(imageFilename)
	filename := uuid.New().String() + ext
	filePath := filepath.Join(staticDir, imageDir, filename)

	if err := os.MkdirAll(filepath.Join(staticDir, imageDir), os.ModePerm); err != nil {
		return err
	}

	dst, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return err
	}

	return nil
}
