package rest

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	gen "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/poll"
	"google.golang.org/grpc/status"
)

type PollHandler struct {
	Usecase        gen.PollServiceClient
	ContextTimeout time.Duration
}

func (h *PollHandler) AddAnswer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.ContextTimeout)
	defer cancel()

	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	type answers struct {
		Answers []domain.Answer `json:"answers"`
	}

	var gotAnswers answers

	if err := DecodeData(w, r.Body, &gotAnswers); err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	_, err = h.Usecase.AddAnswer(ctx, &gen.AddAnswerRequest{
		PollId: int64(idInt),
		UserId: int64(claims.UserID),
		Answers: formatAnswers(gotAnswers.Answers),
	})
	if err != nil {
		handleGRPCPollError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "OK",
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

func (h *PollHandler) GetPolls(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), h.ContextTimeout)
	defer cancel()

	grpcResp, err := h.Usecase.GetAllPolls(ctx, &gen.Empty{})
	if err != nil {
		handleGRPCPollError(w, err)
		return
	}

	var polls []domain.Poll
	for i := range grpcResp.Polls {
		polls = append(polls, domain.Poll{
			ID: uint64(grpcResp.Polls[i].Id),
			Header: grpcResp.Polls[i].Header,
			Delay: int(grpcResp.Polls[i].Delay),
			Screen: grpcResp.Polls[i].Screen,
			Questions: formatQuestions(grpcResp.Polls[i].Questions),
		})
	}

	resp := ServerResponse{
		Description: "OK",
		Data: polls,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

func formatQuestions(grpcQuestions []*gen.Question) []domain.Question {
	var questions []domain.Question
	for i := range grpcQuestions {
		questions = append(questions, domain.Question{
			ID: uint64(grpcQuestions[i].Id),
			Text: grpcQuestions[i].Text,
			Order: grpcQuestions[i].Order,
			Type: grpcQuestions[i].Type,
		})
	}

	return questions
}

func formatAnswers(answers []domain.Answer) []*gen.Answer {
	var grpcAnswers []*gen.Answer
	for i := range answers {
		grpcAnswers = append(grpcAnswers, &gen.Answer{
			Type: answers[i].Type,
			Content: answers[i].Content,
			QuestionId: int64(answers[i].QuestionID),
		})
	}

	return grpcAnswers
}

func handleGRPCPollError(w http.ResponseWriter, err error) {
	newErr := status.Convert(err)
	switch newErr.Code() {
	default:
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
