package domain

import "time"

type Message struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	IsRead    bool      `json:"is_read"`
	Sender    string    `json:"sender"`
}

type Chat struct {
	ChatID       uint      `json:"chat_id"`
	Username     string    `json:"username"`
	Avatar       string    `json:"avatar"`
	PublicName   string    `json:"public_name"`
	MessageCount uint      `json:"message_count,omitempty"`
	Messages     []Message `json:"messages"` // only last 50 messages
}

type Contact struct {
	Username   string `json:"username"`
	PublicName string `json:"public_name"`
	Avatar     string `json:"avatar"`
}
