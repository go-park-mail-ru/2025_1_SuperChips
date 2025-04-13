package domain

import "errors"

var (
	ErrForbidden = errors.New("forbidden")
)

type Like struct {
	PinID int `json:"pin_id,omitempty"`
}

