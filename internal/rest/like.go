package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
)

type LikeService interface {
	LikeFlow(ctx context.Context, pinID, userID int) (string, error)
}

type LikeHandler struct {
	LikeService    LikeService
	ContextTimeout time.Duration
}

func (h *LikeHandler) LikeFlow(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	var likePin domain.Like

	if err := DecodeData(w, r.Body, &likePin); err != nil {
		return
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, h.ContextTimeout)
	defer cancel()

	action, err := h.LikeService.LikeFlow(ctx, likePin.PinID, claims.UserID)
	if err != nil {
		handleLikeError(w, err)
		return
	}

	type likeAction struct {
		Action string `json:"action"`
	}

	liked := likeAction{
		Action: action,
	}

	resp := ServerResponse{
		Description: "OK",
		Data: liked,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

func handleLikeError(w http.ResponseWriter, err error) {
	switch {
	default:
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
