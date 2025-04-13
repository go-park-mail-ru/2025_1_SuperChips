package pincrud

import (
	"mime/multipart"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type PinCRUDRepository interface {
	GetPin(pinID uint64) (domain.PinData, uint64, error)

	DeletePin(pinID uint64, userID uint64) error

	UpdatePin(patch domain.PinDataUpdate, userID uint64) error

	CreatePin(data domain.PinDataCreate, imgName string, userID uint64) (uint64, error)
}

type FileRepository interface {
	Save(file multipart.File, header *multipart.FileHeader) (string, error)
	Delete(imgName string) error
}

type PinCRUDService struct {
	rep     PinCRUDRepository
	imgStrg FileRepository
}
