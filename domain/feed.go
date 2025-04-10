package domain

type PinData struct {
	FlowID      uint64 `json:"flow_id,omitempty"`
	Header      string `json:"header,omitempty"`
	AuthorID    uint64 `json:"author_id"`
	Description string `json:"description,omitempty"`
	MediaURL    string `json:"media_url,omitempty"`
	IsPrivate   bool   `json:"is_private"`
	Created_at  string `json:"-"`
	IsLiked     bool   `json:"is_liked"`
	LikeCount   int    `json:"like_count"`
}
