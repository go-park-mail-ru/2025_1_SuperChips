package star

import (
	"context"
	"slices"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

func (s *StarService) ReassignStarProperty(ctx context.Context, userID int, oldPinID int, newPinID int) error {
	// Получение всех звёздных пинов.
	starPins, err := s.starRep.GetStarPins(ctx, userID)
	if err != nil {
		return err
	}

	// Проверка, что новый пин уже является звёздным.
	newPinIsStar := slices.ContainsFunc(starPins, func(pin domain.PinData) bool {
		return pin.AuthorID == uint64(userID) && pin.FlowID == uint64(newPinID) && pin.IsStar
	})
	if newPinIsStar {
		return ErrAlreadyStar
	}

	// Проверка, что старый пин является звёздным.
	oldPinIsStar := slices.ContainsFunc(starPins, func(pin domain.PinData) bool {
		return pin.AuthorID == uint64(userID) && pin.FlowID == uint64(oldPinID) && pin.IsStar
	})
	if !oldPinIsStar {
		return ErrPinIsNotStar
	}

	// Проверка, что у пользователя есть свободные звёздные слоты.
	slotsCount, err := s.GetSlotsCount(ctx, userID)
	if err != nil {
		return err
	}
	// Минус 1, т.к. при превышении количества звёздных пинов после лишения звёздности старого пина для нового пина должен быть свободный слот.
	if len(starPins) - 1 >= slotsCount {
		return ErrNoFreeStarSlots
	}

	// Передача звёздности.
	err = s.starRep.ReassignStarProperty(ctx, userID, oldPinID, newPinID)
	if err != nil {
		return err
	}

	return nil
}