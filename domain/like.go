package domain

import "errors"

var (
	ErrForbidden = errors.New("forbidden")
)

//easyjson:json
type Like struct {
	PinID int `json:"pin_id,omitempty"`
}

