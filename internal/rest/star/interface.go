package rest

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type StarServicer interface {
	GetStarPins(ctx context.Context, userID int) (int, []domain.PinData, error)

	SetStarProperty(ctx context.Context, userID int, pinID int) error
	UnSetStarProperty(ctx context.Context, userID int, pinID int) error
	ReassignStarProperty(ctx context.Context, userID int, oldPinID int, newPinID int) error
}
