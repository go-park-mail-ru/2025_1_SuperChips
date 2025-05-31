package like

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/validator"
)

type LikeRepository interface {
	LikeFlow(ctx context.Context, pinID, userID int) (string, string, error)
}

type PinRepository interface {
	GetPin(ctx context.Context, pinID, userID uint64) (domain.PinData, uint64, error)
}

type LikeService struct {
	likeRepository LikeRepository
	pinRepo PinRepository
}

func NewLikeService(likeRepository LikeRepository, pinRepo PinRepository) *LikeService {
	return &LikeService{
		likeRepository: likeRepository,
		pinRepo: pinRepo,
	}
}

func (service *LikeService) LikeFlow(ctx context.Context, pinID, userID int) (string, string, error) {
	v := validator.New()

	if !v.Check(pinID > 0 && userID > 0, "id", "cannot be less than or equal to zero") {
		return "", "", v.GetError("id")
	}

	action, username, err := service.likeRepository.LikeFlow(ctx, pinID, userID)
	if err != nil {
		return "", "", err
	}

	if action == "insert" {
		action = "liked"
	} else {
		action = "unliked"
	}

	return action, username, nil
}

