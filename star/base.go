package star

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type StarRepository interface {
	GetStarPins(ctx context.Context, userID int) ([]domain.PinData, error)

	SetStarProperty(ctx context.Context, userID int, pinID int) error
	UnSetStarProperty(ctx context.Context, userID int, pinID int) error
	ReassignStarProperty(ctx context.Context, userID int, oldPinID int, newPinID int) error

	GetPinsCount(ctx context.Context, userID int) (int, error)
}

type StarService struct {
	starRep StarRepository
}

func NewStarPinService(rep StarRepository) *StarService {
	return &StarService{
		starRep: rep,
	}
}