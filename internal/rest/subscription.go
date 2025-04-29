package rest

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
)

type SubscriptionService interface {
	GetUserFollowers(ctx context.Context, id int) ([]domain.PublicUser, error)
	GetUserFollowing(ctx context.Context, id int) ([]domain.PublicUser, error)
	CreateSubscription(ctx context.Context, targetUsername string, currentID int) error
	DeleteSubscription(ctx context.Context, targetUsername string, currentID int) error
}

type SubscriptionData struct {
	TargetUsername string `json:"target_user"`
}

type SubscriptionHandler struct {
	ContextExpiration time.Duration
	SubscriptionService SubscriptionService
}

// GET api/v1/profile/followers
func (h *SubscriptionHandler) GetUserFollowers(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	ctx, cancel := context.WithTimeout(context.Background(), h.ContextExpiration)
	defer cancel()

	followers, err := h.SubscriptionService.GetUserFollowers(ctx, claims.UserID)
	if err != nil {
		println(err.Error())
		handleSubscriptionError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "OK",
		Data: followers,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

// GET api/v1/profile/following
func (h *SubscriptionHandler) GetUserFollowing(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	ctx, cancel := context.WithTimeout(context.Background(), h.ContextExpiration)
	defer cancel()

	following, err := h.SubscriptionService.GetUserFollowing(ctx, claims.UserID)
	if err != nil {
		println(err.Error())
		handleSubscriptionError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "OK",
		Data: following,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

// POST api/v1/subscription
func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
	
	var subData SubscriptionData
	
	if err := DecodeData(w, r.Body, &subData); err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.ContextExpiration)
	defer cancel()

	if err := h.SubscriptionService.CreateSubscription(ctx, subData.TargetUsername, claims.UserID); err != nil {
		println(err.Error())
		handleSubscriptionError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "Created",
	}

	ServerGenerateJSONResponse(w, resp, http.StatusCreated)
}

// DELETE api/v1/subscription
func (h *SubscriptionHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
	
	var subData SubscriptionData
	
	if err := DecodeData(w, r.Body, &subData); err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.ContextExpiration)
	defer cancel()

	if err := h.SubscriptionService.DeleteSubscription(ctx, subData.TargetUsername, claims.UserID); err != nil {
		println(err.Error())
		handleSubscriptionError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "OK",
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

func handleSubscriptionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrConflict):
		HttpErrorToJson(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	case errors.Is(err, domain.ErrNotFound):
		HttpErrorToJson(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	case errors.Is(err, domain.ErrValidation):
		HttpErrorToJson(w, "validation failed", http.StatusBadRequest)
		return
	default:
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
