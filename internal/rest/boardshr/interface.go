package rest

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type BoardShrServicer interface {
	CreateInvitation(ctx context.Context, boardID int, userID int, invitation domain.Invitaion) (string, []string, error)
	DeleteInvitation(ctx context.Context, boardID int, userID int, link string) error
	GetInvitationLinks(ctx context.Context, boardID int, userID int) ([]domain.LinkParams, error)
	UseInvitationLink(ctx context.Context, userID int, link string) (int, error)
	RefuseCoauthoring(ctx context.Context, boardID int, userID int) error
	GetCoauthors(ctx context.Context, boardID int, userID int) ([]string, error)
	DeleteCoauthor(ctx context.Context, boardID int, userID int, coauthorName string) error
}
