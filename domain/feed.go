package domain

type PinData struct {
	FlowID      uint64 `json:"-"`
	Header      string `json:"header,omitempty"`
	Description string `json:"description,omitempty"`
	MediaURL    string `json:"media_url"`
	AuthorID    uint64 `json:"author_id"`
	IsPrivate   bool   `json:"is_private"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	LikeCount   uint64 `json:"like_count"`
}
