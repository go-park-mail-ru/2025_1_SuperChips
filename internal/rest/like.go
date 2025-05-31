package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
)

const LikeType = "like"

type LikeService interface {
	LikeFlow(ctx context.Context, pinID, userID int) (string, string, error)
}

type LikeHandler struct {
	LikeService      LikeService
	ContextTimeout   time.Duration
	NotificationChan chan<- domain.WebMessage
}

// LikeFlow godoc
//	@Summary		Leave a like on a flow
//	@Description	Leaves a like on a flow or deletes the like 
//	@Accept			json
//	@Produce		json
//	@Param			pin_id	body	integer		true	"flow id"	example(456)
//	@Success		200		string	Description	"OK"
//	@Failure		400		string	Description	"Bad Request"
//	@Failure		404		string	Description	"Not Found"
//	@Failure		500		string	Description	"Internal server error"
//	@Router			/api/v1/like [post]
func (h *LikeHandler) LikeFlow(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	var likePin domain.Like

	if err := DecodeData(w, r.Body, &likePin); err != nil {
		return
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, h.ContextTimeout)
	defer cancel()

	action, authorUsername, err := h.LikeService.LikeFlow(ctx, likePin.PinID, claims.UserID)
	if err != nil {
		handleLikeError(w, err)
		return
	}

	if claims.Username != authorUsername && action == "liked" {
		h.NotificationChan <- domain.WebMessage{
			Type: NotificationType,
			Content: domain.Notification{
				Type:             LikeType,
				CreatedAt:        time.Now(),
				SenderUsername:   claims.Username,
				ReceiverUsername: authorUsername,
				AdditionalData:   likePin,
			},
		}
	}

	type likeAction struct {
		Action string `json:"action"`
	}

	liked := likeAction{
		Action: action,
	}

	resp := ServerResponse{
		Description: "OK",
		Data:        liked,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

func handleLikeError(w http.ResponseWriter, err error) {
	switch {
	default:
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
