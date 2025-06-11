package star

import "context"

func (s *StarService) SetStarProperty(ctx context.Context, userID int, pinID int) error {
	// Получение всех звёздных пинов.
	starPins, err := s.starRep.GetStarPins(ctx, userID)
	if err != nil {
		return err
	}

	// Проверка, что пин уже является звёздным.
	for _, pin := range starPins {
		if pin.AuthorID == uint64(userID) && pin.FlowID == uint64(pinID) && pin.IsStar {
			return ErrAlreadyStar
		}
	}

	// Проверка, что у пользователя есть свободный звёздный слот.
	slotsCount, err := s.GetSlotsCount(ctx, userID)
	if err != nil {
		return err
	}
	if len(starPins) >= slotsCount {
		return ErrNoFreeStarSlots
	}

	// Озвездюливание пина.
	err = s.starRep.SetStarProperty(ctx, userID, pinID)
	if err != nil {
		return err
	}

	return nil
}
