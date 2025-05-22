package boardinv

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

func (b *BoardInvServicer) DeleteInvitation(ctx context.Context, boardID int, userID int, link string) error {
	// Проверка, что пользователь является автором доски.
	isAuthor, err := b.repo.IsBoardAuthor(ctx, boardID, userID)
	if !isAuthor || err != nil {
		return domain.ErrForbidden
	}
	
	// Удаление ссылки.
	err = b.repo.DeleteInvitation(ctx, boardID, link)
	if err != nil {
		return err
	}

	return nil
}