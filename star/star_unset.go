package star

import "context"

func (s *StarService) UnSetStarProperty(ctx context.Context, userID int, pinID int) error {
	err := s.starRep.UnSetStarProperty(ctx, userID, pinID)
	if err != nil {
		return err
	}

	return nil
}