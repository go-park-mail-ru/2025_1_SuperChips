package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	connect       sync.Map
	currentOffset time.Time
	repo          ChatRepository
}

func CreateHub(repo ChatRepository) *Hub {
	return &Hub{
		connect:       sync.Map{},
		currentOffset: time.Now(),
		repo:          repo,
	}
}

func (h *Hub) AddClient(username string, client *websocket.Conn) {
	h.connect.Store(client, username)

	client.SetCloseHandler(func(code int, text string) error {
		h.connect.Delete(client)
		return nil
	})
}

func (h *Hub) MarkRead(ctx context.Context, messageID, chatID int, targetUsername, senderUsername string) error {
	if err := h.repo.MarkRead(ctx, messageID, chatID); err != nil {
		return fmt.Errorf("couldn't mark messages as read: %v", err)
	}

	found := false
	var targetConn *websocket.Conn

	h.connect.Range(func(key, value any) bool {
		conn := key.(*websocket.Conn)
		username := value.(string)
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

func (h *Hub) Send(ctx context.Context, message domain.Message, targetUsername string) error {
	found := false
	var targetConn *websocket.Conn

	h.connect.Range(func(key, value any) bool {
		conn := key.(*websocket.Conn)
		username := value.(string)
		if targetUsername == username {
			found = true
			targetConn = conn
			// range realization neat:
			// false stops the iteration
			return false
		}

		return true
	})

	message.Recipient = targetUsername

	if err := h.repo.AddMessage(ctx, message); err != nil {
		log.Printf("error while adding message to db: %v", err)
		return err
	}

	// target user offline
	if !found {
		log.Println("user not online")
		return ErrTargetNotFound
	}

	err := targetConn.WriteJSON(message)
	if err != nil {
		log.Printf("delivery failure: %v", err)
		targetConn.Close()
		h.connect.Delete(targetConn)
		return ErrDeliveryFailure
	}

	return nil
}

type ChatConn struct {
	*websocket.Conn
}

func (c *ChatConn) ReadJSON(v interface{}) error {
    _, r, err := c.NextReader()
    if err != nil {
        return err
    }

    buf := new(bytes.Buffer)
    if _, err := buf.ReadFrom(r); err != nil {
        return err
    }

    // Декодируем из буфера
    if err := json.Unmarshal(buf.Bytes(), v); err != nil {
        // Если нужно — логируем «сырое» тело
        fmt.Printf("resp body: %s\n", buf.String())
        return err
    }

    return nil
}

// func (h *Hub) Run(ctx context.Context) {
// 	t := time.NewTicker(5 * time.Second)
// 	defer t.Stop()

// 	for {
// 		select {
// 		case <-t.C:
// 			h.connect.Range(func(key, value any) bool {
// 				connect := key.(*websocket.Conn)
// 				username := value.(string)
// 				//для каждлого клиента читаем новые изменения
// 				//тут может быть что угодно - сообщения, тексты, тд
// 				messages, err := h.repo.GetNewMessages(ctx, username, h.currentOffset)
// 				if err != nil {
// 					log.Printf("error getting new messages: %v", err)
// 				}
// 				for _, message := range messages {
// 					err := connect.WriteJSON(message)
// 					if err != nil {
// 						continue
// 					}
// 				}
// 				return true
// 			})
// 			h.currentOffset = h.currentOffset.Add(5 * time.Second)
// 		case <-ctx.Done():
// 			return
// 		}
// 	}
// }
