package like

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/validator"
)

type LikeRepository interface {
	LikeFlow(ctx context.Context, pinID, userID int) (string, string, error)
}

type LikeService struct {
	likeRepository LikeRepository
}

func NewLikeService(likeRepository LikeRepository) *LikeService {
	return &LikeService{
		likeRepository: likeRepository,
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

