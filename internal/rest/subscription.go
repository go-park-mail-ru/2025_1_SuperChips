package rest

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
)

type SubscriptionService interface {
	GetUserFollowers(ctx context.Context, id, page, size int) ([]domain.PublicUser, error)
	GetUserFollowing(ctx context.Context, id, page, size int) ([]domain.PublicUser, error)
	CreateSubscription(ctx context.Context, username, targetUsername string, currentID int) error
	DeleteSubscription(ctx context.Context, targetUsername string, currentID int) error
}

const SubscriptionType = "subscription"

//easyjson:json
type SubscriptionData struct {
	TargetUsername string `json:"target_user"`
}

type SubscriptionHandler struct {
	ContextExpiration   time.Duration
	SubscriptionService SubscriptionService
	NotificationChan    chan<- domain.WebMessage
}

// GetUserFollowers godoc
//	@Summary		Get user's followers
//	@Description	Returns a pageSized number of user's followers
//	@Produce		json
//	@Param			page	path	int							true	"requested page"	example("?page=3")
//	@Param			page	path	int							true	"requested size"	example("?size=15")
//	@Success		200		string	serverResponse.Data			"OK"
//	@Failure		404		string	serverResponse.Description	"page not found"
//	@Failure		400		string	serverResponse.Description	"bad request"
//	@Router			/api/v1/profile/followers [get]
func (h *SubscriptionHandler) GetUserFollowers(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	ctx, cancel := context.WithTimeout(context.Background(), h.ContextExpiration)
	defer cancel()

	page, size, err := getQueryPagination(w, r)
	if err != nil {
		return
	}

	followers, err := h.SubscriptionService.GetUserFollowers(ctx, claims.UserID, page, size)
	if err != nil {
		handleSubscriptionError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "OK",
		Data:        followers,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

// GetUserFollowers godoc
//	@Summary		Get user's subscriptions (or who they follow, in other words)
//	@Description	Returns a pageSized number of user's subscriptions
//	@Produce		json
//	@Param			page	path	int							true	"requested page"	example("?page=3")
//	@Param			page	path	int							true	"requested size"	example("?size=15")
//	@Success		200		string	serverResponse.Data			"OK"
//	@Failure		404		string	serverResponse.Description	"page not found"
//	@Failure		400		string	serverResponse.Description	"bad request"
//	@Router			/api/v1/profile/following [get]
func (h *SubscriptionHandler) GetUserFollowing(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	ctx, cancel := context.WithTimeout(context.Background(), h.ContextExpiration)
	defer cancel()

	page, size, err := getQueryPagination(w, r)
	if err != nil {
		return
	}

	following, err := h.SubscriptionService.GetUserFollowing(ctx, claims.UserID, page, size)
	if err != nil {
		handleSubscriptionError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "OK",
		Data:        following,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

// AddSubscription godoc
//	@Summary		Subscribe to target user
//	@Description	Tries to subscribe the user to the target user
//	@Accept			json
//	@Produce		json
//	@Param			target_user	body	string		true	"target user's username"	example("cool_guy")
//	@Success		200			string	Description	"OK"
//	@Failure		400			string	Description	"Bad Request"
//	@Failure		403			string	Description	"Unauthorized"
//	@Failure		500			string	Description	"Internal server error"
//	@Router			/api/v1/subscription [post]
func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	var subData SubscriptionData

	if err := DecodeData(w, r.Body, &subData); err != nil {
		return
	}

	if len(subData.TargetUsername) <= 2 {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.ContextExpiration)
	defer cancel()

	if err := h.SubscriptionService.CreateSubscription(ctx, claims.Username, subData.TargetUsername, claims.UserID); err != nil {
		log.Printf("create sub err: %v", err)
		handleSubscriptionError(w, err)
		return
	}

	// send notification
	if claims.Username != subData.TargetUsername {
		h.NotificationChan <- domain.WebMessage{
			Type: NotificationType,
			Content: domain.Notification{
				Type:             SubscriptionType,
				CreatedAt:        time.Now(),
				SenderUsername:   claims.Username,
				ReceiverUsername: subData.TargetUsername,
				AdditionalData:   nil,
			},
		}
	}

	resp := ServerResponse{
		Description: "Created",
	}

	ServerGenerateJSONResponse(w, resp, http.StatusCreated)
}

// DeleteSubscription godoc
//	@Summary		Unsubscribe from target user
//	@Description	Tries to unsubscribe the user from the target user
//	@Accept			json
//	@Produce		json
//	@Param			target_user	body	string		true	"target user's username"	example("cool_guy")
//	@Success		200			string	Description	"OK"
//	@Failure		400			string	Description	"Bad Request"
//	@Failure		403			string	Description	"Unauthorized"
//	@Failure		500			string	Description	"Internal server error"
//	@Router			/api/v1/subscription [delete]
func (h *SubscriptionHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	var subData SubscriptionData

	if err := DecodeData(w, r.Body, &subData); err != nil {
		return
	}

	if len(subData.TargetUsername) <= 2 {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.ContextExpiration)
	defer cancel()

	if err := h.SubscriptionService.DeleteSubscription(ctx, subData.TargetUsername, claims.UserID); err != nil {
		log.Printf("delete sub err: %v", err)
		handleSubscriptionError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "OK",
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

func getQueryPagination(w http.ResponseWriter, r *http.Request) (int, int, error) {
	page := r.URL.Query().Get("page")
	if page == "" {
		HttpErrorToJson(w, "page is not specified", http.StatusBadRequest)
		return 0, 0, fmt.Errorf("page is not specified")
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		HttpErrorToJson(w, "couldn't parse page into number", http.StatusBadRequest)
		return 0, 0, err
	}

	if pageInt <= 0 {
		HttpErrorToJson(w, "page cannot be less than or equal to zero", http.StatusBadRequest)
		return 0, 0, fmt.Errorf("page cannot be less than or equal to zero")
	}

	pageSize := r.URL.Query().Get("size")
	if pageSize == "" {
		HttpErrorToJson(w, "size is not specified", http.StatusBadRequest)
		return 0, 0, fmt.Errorf("size is not specified")
	}

	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		HttpErrorToJson(w, "couldn't parse size into number", http.StatusBadRequest)
		return 0, 0, err
	}

	if pageSizeInt <= 0 || pageSizeInt > 30 {
		HttpErrorToJson(w, "page size must be between 1 and 30", http.StatusBadRequest)
		return 0, 0, fmt.Errorf("page size must be between 1 and 30")
	}

	return pageInt, pageSizeInt, nil
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
