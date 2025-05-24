package boardshr

import (
	"context"
	"slices"
	"time"
)

func (b *BoardShrServicer) UseInvitationLink(ctx context.Context, userID int, link string) error {
	// Проверка, что ссылка действительная.
	boardID, linkParams, err := b.repo.GetLinkParams(ctx, link)
	if err != nil {
		return err
	}

	// Проверка, что пользователь не является автором или соавтором доски.
	isEditor, err := b.repo.IsBoardEditor(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if isEditor {
		return ErrAlreadyEditor
	}

	// Для персональных ссылок: проверка на право пользования.
	if linkParams.Names != nil {
		name, err := b.repo.GetUsernameFromUserID(ctx, userID)
		if err != nil {
			return err
		}
		if !slices.Contains(*linkParams.Names, name) {
			return ErrForbbiden
		}
	}

	// Проверка на истечение времени активности ссылки.
	if linkParams.TimeLimit != nil && (*linkParams.TimeLimit).Before(time.Now()) {
		return ErrLinkExpired
	}

	// Проверка на превышение количества использований ссылки.
	if linkParams.UsageLimit != nil && linkParams.UsageCount >= *linkParams.UsageLimit {
		return ErrLinkExpired
	}

	// Добавление пользователя в соавторы доски по ссылке.
	err = b.repo.AddBoardCoauthorByLink(ctx, boardID, userID, link)
	if err != nil {
		return err
	}

	return nil
}
