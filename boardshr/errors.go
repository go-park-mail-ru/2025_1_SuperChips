package boardshr

import "errors"

var (
	ErrLinkNotFound         = errors.New("link to the board does not exist")
	ErrNonExistentUsername = errors.New("some usernames do not exist")
	ErrInconsistentDataInDB = errors.New("inconsistent data in DB")
	ErrAlreadyEditor        = errors.New("user is already an editor of the board")
	ErrForbbiden            = errors.New("access is forbidden")
	ErrLinkExpired          = errors.New("link's time or usage limit has expired")
	ErrFailCoauthorInsert   = errors.New("failed to add a coauthor")
	ErrFailCoauthorDelete   = errors.New("failed to delete a coauthor")
	ErrAuthorRefuseEditing  = errors.New("author can't refuse editing")
)
