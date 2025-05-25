package comment

import (
	"context"
	"path/filepath"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type CommentRepository interface {
	GetComments(ctx context.Context, flowID, userID, page, size int) ([]domain.Comment, error)
	LikeComment(ctx context.Context, commentID, userID int) (string, error)
	AddComment(ctx context.Context, flowID, userID int, content string) error
	DeleteComment(ctx context.Context, commentID, userID int) error
}

type PinRepository interface {
	GetPin(ctx context.Context, pinID, userID uint64) (domain.PinData, uint64, error)
}

type CommentService struct {
	repo CommentRepository
	pinRepo PinRepository
	avatarDir string
	staticDir string
	baseURL string
}

func NewCommentService(repo CommentRepository, pinRepo PinRepository, baseURL, staticDir, avatarDir string) *CommentService {
	return &CommentService{
		repo: repo,
		pinRepo: pinRepo,
		avatarDir: avatarDir,
		staticDir: staticDir,
		baseURL: baseURL,
	}
}

func (s *CommentService) GetComments(ctx context.Context, flowID, userID, page, size int) ([]domain.Comment, error) {
	comments, err := s.repo.GetComments(ctx, flowID, userID, page, size)
	if err != nil {
		return nil, err
	}

	for i := range comments {
		if !comments[i].AuthorIsExternalAvatar {
			comments[i].AuthorAvatar = s.generateAvatarURL(comments[i].AuthorAvatar)
		}
	}

	return comments, nil
}

func (s *CommentService) LikeComment(ctx context.Context, flowID, commentID, userID int) (string, error) {
	flow, authorID, err := s.pinRepo.GetPin(ctx, uint64(flowID), uint64(userID))
	if err != nil {
		return "", err
	}

	if flow.IsPrivate && authorID != uint64(userID) {
		return "", domain.ErrForbidden
	}
	
	like, err := s.repo.LikeComment(ctx, commentID, userID)
	if err != nil {
		return "", err
	}

	return like, nil
}

func (s *CommentService) AddComment(ctx context.Context, flowID, userID int, content string) error {
	flow, authorID, err := s.pinRepo.GetPin(ctx, uint64(flowID), uint64(userID))
	if err != nil {
		return err
	}

	if flow.IsPrivate && authorID != uint64(userID) {
		return domain.ErrForbidden
	}

	if err := s.repo.AddComment(ctx, flowID, userID, content); err != nil {
		return err
	}

	return nil
}

func (s* CommentService) DeleteComment(ctx context.Context, commentID, userID int) error {
	if err := s.repo.DeleteComment(ctx, commentID, userID); err != nil {
		return err
	}

	return nil
}	

func (s *CommentService) generateAvatarURL(filename string) string {
	if filename == "" {
		return ""
	}

	return s.baseURL + filepath.Join(s.staticDir, s.avatarDir, filename)
}
