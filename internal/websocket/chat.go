package websocket

import (
	"bytes"
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

const (
    pongWait   = 60 * time.Second
    pingPeriod = (pongWait * 9) / 10
)

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

    client.SetReadDeadline(time.Now().Add(pongWait))
    client.SetPongHandler(func(string) error {
        client.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    go func() {
        ticker := time.NewTicker(pingPeriod)
        defer ticker.Stop()
        for {
            if err := client.WriteMessage(websocket.PingMessage, nil); err != nil {
                client.Close()
                h.connect.Delete(client)
                return
            }
            time.Sleep(pingPeriod)
        }
    }()

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

