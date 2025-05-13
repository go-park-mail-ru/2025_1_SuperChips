package domain

import (
	"errors"
	"html"
	"time"
)

type Gradient struct {
	First  string `json:"first_color"`
	Second string `json:"second_color"`
	Third  string `json:"third_color"`
	Fourth string `json:"fourth_color"`
}

//easyjson:json
type Board struct {
	ID             int       `json:"id"`
	AuthorID       int       `json:"author_id"`
	AuthorUsername string    `json:"author_username,omitempty"`
	Name           string    `json:"name"`
	CreatedAt      time.Time `json:"-"`
	IsPrivate      bool      `json:"is_private"`
	FlowCount      int       `json:"flow_count"`
	Preview        []PinData `json:"preview,omitempty"`
	Gradient       []string  `json:"gradient,omitempty"`
}

//easyjson:json
type BoardRequest struct {
	FlowID int `json:"flow_id,omitempty"`
}

//easyjson:json
type UpdateData struct {
	Name      string `json:"name"`
	IsPrivate bool   `json:"is_private"`
}

func (b *Board) Escape() {
	b.Name = html.EscapeString(b.Name)
}

func EscapeBoards(boards []Board) {
	for i := range boards {
		boards[i].Escape()
		for j := range boards[i].Preview {
			boards[i].Preview[j].Escape()
		}
	}
}

var (
	ErrNoBoardName        = errors.New("board must have a name")
	ErrBoardAlreadyExists = errors.New("a board with that name already exists in your account")
)
