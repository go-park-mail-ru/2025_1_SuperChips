package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/gorilla/websocket"
)

const NotificationType = "notification"

type NotificationRepository interface {
	GetNewNotifications(ctx context.Context, userID uint64) ([]domain.Notification, error)
	AddNotification(ctx context.Context, notification domain.Notification) (domain.NewNotificationData, error)
	DeleteNotification(ctx context.Context, id, usernameID uint64) error
}

func (h *Hub) SendNotification(ctx context.Context, webMsg domain.WebMessage) error {
	var targetConn *websocket.Conn

	var notification domain.Notification

	byteData, err := json.Marshal(webMsg.Content)
	if err != nil {
		log.Printf("notification: error marshalling message %v", err)
		return fmt.Errorf("notification: error marshalling message: %v", err)
	}

	if err := json.Unmarshal(byteData, &notification); err != nil {
		log.Printf("notification: error unmarshalling message: %v", err)
		return fmt.Errorf("notification: error unmarshalling message: %v", err)
	}

	conn, found := h.connect.Load(notification.ReceiverUsername)
	targetConn, ok := conn.(*websocket.Conn)
	if !ok {
		return ErrDeliveryFailure
	}

	newData, err := h.notificationRepo.AddNotification(ctx, notification)
	if err != nil {
		log.Printf("couldn't add notification to db: %v", err)
		return err
	}

	notification.ID = newData.ID
	notification.CreatedAt = newData.Timestamp
	notification.SenderAvatar = newData.Avatar

	if !found {
		return ErrTargetNotFound
	}

	webMsg = domain.WebMessage{
		Type: NotificationType,
		Content: notification,
	}

	err = targetConn.WriteJSON(webMsg)
	if err != nil {
		log.Printf("delivery failure: %v", err)
		targetConn.Close()
		h.connect.Delete(notification.ReceiverUsername)
		return ErrDeliveryFailure
	}

	return nil
}

func (h *Hub) DeleteNotification(ctx context.Context, webMsg domain.WebMessage, usernameID uint64) error {
	byteData, err := json.Marshal(webMsg.Content)
	if err != nil {
		log.Printf("delete notification: error marshalling message: %v", err)
		return fmt.Errorf("delete notification: error marshalling message: %v", err)
	}

	type DeletionID struct {
		ID uint64 `json:"id"`
	}

	var delID DeletionID

	if err := json.Unmarshal(byteData, &delID); err != nil {
		log.Printf("delete notification: error unmarshalling message: %v", err)
		return fmt.Errorf("delete notification: error unmarshalling message: %v", err)
	}

	if err := h.notificationRepo.DeleteNotification(ctx, delID.ID, usernameID); err != nil {
		return err
	}

	return nil
}