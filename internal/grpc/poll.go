package grpc

import (
	"context"

	gen "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/poll"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type PollUsecase interface {
	GetAllPolls(ctx context.Context) ([]domain.Poll, error)
	AddAnswer(ctx context.Context, pollID, userID int, answers []domain.Answer) error
}

type GrpcPollHandler struct {
	gen.UnimplementedPollServiceServer
	usecase PollUsecase
}

func NewGrpcPollHandler(usecase PollUsecase) *GrpcPollHandler {
	return &GrpcPollHandler{
		usecase: usecase,
	}
}

func (p *GrpcPollHandler) GetAllPolls(ctx context.Context, in *gen.Empty) (*gen.GetAllPollsResponse, error) {
	polls, err := p.usecase.GetAllPolls(ctx)
	if err != nil {
		return nil, err
	}

	var grpcPolls []*gen.Poll
	for i := range polls {
		grpcPolls = append(grpcPolls, &gen.Poll{
			Id: int64(polls[i].ID),
			Header: polls[i].Header,
			Delay: int64(polls[i].Delay),
			Screen: polls[i].Screen,
			Questions: questionToGRPC(polls[i].Questions),
		})
	}

	return &gen.GetAllPollsResponse{
		Polls: grpcPolls,
	}, nil
}

func (p *GrpcPollHandler) AddAnswer(ctx context.Context, in *gen.AddAnswerRequest) (*gen.Empty, error) {
	var normalAnswers []domain.Answer
	for i := range in.Answers {
		normalAnswers = append(normalAnswers, domain.Answer{
			Type: in.Answers[i].Type,
			Content: in.Answers[i].Content,
			QuestionID: int(in.Answers[i].QuestionId),
		})
	}

	if err := p.usecase.AddAnswer(ctx, int(in.PollId), int(in.UserId), normalAnswers); err != nil {
		return nil, err
	}

	return &gen.Empty{}, nil
}

func questionToGRPC(questions []domain.Question) []*gen.Question {
	var grpcQuestions []*gen.Question
	for i := range questions {
		grpcQuestions = append(grpcQuestions, &gen.Question{
			Id: int64(questions[i].ID),
			Text: questions[i].Text,
			Order: questions[i].Order,
			Type: questions[i].Type,
		})
	}

	return grpcQuestions
}

