package boardshr

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

func (b *BoardShrService) GetCoauthors(ctx context.Context, boardID int, userID int) (domain.Contact, []domain.Contact, error) {
	// Проверка, что пользователь является автором или соавтором доски.
	isAuthor, err := b.repo.IsBoardEditor(ctx, boardID, userID)
	if err != nil {
		return domain.Contact{}, nil, err
	}
	if !isAuthor {
		return domain.Contact{}, nil, ErrForbbiden
	}

	author, err := b.repo.GetAuthor(ctx, boardID)
	if err != nil {
		return domain.Contact{}, nil, err
	}

	coauthors, err := b.repo.GetCoauthors(ctx, boardID)
	if err != nil {
		return domain.Contact{}, nil, err
	}

	// Генерация ссылок на аватары.
	if !author.IsExternalAvatar {
		author.Avatar = b.generateAvatarURL(author.Avatar)
	}
	for i := range coauthors {
		if !coauthors[i].IsExternalAvatar {
			coauthors[i].Avatar = b.generateAvatarURL(coauthors[i].Avatar)
		}
	}

	return author, coauthors, nil
}
