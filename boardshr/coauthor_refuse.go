package boardshr

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

func (b *BoardShrServicer) RefuseCoauthoring(ctx context.Context, boardID int, userID int) error {
	// Проверка, что пользователь является автором или соавтором доски.
	isEditor, err := b.repo.IsBoardEditor(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isEditor {
		return domain.ErrForbidden
	}

	// Проверка, что пользователь не является автором доски (автор не может покинуть доску).
	isAuthor, err := b.repo.IsBoardAuthor(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if isAuthor {
		return ErrAuthorRefuseEditing
	}

	err = b.repo.DeleteCoauthor(ctx, boardID, userID)
	if err != nil {
		return err
	}

	return nil
}
