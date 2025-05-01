package subscription

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type SubscriptionRepository interface {
	GetUserFollowers(ctx context.Context, id int) ([]domain.PublicUser, error)
	GetUserFollowing(ctx context.Context, id int) ([]domain.PublicUser, error)
	CreateSubscription(ctx context.Context, targetUsername string, currentID int) error
	DeleteSubscription(ctx context.Context, targetUsername string, currentID int) error	
}

type SubscriptionService struct {
	repo SubscriptionRepository
}

func NewSubscriptionUsecase(repo SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{
		repo: repo,
	}
}

func (service *SubscriptionService) GetUserFollowers(ctx context.Context, id int) ([]domain.PublicUser, error) {
	return service.repo.GetUserFollowers(ctx, id)
}

func (service *SubscriptionService) GetUserFollowing(ctx context.Context, id int) ([]domain.PublicUser, error) {
	return service.repo.GetUserFollowing(ctx, id)
}

func (service *SubscriptionService) CreateSubscription(ctx context.Context, targetUsername string, currentID int) error {
	return service.repo.CreateSubscription(ctx, targetUsername, currentID)
}

func (service *SubscriptionService) DeleteSubscription(ctx context.Context, targetUsername string, currentID int) error {
	return service.repo.DeleteSubscription(ctx, targetUsername, currentID)
}

