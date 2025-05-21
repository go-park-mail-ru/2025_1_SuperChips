package rest

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	chatWebsocket "github.com/go-park-mail-ru/2025_1_SuperChips/internal/websocket"
	"github.com/gorilla/websocket"
)

type NotificationService interface {
	GetNotifications(ctx context.Context, userID uint) ([]domain.Notification, error)
}

type NotificationHandler struct {
	NotificationService NotificationService
	ContextExpiration   time.Duration
}

func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.ContextExpiration)
	defer cancel()

	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	notifications, err := h.NotificationService.GetNotifications(ctx, uint(claims.UserID))
	if err != nil {
		handleNotificationError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "OK",
		Data:        notifications,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

func handleNotificationError(w http.ResponseWriter, err error) {
	switch err {
	case domain.ErrNotFound:
		http.Error(w, "not found", http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleNotification(ctx context.Context, conn *websocket.Conn,
	webMsg domain.WebMessage, claims *auth.Claims, hub *chatWebsocket.Hub) error {
	if err := hub.SendNotification(ctx, webMsg); err != nil {
		log.Printf("error sending notification: %v", err)
	}

	return nil
}
