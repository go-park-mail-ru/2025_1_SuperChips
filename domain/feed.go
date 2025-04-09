package domain

type PinData struct {
	FlowID      uint64 `json:"flow_id,omitempty"`
	Header      string `json:"header,omitempty"`
	AuthorID    uint64 `json:"author_id"`
	Description string `json:"description,omitempty"`
	MediaURL    string `json:"image,omitempty"`
	IsPrivate   bool   `json:"-"`
	Created_at  string `json:"-"`
	IsLiked     bool   `json:"is_liked"`
	LikeCount   int    `json:"like_count"`
}
