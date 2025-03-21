package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type PinSlice struct {
	Pins []domain.PinData
}

func NewPinSliceStorage(cfg configs.Config) PinSlice {
    p := PinSlice{}
    p.initialize(cfg)

    return p
}

func (p *PinSlice) GetPins(page int, pageSize int) []domain.PinData {
    startIndex := (page - 1) * pageSize
    endIndex := startIndex + pageSize
    if endIndex > len(p.Pins) {
        endIndex = len(p.Pins)
    }

    if startIndex >= len(p.Pins) {
        return make([]domain.PinData, 0)
    }

    pagedImages := p.Pins[startIndex:endIndex]

    return pagedImages
}

func (p *PinSlice) initialize(cfg configs.Config) {
	baseDir := cfg.ImageBaseDir

	files, err := os.ReadDir(baseDir)
    if err != nil {
        return
    }

	id := 1

    for _, file := range files {
        if !file.IsDir() && isImageFile(file.Name()) {
            p.Pins = append(p.Pins, domain.PinData{
				Header: fmt.Sprintf("Header %d", id),
				Image: fmt.Sprintf("https://yourflow.ru/static/img/%s", file.Name()),
				Author: fmt.Sprintf("Author %d", -id),
			})
			id++
        }
    }

}

func isImageFile(filename string) bool {
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

