package rest

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type BoardInvServicer interface {
	CreateInvitation(ctx context.Context, boardID int, userID int, invitation domain.Invitaion) (string, []string, error)
	DeleteInvitation(ctx context.Context, boardID int, userID int, link string) error
	GetInvitationLinks(ctx context.Context, boardID int, userID int) ([]domain.LinkParams, error)
}
