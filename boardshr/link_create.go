package boardshr

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type NameToID struct {
	Username string
	ID       *int
}

func (b *BoardShrService) CreateInvitation(ctx context.Context, boardID int, userID int, invitation domain.Invitaion) (string, []string, error) {
	// Проверка, что пользователь является автором доски.
	isAuthor, err := b.repo.IsBoardAuthor(ctx, boardID, userID)
	if !isAuthor || err != nil {
		return "", nil, domain.ErrForbidden
	}

	// Индексы валидных имён и невалидные имена.
	validInviteeIDs := make([]int, 0)
	invalidInviteeNames := make([]string, 0)
	
	if invitation.Names != nil && len(*invitation.Names) != 0 {
		// Получение ID пользователей по их именам.
		inviteesData, err := b.repo.GetUserIDsFromUsernames(ctx, *invitation.Names)
		if err != nil {
			return "", nil, err
		}

		// Фильтрация существующих и несуществующих имён.
		for _, inviteeData := range inviteesData {
			if inviteeData.ID != nil {
				validInviteeIDs = append(validInviteeIDs, *inviteeData.ID)
			} else {
				invalidInviteeNames = append(invalidInviteeNames, inviteeData.Username)
			}
		}
	}

	// Создание ссылки. Персональная ссылка создаётся только для существующих имён.
	link, err := b.repo.CreateInvitation(ctx, boardID, userID, invitation, validInviteeIDs)
	if err != nil {
		return "", nil, err
	}

	// В случае частичного успеха также возвращаются имена, которых нет в БД.
	if len(invalidInviteeNames) != 0 {
		return link, invalidInviteeNames, ErrNonExistentUsername
	}

	return link, nil, nil
}
