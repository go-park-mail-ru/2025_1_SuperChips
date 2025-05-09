package rest

import (
	"context"
	"mime/multipart"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type PinCRUDServicer interface {
	GetPublicPin(ctx context.Context, pinID uint64) (domain.PinData, error)
	GetAnyPin(ctx context.Context, pinID uint64, userID uint64) (domain.PinData, error)
	DeletePin(ctx context.Context, pinID uint64, userID uint64) error
	UpdatePin(ctx context.Context, data domain.PinDataUpdate, userID uint64) error
	CreatePin(ctx context.Context, data domain.PinDataCreate, file multipart.File, header *multipart.FileHeader, extension string, userID uint64) (uint64, error)
}
