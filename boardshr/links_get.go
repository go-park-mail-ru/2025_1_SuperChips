package boardshr

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

func (b *BoardShrServicer) GetInvitationLinks(ctx context.Context, boardID int, userID int) ([]domain.LinkParams, error) {
	// Проверка, что пользователь является автором доски.
	isAuthor, err := b.repo.IsBoardAuthor(ctx, boardID, userID)
	if !isAuthor || err != nil {
		return nil, domain.ErrForbidden
	}

	// Получение ссылок на доску.
	links, err := b.repo.GetInvitationLinks(ctx, boardID)
	if err != nil {
		return nil, err
	}

	return links, nil
}
