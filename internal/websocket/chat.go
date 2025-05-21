package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/gorilla/websocket"
)

var (
	ErrTargetNotFound  = errors.New("target user not found")
	ErrDeliveryFailure = errors.New("couldn't send message to target user")
)

type ChatRepository interface {
	GetNewMessages(ctx context.Context, username string, offset time.Time) ([]domain.Message, error)
	AddMessage(ctx context.Context, message domain.Message) error
	MarkRead(ctx context.Context, messageID, chatID int) error
}

type Hub struct {
	connect          sync.Map
	currentOffset    time.Time
	chatRepo         ChatRepository
	notificationRepo NotificationRepository
}

func CreateHub(chatRepo ChatRepository, notificationRepo NotificationRepository) *Hub {
	return &Hub{
		connect:          sync.Map{},
		currentOffset:    time.Now().UTC(),
		chatRepo:         chatRepo,
		notificationRepo: notificationRepo,
	}
}

func (h *Hub) AddClient(username string, client *websocket.Conn) {
	h.connect.Store(username, client)

	client.SetCloseHandler(func(code int, text string) error {
		h.connect.Delete(username)
		return nil
	})
}

func (h *Hub) MarkRead(ctx context.Context, messageID, chatID int, targetUsername, senderUsername string) error {
	if err := h.chatRepo.MarkRead(ctx, messageID, chatID); err != nil {
		return fmt.Errorf("couldn't mark messages as read: %v", err)
	}

	found := false
	var targetConn *websocket.Conn

	h.connect.Range(func(key, value any) bool {
		username := key.(string)
		conn := value.(*websocket.Conn)
		if username == targetUsername {
			found = true
			targetConn = conn
			return false
		}

		return true
	})

	// user is offline, so end here
	if !found {
		return nil
	}

	type MessageRead struct {
		Description string `json:"description"`
		MessageID   int    `json:"message_id"`
		IsRead      bool   `json:"is_read"`
		Sender      string `json:"sender"`
		ChatID      int    `json:"chat_id"`
	}

	message := MessageRead{
		Description: "mark_read",
		MessageID:   messageID,
		IsRead:      true,
		Sender:      senderUsername,
		ChatID:      chatID,
	}

	targetConn.WriteJSON(message)

	return nil
}

func (h *Hub) SendMessage(ctx context.Context, msg domain.WebMessage, senderUsername string) error {
	log.Println("sending message for some reason")
	
	found := false
	var targetConn *websocket.Conn

	var message domain.Message

	byteData, err := json.Marshal(msg.Content)
	if err != nil {
		log.Println("notification: error marshalling message")
		return fmt.Errorf("notification: error marshalling message")
	}

	if err := json.Unmarshal(byteData, &message); err != nil {
		log.Println("notification: error unmarshalling message")
		return fmt.Errorf("notification: error unmarshalling message")
	}

	message.Sender = senderUsername

	h.connect.Range(func(key, value any) bool {
		username := key.(string)
		conn := value.(*websocket.Conn)
		if message.Recipient == username {
			found = true
			targetConn = conn
			// range realization neat:
			// false stops the iteration
			return false
		}

		return true
	})

	if err := h.chatRepo.AddMessage(ctx, message); err != nil {
		log.Printf("error while adding message to db: %v", err)
		return err
	}

	// target user offline
	if !found {
		log.Println("user not online")
		return ErrTargetNotFound
	}

	msg = domain.WebMessage{
		Type: "message",
		Content: message,
	}

	err = targetConn.WriteJSON(msg)
	if err != nil {
		log.Printf("delivery failure: %v", err)
		targetConn.Close()
		h.connect.Delete(message.Recipient)
		return ErrDeliveryFailure
	}

	return nil
}

func (h *Hub) Run(ctx context.Context) {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			h.connect.Range(func(key, value any) bool {
				username := key.(string)
				conn := value.(*websocket.Conn)
				messages, err := h.chatRepo.GetNewMessages(ctx, username, h.currentOffset)
				if err != nil {
					log.Printf("error getting new messages: %v", err)
				}
				for _, message := range messages {
					webMsg := domain.WebMessage{
						Type: "message",
						Content: message,
					}

					err := conn.WriteJSON(webMsg)
					if err != nil {
						continue
					}
				}
				return true
			})
			h.currentOffset = h.currentOffset.Add(5 * time.Second)
		case <-ctx.Done():
			return
		}
	}
}
