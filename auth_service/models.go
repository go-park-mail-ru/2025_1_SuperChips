package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/validator"
	"github.com/go-park-mail-ru/2025_1_SuperChips/utils/wrapper"
)

type Board struct {
	ID             int       `json:"id"`
	AuthorID       int       `json:"author_id"`
	AuthorUsername string    `json:"author_username,omitempty"`
	Name           string    `json:"name"`
	CreatedAt      time.Time `json:"created_at"`
	IsPrivate      bool      `json:"is_private"`
	FlowCount      int       `json:"flow_count"`
	Preview        []PinData `json:"preview,omitempty"`
}

type User struct {
	Id         uint64    `json:"user_id,omitempty"`
	Username   string    `json:"username"`
	Password   string    `json:"password,omitempty"`
	Email      string    `json:"email"`
	Avatar     string    `json:"avatar,omitempty"`
	Birthday   time.Time `json:"birthday"`
	About      string    `json:"about,omitempty"`
	PublicName string    `json:"public_name,omitempty"`
	JWTVersion uint64    `json:"-"`
}

type PublicUser struct {
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Avatar     string    `json:"avatar,omitempty"`
	Birthday   time.Time `json:"birthday"`
	PublicName string    `json:"public_name,omitempty"`
	About      string    `json:"about,omitempty"`
}

type PinData struct {
	FlowID         uint64 `json:"flow_id,omitempty"`
	Header         string `json:"header,omitempty"`
	AuthorID       uint64 `json:"author_id,omitempty"`
	AuthorUsername string `json:"author_username"`
	Description    string `json:"description,omitempty"`
	MediaURL       string `json:"media_url,omitempty"`
	IsPrivate      bool   `json:"is_private"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
	IsLiked        bool   `json:"is_liked"`
	LikeCount      int    `json:"like_count"`
	Width          int    `json:"width,omitempty"`
	Height         int    `json:"height,omitempty"`
}

var (
	ErrForbidden            = errors.New("forbidden")
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

type StatusError interface {
	error
	StatusCode() int
}

type StatusCodeError struct {
	Code int
	Msg  string
}

func (e *StatusCodeError) Error() string {
	return e.Msg
}

func (e *StatusCodeError) StatusCode() int {
	return e.Code
}

var (
	ErrUnauthorized = &StatusCodeError{Code: 401, Msg: "invalid credentials"}
	ErrValidation   = &StatusCodeError{Code: 400, Msg: "validation failed"}
	ErrConflict     = &StatusCodeError{Code: 409, Msg: "resource conflict"}
	ErrNotFound     = &StatusCodeError{Code: 404, Msg: "resource not found"}
	ErrInternal     = &StatusCodeError{Code: 500, Msg: "internal server error"}
)

func WrapError(base error, err error) error {
	return fmt.Errorf("%w: %w", base, err)
}

func ValidateEmail(email string) error {
	v := validator.New()

	if !v.Check(len(email) <= 64 && len(email) > 3, "email", "cannot be shorter than 3 symbols or longer than 64 symbols") {
		return wrapper.WrapError(ErrValidation, v.GetError("email"))
	}

	if !isValidEmail(email) {
		return wrapper.WrapError(ErrValidation, ErrInvalidEmail)
	}

	return nil
}

func ValidatePassword(password string) error {
	v := validator.New()
	if !v.Check(password != "", "password", "cannot be empty") {
		return wrapper.WrapError(ErrValidation, v.GetError("password"))
	}

	if !v.Check(len(password) <= 96, "password", "cannot be longer than 96 symbols") {
		return wrapper.WrapError(ErrValidation, v.GetError("password"))
	}

	return nil
}

func ValidateEmailAndPassword(email, password string) error {
	if err := ValidateEmail(email); err != nil {
		return err
	}

	if err := ValidatePassword(password); err != nil {
		return err
	}

	return nil
}

func ValidateUsername(username string) error {
	v := validator.New()

	if !v.Check(len(username) <= 32 && len(username) > 2, "username", "username cannot be shorter than 2 symbols or longer than 32 symbols") {
		return wrapper.WrapError(ErrValidation, v.GetError("username"))
	}

	if !validator.Matches(username, validator.UsernameRX) {
		return ErrInvalidUsername
	}

	return nil
}

func (u User) ValidateUser() error {
	if err := ValidateEmailAndPassword(u.Email, u.Password); err != nil {
		return wrapper.WrapError(ErrValidation, err)
	}

	if err := ValidateUsername(u.Username); err != nil {
		return wrapper.WrapError(ErrValidation, err)
	}

	return nil
}

func (u User) ValidateUserNoPassword() error {
	if err := ValidateEmail(u.Email); err != nil {
		return wrapper.WrapError(ErrValidation, err)
	}

	if err := ValidateUsername(u.Username); err != nil {
		return wrapper.WrapError(ErrValidation, err)
	}

	return nil
}

func isValidEmail(email string) bool {
	return validator.Matches(email, validator.EmailRX)
}
