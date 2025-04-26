package poll

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)


type PollRepository interface {
	GetAllPolls(ctx context.Context) ([]domain.Poll, error)
	AddAnswer(ctx context.Context, pollID, userID uint64, answers []domain.Answer) error
}

type PollService struct {
	repo PollRepository
}

func NewPollService(repo PollRepository) *PollService {
	return &PollService{
		repo: repo,
	}
}

func (p *PollService) GetAllPolls(ctx context.Context) ([]domain.Poll, error) {
	return p.repo.GetAllPolls(ctx)
}

func (p *PollService) AddAnswer(ctx context.Context, pollID, userID int, answer []domain.Answer) error {
	return p.repo.AddAnswer(ctx, uint64(pollID), uint64(userID), answer)
}
