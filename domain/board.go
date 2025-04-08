package domain

import (
	"errors"
	"time"
)

type Board struct {
	Id        int       `json:"id"`
	AuthorID  int       `json:"author_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	IsPrivate bool      `json:"is_private"`
	FlowCount int       `json:"flow_count"`
	Preview   []PinData `json:"preview,omitempty"`
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

func (b Board) ValidateBoard() error {
	if len(b.Name) == 0 {
		return ErrNoBoardName
	}

	if b.AuthorID <= 0 {
		return ErrValidation
	}

	return nil
}
