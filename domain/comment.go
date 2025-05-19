package domain

import (
	"fmt"
	"html"
	"time"
)

//easyjson:json
type Comment struct {
	ID                     int       `json:"id"`
	FlowID                 int       `json:"flow_id"`
	AuthorID               int       `json:"-"`
	AuthorUsername         string    `json:"author_username"`
	AuthorAvatar           string    `json:"author_avatar"`
	AuthorIsExternalAvatar bool      `json:"-"`
	Content                string    `json:"content"`
	Timestamp              time.Time `json:"timestamp"`
	LikeCount              int       `json:"like_count"`
}

func (c *Comment) Validate() error {
	if c.Content == "" {
		return fmt.Errorf("content cannot be empty")
	}
	return nil
}

func (c *Comment) Escape() {
	c.AuthorUsername = html.EscapeString(c.AuthorUsername)
	c.Content = html.EscapeString(c.Content)
}
