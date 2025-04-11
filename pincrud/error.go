package pincrud

import "errors"

var (
	ErrForbidden        = errors.New("access to private pin is forbidden")
	ErrPinNotFound      = errors.New("no pin with given id")
	ErrUntracked        = errors.New("untracked service error")
	ErrNoFieldsToUpdate = errors.New("no fields to update")
	ErrInvalidImageExt  = errors.New("invalid image extension")
)
