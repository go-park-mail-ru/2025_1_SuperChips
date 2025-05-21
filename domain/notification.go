package domain

import "time"

type WebMessage struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

type Notification struct {
	ID                   uint        `json:"id"`
	Type                 string      `json:"type"`
	CreatedAt            time.Time   `json:"created_at"`
	SenderUsername       string      `json:"sender"`
	SenderAvatar         string      `json:"sender_avatar"`
	SenderExternalAvatar bool        `json:"-"`
	ReceiverUsername     string      `json:"receiver"`
	IsRead               bool        `json:"is_read"`
	AdditionalData       interface{} `json:"additional_data"`
}

// useful data to know when you create a notification
type NewNotificationData struct {
	ID               uint      // id
	Avatar           string    // sender avatar
	isExternalAvatar string    // whether sender's avatar is from an external source
	Timestamp        time.Time // timestamp)
}
