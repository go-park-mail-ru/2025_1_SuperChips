package domain

import "time"

type Message struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	IsRead    bool      `json:"is_read"`
	Sender    string    `json:"sender"`
}
