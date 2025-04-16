package domain

import (
	"errors"
	"html"
	"time"
)

type Board struct {
	ID             int       `json:"id"`
	AuthorID       int       `json:"author_id"`
	AuthorUsername string    `json:"author_username,omitempty"`
	Name           string    `json:"name"`
	CreatedAt      time.Time `json:"-"`
	IsPrivate      bool      `json:"is_private"`
	FlowCount      int       `json:"flow_count"`
	Preview        []PinData `json:"preview,omitempty"`
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

type BoardPost struct {
	BoardID int
	FlowID  int
	SavedAt time.Time
}

var (
	ErrNoBoardName        = errors.New("board must have a name")
	ErrBoardAlreadyExists = errors.New("a board with that name already exists in your account")
)
