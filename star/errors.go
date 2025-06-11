package star

import "errors"

var (
	ErrPinNotFound          = errors.New("Pin not found")
	ErrInconsistentDataInDB = errors.New("Inconsistent data in DB")
	ErrAlreadyStar          = errors.New("Pin is already star")
	ErrNoFreeStarSlots      = errors.New("No free star slots")
	ErrPinIsNotStar         = errors.New("Pin is not star")
)
