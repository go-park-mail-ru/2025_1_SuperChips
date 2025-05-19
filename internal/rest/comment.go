package rest

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
)

type CommentService interface {
	GetComments(ctx context.Context, flowID, userID, page, size int) ([]domain.Comment, error)
	LikeComment(ctx context.Context, flowID, commentID, userID int) (string, error)
	AddComment(ctx context.Context, flowID, userID int, content string) error
	DeleteComment(ctx context.Context, commentID, userID int) error
}

type CommentHandler struct {
	Service           CommentService
	ContextExpiration time.Duration
}

func (h *CommentHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	userID := 0
	claims, ok := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
	if ok {
		userID = claims.UserID
	}

	page, size, err := getQueryPagination(w, r)
	if err != nil {
		return
	}

	flowIDStr := r.PathValue("flow_id")
	flowID, err := strconv.Atoi(flowIDStr)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.ContextExpiration)
	defer cancel()

	comments, err := h.Service.GetComments(ctx, flowID, userID, page, size)
	if err != nil {
		handleCommentError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "OK",
		Data: comments,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

func (h *CommentHandler) LikeComment(w http.ResponseWriter, r *http.Request) {
	flowIDStr := r.PathValue("flow_id")
	flowID, err := strconv.Atoi(flowIDStr)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	commentIDStr := r.PathValue("comment_id")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	ctx, cancel := context.WithTimeout(context.Background(), h.ContextExpiration)
	defer cancel()

	action, err := h.Service.LikeComment(ctx, flowID, commentID, claims.UserID)
	if err != nil {
		handleCommentError(w, err)
		return
	}

	type likeAction struct {
		Action string `json:"action"`
	}

	likeMsg := likeAction{
		Action: action,
	}

	resp := ServerResponse{
		Description: "OK",
		Data: likeMsg,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

func (h *CommentHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	var comment domain.Comment

	if err := DecodeData(w, r.Body, &comment); err != nil {
		return
	}

	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	flowIDStr := r.PathValue("flow_id")
	flowID, err := strconv.Atoi(flowIDStr)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := comment.Validate(); err != nil {
		HttpErrorToJson(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.ContextExpiration)
	defer cancel()

	if err := h.Service.AddComment(ctx, flowID, claims.UserID, comment.Content); err != nil {
		handleCommentError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "Created",
	}

	ServerGenerateJSONResponse(w, resp, http.StatusCreated)
}

func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	commentIDStr := r.PathValue("comment_id")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	ctx, cancel := context.WithTimeout(context.Background(), h.ContextExpiration)
	defer cancel()

	if err := h.Service.DeleteComment(ctx, commentID, claims.UserID); err != nil {
		handleCommentError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "Deleted",
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

func handleCommentError(w http.ResponseWriter, err error) {
	switch {
	default:
		HttpErrorToJson(w, err.Error(), http.StatusInternalServerError)
	}
}