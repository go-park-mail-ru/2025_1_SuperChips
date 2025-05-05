package domain

import "time"

type Message struct {
	MessageID uint      `json:"message_id"`
	Content   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	IsRead    bool      `json:"is_read"`
	Sender    string    `json:"sender"`
	Recipient string    `json:"recipient"`
	ChatID    uint64    `json:"chat_id"`
	Sent      bool      `json:"-"`
}

type Chat struct {
	ChatID           uint      `json:"chat_id"`
	Username         string    `json:"username"`
	Avatar           string    `json:"avatar"`
	PublicName       string    `json:"public_name"`
	IsExternalAvatar bool      `json:"-"`
	MessageCount     uint      `json:"message_count,omitempty"`
	LastMessage      *Message  `json:"last_message,omitempty"`
	Messages         []Message `json:"messages,omitempty"` // only last 200 messages
}

type Contact struct {
	Username         string `json:"username"`
	PublicName       string `json:"public_name"`
	Avatar           string `json:"avatar"`
	IsExternalAvatar bool   `json:"-"`
}
