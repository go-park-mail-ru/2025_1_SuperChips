package domain

type PinData struct {
	FlowID         uint64 `json:"flow_id,omitempty"`
	Header         string `json:"header,omitempty"`
	AuthorID       uint64 `json:"author_id,omitempty"`
	AuthorUsername string `json:"author_username"`
	Description    string `json:"description,omitempty"`
	MediaURL       string `json:"media_url,omitempty"`
	IsPrivate      bool   `json:"is_private"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	IsLiked        bool   `json:"is_liked"`
	LikeCount      int    `json:"like_count"`
}
