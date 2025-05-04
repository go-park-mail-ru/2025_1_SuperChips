package subscription

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type SubscriptionRepository interface {
	GetUserFollowers(ctx context.Context, id, page, size int) ([]domain.PublicUser, error)
	GetUserFollowing(ctx context.Context, id, page, size int) ([]domain.PublicUser, error)
	CreateSubscription(ctx context.Context, targetUsername string, currentID int) error
	DeleteSubscription(ctx context.Context, targetUsername string, currentID int) error	
}

type ChatRepository interface {
	CreateContact(ctx context.Context, username, targetUsername string) error
}

type SubscriptionService struct {
	subRepo SubscriptionRepository
	chatRepo ChatRepository
}

func NewSubscriptionUsecase(repo SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{
		subRepo: repo,
	}
}

func (service *SubscriptionService) GetUserFollowers(ctx context.Context, id, page, size int) ([]domain.PublicUser, error) {
	return service.subRepo.GetUserFollowers(ctx, id, page, size)
}

func (service *SubscriptionService) GetUserFollowing(ctx context.Context, id, page, size int) ([]domain.PublicUser, error) {
	return service.subRepo.GetUserFollowing(ctx, id, page, size)
}

func (service *SubscriptionService) CreateSubscription(ctx context.Context, targetUsername string, currentID int) error {
	return service.subRepo.CreateSubscription(ctx, targetUsername, currentID)
}

func (service *SubscriptionService) DeleteSubscription(ctx context.Context, targetUsername string, currentID int) error {
	return service.subRepo.DeleteSubscription(ctx, targetUsername, currentID)
}

