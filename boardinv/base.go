package boardinv

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type BoardInvRepository interface {
	CreateInvitation(ctx context.Context, boardID int, UserID int, invitation domain.Invitaion, inviteeIDs []int) (string, error)
	DeleteInvitation(ctx context.Context, boardID int, link string) error
	GetInvitationLinks(ctx context.Context, boardID int) ([]domain.LinkParams, error)

	IsBoardAuthor(ctx context.Context, boardID int, userID int) (bool, error)
	GetUserIDsFromUsernames(ctx context.Context, inviteeNames []string) ([]NameToID, error)
}

type BoardInvServicer struct {
	repo BoardInvRepository
}

func NewBoardInvService(b BoardInvRepository) *BoardInvServicer {
	return &BoardInvServicer{
		repo: b,
	}
}
