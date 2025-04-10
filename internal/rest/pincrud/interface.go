package rest

import (
	"mime/multipart"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type PinCRUDServiceInterface interface {
	GetPublicPin(pinID uint64) (domain.PinData, error)
	GetAnyPin(pinID uint64, userID uint64) (domain.PinData, error)

	DeletePinByID(pinID uint64, userID uint64) error

	UpdatePin(data domain.PinDataUpdate, userID uint64) error

	CreatePin(data domain.PinDataCreate, file multipart.File, header *multipart.FileHeader, userID uint64) error
}

type PinCRUDHandler struct {
	Config     configs.Config
	PinService PinCRUDServiceInterface
}
