package feed

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type StatusError interface {
	error
	StatusCode() int
}

type statusError struct {
	code int
	msg  string
}

func (e *statusError) Error() string {
	return e.msg
}

func (e *statusError) StatusCode() int {
	return e.code
}

func isImageFile(filename string) bool {
    ext := strings.ToLower(filepath.Ext(filename))
    switch ext {
    case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
        return true
    default:
        return false
    }
}

type PinSlice struct {
	Pins []PinData
}

func (p *PinSlice) Initialize() {
	baseDir := "./static/img"

	files, err := os.ReadDir(baseDir)
    if err != nil {
        return
    }

	id := 1

    for _, file := range files {
        if !file.IsDir() && isImageFile(file.Name()) {
            p.Pins = append(p.Pins, PinData{
				Header: fmt.Sprintf("Header %d", id),
				Image: file.Name(),
				Author: fmt.Sprintf("Author %d", -id),
			})
			id++
        }
    }

}