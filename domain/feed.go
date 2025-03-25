package domain

type PinData struct {
	FlowID      uint64 `json:"-"`
	Header      string `json:"header,omitempty"`
	Description string `json:"description,omitempty"`
	MediaURL    string `json:"image,omitempty"`
	AuthorID    uint64 `json:"author_id"`
	IsPrivate   bool   `json:"-"`
	Created_at  string `json:"-"`
	Updated_at  string `json:"-"`
}
