package pincrud

import (
	"context"
	"mime/multipart"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type PinCRUDRepository interface {
	GetPin(ctx context.Context, pinID, userID uint64) (domain.PinData, uint64, error)

	DeletePin(ctx context.Context, pinID uint64, userID uint64) error

	UpdatePin(ctx context.Context, patch domain.PinDataUpdate, userID uint64) error

	CreatePin(ctx context.Context, data domain.PinDataCreate, imgName string, userID uint64) (uint64, error)

	GetPinCleanMediaURL(ctx context.Context, pinID uint64) (string, uint64, error)
}

type FileRepository interface {
	Save(file multipart.File, header *multipart.FileHeader) (string, error)
	Delete(imgName string) error
}
