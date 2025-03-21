package pin

import (
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type PinRepository interface {
	GetPins(page int, pageSize int) []domain.PinData
}

type PinService struct {
	repo PinRepository
}

func NewPinService(r PinRepository) *PinService {
	return &PinService{
		repo: r,
	}
}

func (p *PinService) GetPins(page int, pageSize int) []domain.PinData {
	return p.repo.GetPins(page, pageSize)
}

