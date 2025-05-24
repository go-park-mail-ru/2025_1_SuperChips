package boardshr

import "context"

func (b *BoardShrServicer) DeleteCoauthor(ctx context.Context, boardID int, userID int, coauthorName string) error {
	// Проверка, что пользователь является автором доски.
	isAuthor, err := b.repo.IsBoardAuthor(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isAuthor {
		return ErrForbbiden
	}

	// Получение ID соавтора по имени.
	coauthorID, err := b.repo.GetUserIDFromUsername(ctx, coauthorName)
	if err != nil {
		return err
	}

	// Удаление соавтора.
	err = b.repo.DeleteCoauthor(ctx, boardID, coauthorID)
	if err != nil {
		return err
	}

	return nil
}