package subscription

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type SubscriptionRepository interface {
	GetUserFollowers(ctx context.Context, id, page, size int) ([]domain.PublicUser, error)
	GetUserFollowing(ctx context.Context, id, page, size int) ([]domain.PublicUser, error)
	CreateSubscription(ctx context.Context, targetUsername string, currentID int) error
	DeleteSubscription(ctx context.Context, targetUsername string, currentID int) error	
}

type ContactRepository interface {
	AddToContacts(ctx context.Context, username, targetUsername string) error
}

type SubscriptionService struct {
	subRepo SubscriptionRepository
	contactRepo ContactRepository
	baseURL  string
	staticDir string
	avatarDir string
}

func NewSubscriptionUsecase(repo SubscriptionRepository, contactRepo ContactRepository, baseURL, staticDir, avatarDir string) *SubscriptionService {
	return &SubscriptionService{
		subRepo: repo,
		contactRepo: contactRepo,
		baseURL: baseURL,
		staticDir: staticDir,
		avatarDir: avatarDir,
	}
}

func (service *SubscriptionService) GetUserFollowers(ctx context.Context, id, page, size int) ([]domain.PublicUser, error) {
	followers, err := service.subRepo.GetUserFollowers(ctx, id, page, size)
	if err != nil {
		return nil, err
	}

	for i := range followers {
		if !followers[i].IsExternalAvatar {
			followers[i].Avatar = service.generateAvatarURL(followers[i].Avatar)
		}
	}

	return followers, nil
}

func (service *SubscriptionService) GetUserFollowing(ctx context.Context, id, page, size int) ([]domain.PublicUser, error) {
	following, err := service.subRepo.GetUserFollowing(ctx, id, page, size)
	if err != nil {
		return nil, err
	}

	for i := range following {
		if !following[i].IsExternalAvatar {
			following[i].Avatar = service.generateAvatarURL(following[i].Avatar)
		}
	}

	return following, nil
}

func (service *SubscriptionService) CreateSubscription(ctx context.Context, username, targetUsername string, currentID int) error {
	err := service.contactRepo.AddToContacts(ctx, username, targetUsername)
	if err != nil && !errors.Is(err, domain.ErrConflict) {
		return err
	}

	return service.subRepo.CreateSubscription(ctx, targetUsername, currentID)
}

func (service *SubscriptionService) DeleteSubscription(ctx context.Context, targetUsername string, currentID int) error {
	return service.subRepo.DeleteSubscription(ctx, targetUsername, currentID)
}

func (s *SubscriptionService) generateAvatarURL(filename string) string {
	if filename == "" {
		return ""
	}

	return s.baseURL + filepath.Join(s.staticDir, s.avatarDir, filename)
}

