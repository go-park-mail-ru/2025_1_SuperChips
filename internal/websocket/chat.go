package websocket

import (
	"context"
	"errors"
	"fmt"
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
	GetNewMessages(ctx context.Context, userID int64, offset time.Time) ([]string, error)
	AddMessage(ctx context.Context, message domain.Message) error
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

func (h *Hub) AddClient(userID int64, client *websocket.Conn) {
	h.connect.Store(client, userID)

	go func() {
		for {
			_, _, err := client.NextReader()
			if err != nil {
				err = client.Close()
				if err != nil {
					return
				}
				return
			}
		}
	}()

	client.SetCloseHandler(func(code int, text string) error {
		h.connect.Delete(client)
		return nil
	})
}

func (h *Hub) Send(ctx context.Context, message domain.Message, targetUserID int64) error {
	found := false
	var targetConn *websocket.Conn

	h.connect.Range(func(key, value any) bool {
		conn := key.(*websocket.Conn)
		userID := value.(int64)
		if targetUserID == userID {
			found = true
			targetConn = conn
			// range realization neat:
			// false stops the iteration
			return false
		}

		return true
	})

	if !found {
		return ErrTargetNotFound
	}
	
	if err := h.repo.AddMessage(ctx, message); err != nil {
		return fmt.Errorf("error while adding message to db: %v", err)
	}

	err := targetConn.WriteJSON(message)
	if err != nil {
		targetConn.Close()
		h.connect.Delete(targetConn)
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
				connect := key.(*websocket.Conn)
				userID := value.(int64)
				//для каждлого клиента читаем новые изменения
				//тут может быть что угодно - сообщения, тексты, тд
				messages, _ := h.repo.GetNewMessages(ctx, userID, h.currentOffset)
				for _, message := range messages {
					err := connect.WriteJSON(message)
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
