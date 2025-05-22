package boardinv

import "errors"

var (
	ErrLinkNotFound         = errors.New("link to the board does not exist")
	ErrNonExistentUsernames = errors.New("some usernames do not exist")
	ErrInconsistentDataInDB = errors.New("inconsistent data in DB")
)
