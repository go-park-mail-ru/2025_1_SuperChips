package domain

import (
	"html"
	"time"
)

//easyjson:json
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

//easyjson:json
type Chat struct {
	ChatID           uint      `json:"chat_id"`
	Username         string    `json:"username"`
	Avatar           string    `json:"avatar"`
	PublicName       string    `json:"public_name"`
	IsExternalAvatar bool      `json:"-"`
	MessageCount     uint      `json:"message_count,omitempty"`
	LastMessage      *Message  `json:"last_message,omitempty"`
	Messages         []Message `json:"messages,omitempty"`
}

//easyjson:json
type Contact struct {
	Username         string `json:"username"`
	PublicName       string `json:"public_name"`
	Avatar           string `json:"avatar"`
	IsExternalAvatar bool   `json:"-"`
}

func (m *Message) Escape() {
	m.Content = html.EscapeString(m.Content)
	m.Sender = html.EscapeString(m.Sender)
	m.Recipient = html.EscapeString(m.Recipient)
}

func (c *Chat) Escape() {
	c.Username = html.EscapeString(c.Username)
	c.PublicName = html.EscapeString(c.PublicName)
	
	if !c.IsExternalAvatar {
		c.Avatar = html.EscapeString(c.Avatar)
	}
	
	if c.LastMessage != nil {
		c.LastMessage.Escape()
	}
	
	for i := range c.Messages {
		c.Messages[i].Escape()
	}
}

func (c *Contact) Escape() {
	c.Username = html.EscapeString(c.Username)
	c.PublicName = html.EscapeString(c.PublicName)
	
	if !c.IsExternalAvatar {
		c.Avatar = html.EscapeString(c.Avatar)
	}
}
