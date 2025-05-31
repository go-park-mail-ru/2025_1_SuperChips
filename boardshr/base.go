package boardshr

import (
	"context"
	"path/filepath"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type BoardShrRepository interface {
	CreateInvitation(ctx context.Context, boardID int, UserID int, invitation domain.Invitaion, inviteeIDs []int) (string, error)
	DeleteInvitation(ctx context.Context, boardID int, link string) error
	GetInvitationLinks(ctx context.Context, boardID int) ([]domain.LinkParams, error)

	AddBoardCoauthorByLink(ctx context.Context, boardID int, userID int, link string) error
	DeleteCoauthor(ctx context.Context, boardID int, userID int) error
	GetAuthor(ctx context.Context, boardID int) (domain.Contact, error)
	GetCoauthors(ctx context.Context, boardID int) ([]domain.Contact, error)

	IsBoardAuthor(ctx context.Context, boardID int, userID int) (bool, error)
	IsBoardEditor(ctx context.Context, boardID int, userID int) (bool, error)

	GetUserIDsFromUsernames(ctx context.Context, names []string) ([]NameToID, error)
	GetUserIDFromUsername(ctx context.Context, name string) (int, error)
	GetUsernameFromUserID(ctx context.Context, userID int) (string, error)

	GetLinkParams(ctx context.Context, link string) (int, domain.LinkParams, error)
}

type BoardShrService struct {
	repo BoardShrRepository
	baseURL   string
	staticDir string
	avatarDir string
}

func NewBoardShrService(b BoardShrRepository, baseURL, staticDir, avatarDir string) *BoardShrService {
	return &BoardShrService{
		repo: b,
		baseURL: baseURL,
		staticDir: staticDir,
		avatarDir: avatarDir,
	}
}

func (b *BoardShrService) generateAvatarURL(filename string) string {
	if filename == "" {
		return ""
	}

	return b.baseURL + filepath.Join(b.staticDir, b.avatarDir, filename)
}
