package repository

import (
	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/entity"
)

type PinStorage interface {
	NewStorage(cfg configs.Config)
    GetPinPage(page int, pageSize int) []entity.PinData
}
