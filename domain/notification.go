package domain

type WebMessage struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

type Notification struct {
	ID                   string      `json:"id"`
	Type                 string      `json:"type"`
	CreatedAt            string      `json:"created_at"`
	SenderUsername       string      `json:"sender"`
	SenderAvatar         string      `json:"sender_avatar"`
	SenderExternalAvatar bool        `json:"-"`
	ReceiverUsername     string      `json:"receiver"`
	IsRead               bool        `json:"is_read"`
	AdditionalData       interface{} `json:"additional_data"`
}
