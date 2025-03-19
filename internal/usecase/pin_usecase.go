package usecase

import (
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/entity"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository"
)

type PinService struct {
	repo repository.PinStorage
}

func NewPinService(repo repository.PinStorage) *PinService {
	return &PinService{
		repo: repo,
	}
}

func (p PinService) GetPins(page int, pageSize int) []entity.PinData {
	return p.repo.GetPinPage(page, pageSize)
}

