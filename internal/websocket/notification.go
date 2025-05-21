package websocket

import (
	"context"
	"fmt"
	"log"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/gorilla/websocket"
)

type NotificationRepository interface {
	GetNewNotifications(ctx context.Context, userID uint64) ([]domain.Notification, error)
	AddNotification(ctx context.Context, notification domain.Notification) error
}


func (h *Hub) SendNotification(ctx context.Context, webMsg domain.WebMessage) error {
	found := false
	var targetConn *websocket.Conn

	notification, ok := webMsg.Content.(domain.Notification)
	if !ok {
		log.Println("error casting message content")
		return fmt.Errorf("error casting message content")
	}

	h.connect.Range(func(key, value any) bool {
		username := key.(string)
		conn := value.(*websocket.Conn)
		if notification.ReceiverUsername == username {
			found = true
			targetConn = conn
			return false
		}

		return true
	})

	if err := h.notificationRepo.AddNotification(ctx, notification); err != nil {
		log.Printf("couldn't add notification to db: %v", err)
		return err
	}

	if !found {
		log.Println("user not online")
	}

	err := targetConn.WriteJSON(notification)
	if err != nil {
		log.Printf("delivery failure: %v", err)
		targetConn.Close()
		h.connect.Delete(notification.ReceiverUsername)
		return ErrDeliveryFailure
	}

	return nil
}