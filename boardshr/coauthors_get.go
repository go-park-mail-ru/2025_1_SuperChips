package boardshr

import "context"

func (b *BoardShrServicer) GetCoauthors(ctx context.Context, boardID int, userID int) ([]string, error) {
	// Проверка, что пользователь является автором доски.
	isAuthor, err := b.repo.IsBoardAuthor(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isAuthor {
		return nil, ErrForbbiden
	}

	names, err := b.repo.GetCoauthors(ctx, boardID)
	if err != nil {
		return nil, err
	}

	return names, nil
}
