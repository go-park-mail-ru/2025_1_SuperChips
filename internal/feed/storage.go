package feed

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
)

func isImageFile(filename string) bool {
    ext := strings.ToLower(filepath.Ext(filename))
    switch ext {
    case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
        return true
    default:
        return false
    }
}

type PinStorage struct {
	Pins []PinData
}

func (p *PinStorage) Initialize(cfg configs.Config) {
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
				Image: fmt.Sprintf("http://%s%s/static/img/%s", cfg.IpAddress, cfg.Port, file.Name()),
				Author: fmt.Sprintf("Author %d", -id),
			})
			id++
        }
    }

}