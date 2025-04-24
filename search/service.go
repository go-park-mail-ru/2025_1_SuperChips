package search

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type SearchRepository interface {
	SearchPins(ctx context.Context, query string, page, pageSize int) ([]domain.PinData, error)
	SearchUsers(ctx context.Context, query string, page, pageSize int) ([]domain.PublicUser, error)
	SearchBoards(ctx context.Context, query string, page, pageSize int) ([]domain.Board, error) 
}

type SearchService struct {
	repo SearchRepository
	baseURL string
	imageDir string
	staticDir string
	avatarDir string
}

func NewSearchService(repo SearchRepository, baseURL, imageDir, staticDir, avatarDir string) *SearchService {
	return &SearchService{
		repo: repo,
		baseURL: baseURL,
		imageDir: imageDir,
		staticDir: staticDir,
		avatarDir: avatarDir,
	}
}

func (s *SearchService) SearchPins(ctx context.Context, query string, page, pageSize int) ([]domain.PinData, error) {
	pins, err := s.repo.SearchPins(ctx, query, page, pageSize)
	if err != nil {
		return nil, err
	}

	for v := range pins {
		pins[v].MediaURL = s.generateImageURL(pins[v].MediaURL)
	}

	return pins, err
}

func (s *SearchService) SearchBoards(ctx context.Context, query string, page, pageSize int) ([]domain.Board, error) {
	return s.repo.SearchBoards(ctx, query, page, pageSize)
}

func (s *SearchService) SearchUsers(ctx context.Context, query string, page, pageSize int) ([]domain.PublicUser, error) {
	users, err := s.repo.SearchUsers(ctx, query, page, pageSize)
	if err != nil {
		return nil, err
	}

	for i := range users {
		users[i].Avatar = s.generateAvatarURL(users[i].Avatar)
	}

	return users, nil
}

func (s *SearchService) generateImageURL(filename string) string {
	return s.baseURL + filepath.Join(strings.ReplaceAll(s.imageDir, ".", ""), filename)
}

func (s *SearchService) generateAvatarURL(filename string) string {
	if filename == "" {
		return ""
	}

	return s.baseURL + filepath.Join(s.staticDir, s.avatarDir, filename)
}

