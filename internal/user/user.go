package user

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

type StatusError interface {
	error
	StatusCode() int
}

type statusError struct {
	code int
	msg  string
}

func (e *statusError) Error() string {
	return e.msg
}

func (e *statusError) StatusCode() int {
	return e.code
}

type User struct {
	Id       uint64    `json:"-"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	Email    string    `json:"email"`
	Avatar   string    `json:"avatar,omitempty"`
	Birthday time.Time `json:"birthday"`
}

type PublicUser struct {
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Avatar   string    `json:"avatar,omitempty"`
	Birthday time.Time `json:"birthday"`
}

var (
	ErrForbidden  = &statusError{code: 403, msg: "invalid credentials"}
	ErrValidation = &statusError{code: 400, msg: "validation failed"}
	ErrConflict   = &statusError{code: 409, msg: "resource conflict"}
	ErrNotFound   = &statusError{code: 404, msg: "resource not found"}
	ErrInternal   = &statusError{code: 500, msg: "internal server error"}
)

var (
	ErrInvalidEmail         = errors.New("invalid email")
	ErrInvalidUsername      = errors.New("invalid username")
	ErrNoPassword           = errors.New("no password")
	ErrInvalidBirthday      = errors.New("invalid birthday")
	ErrEmailAlreadyTaken    = errors.New("the email is already used")
	ErrUsernameAlreadyTaken = errors.New("the username is already used")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrPasswordTooLong      = errors.New("password is too long")
	ErrInternalError        = errors.New("internal error")
	ErrUserNotFound         = errors.New("user not found")
)

func wrapError(base error, err error) error {
	return fmt.Errorf("%w: %w", base, err)
}

func isValidEmail(email string) bool {
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	return emailRegex.MatchString(email)
}

func (u User) ValidateUser() error {
	if len(u.Email) > 64 || len(u.Email) < 3 {
		return wrapError(ErrValidation, ErrInvalidEmail)
	}

	if !isValidEmail(u.Email) {
		return wrapError(ErrValidation, ErrInvalidEmail)
	}

	if len(u.Username) > 32 || len(u.Username) < 2 {
		return wrapError(ErrValidation, ErrInvalidUsername)
	}

	if u.Password == "" {
		return wrapError(ErrValidation, ErrNoPassword)
	}

	if len(u.Password) > 96 {
		return wrapError(ErrValidation, ErrPasswordTooLong)
	}

	if u.Birthday.After(time.Now()) || time.Since(u.Birthday) > 150*365*24*time.Hour {
		return wrapError(ErrValidation, ErrInvalidBirthday)
	}

	return nil
}

