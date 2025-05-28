package domain

import (
	"html"
)

//easyjson:json
type PinData struct {
	FlowID         uint64 `json:"flow_id,omitempty"`
	Header         string `json:"header,omitempty"`
	AuthorID       uint64 `json:"author_id,omitempty"`
	AuthorUsername string `json:"author_username"`
	Description    string `json:"description,omitempty"`
	MediaURL       string `json:"media_url,omitempty"`
	IsPrivate      bool   `json:"is_private"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
	IsLiked        bool   `json:"is_liked"`
	IsNSFW         bool   `json:"is_nsfw"`
	LikeCount      int    `json:"like_count"`
	Width          int    `json:"width,omitempty"`
	Height         int    `json:"height,omitempty"`
}

func (p *PinData) Escape() {
	p.Header = html.EscapeString(p.Header)
	p.Description = html.EscapeString(p.Description)
}

func EscapeFlows(flows []PinData) {
	for i := range flows {
		flows[i].Escape()
	}
}
