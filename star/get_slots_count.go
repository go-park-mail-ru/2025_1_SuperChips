package star

import "context"


func (s *StarService) GetSlotsCount(ctx context.Context, userID int) (int, error) {
	pinsCount, err := s.starRep.GetPinsCount(ctx, userID)
	if err != nil {
		return 0, err
	}
	
	slotsCount := 0
	switch {
	case pinsCount >=  1 && pinsCount <  7: slotsCount = 1
	case pinsCount >=  7 && pinsCount < 15: slotsCount = 2
	case pinsCount >= 15 && pinsCount < 20: slotsCount = 3
	case pinsCount >= 20:                   slotsCount = 4 + (pinsCount - 20) / 10
	}

	return slotsCount, nil
}