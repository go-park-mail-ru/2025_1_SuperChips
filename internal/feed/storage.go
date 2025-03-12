package feed

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
)

type PinStorage interface {
    CountPages(pageSize int) int
    GetPinPage(page int, pageSize int) []PinData
    
}

type PinSlice struct {
	Pins []PinData
}

func NewPinSliceStorage(cfg configs.Config) *PinSlice {
    newStorage := PinSlice{}
    newStorage.initialize(cfg)

    return &newStorage
}

func (p *PinSlice) CountPages(pageSize int) int {
    return (len(p.Pins) + pageSize - 1) / pageSize
}

func (p *PinSlice) GetPinPage(page int, pageSize int) []PinData {
    startIndex := (page - 1) * pageSize
    endIndex := startIndex + pageSize
    if endIndex > len(p.Pins) {
        endIndex = len(p.Pins)
    }

    if startIndex >= len(p.Pins) {
        return make([]PinData, 0)
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
            p.Pins = append(p.Pins, PinData{
				Header: fmt.Sprintf("Header %d", id),
				Image: fmt.Sprintf("http://%s%s/static/img/%s", cfg.IpAddress, cfg.Port, file.Name()),
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

