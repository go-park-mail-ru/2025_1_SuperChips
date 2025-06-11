package star

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

func (s *StarService) GetStarPins(ctx context.Context, userID int) (int, []domain.PinData, error) {
	starPins, err := s.starRep.GetStarPins(ctx, userID)
	if err != nil {
		return 0, nil, err
	}

	slotsCount, err := s.GetSlotsCount(ctx, userID)
	if err != nil {
		return 0, nil, err
	}

	return slotsCount, starPins, err
}
