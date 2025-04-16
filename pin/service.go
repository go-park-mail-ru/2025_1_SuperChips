package pin

import (
	"path/filepath"
	"strings"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type PinRepository interface {
	GetPins(page int, pageSize int) ([]domain.PinData, error)
}

type PinService struct {
	repo PinRepository
	baseURL string
	imageDir string
}

func NewPinService(r PinRepository, baseURL, imageDir string) *PinService {
	return &PinService{
		repo: r,
		baseURL: baseURL,
		imageDir: imageDir,
	}
}

func (p *PinService) GetPins(page int, pageSize int) ([]domain.PinData, error) {
	pins, err := p.repo.GetPins(page, pageSize)
	if err != nil {
		return []domain.PinData{}, err
	}

	for _, v := range pins {
		v.MediaURL = p.generateImageURL(v.MediaURL)
	}

	return pins, nil
}

func (p *PinService) generateImageURL(filename string) string {
	return p.baseURL + filepath.Join(strings.ReplaceAll(p.imageDir, ".", ""), filename)
}