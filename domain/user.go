package domain

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

type User struct {
	Id         uint64    `json:"-"`
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	Email      string    `json:"email"`
	Avatar     string    `json:"avatar,omitempty"`
	Birthday   time.Time `json:"birthday"`
	About      string    `json:"about,omitempty"`
	PublicName string    `json:"public_name,omitempty"`
	JWTVersion uint64    `json:"-"`
}

type PublicUser struct {
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Avatar   string    `json:"avatar,omitempty"`
	Birthday time.Time `json:"birthday"`
}

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

func ValidateEmailAndPassword(email, password string) error {
	if len(email) > 64 || len(email) < 3 {
		return WrapError(ErrValidation, ErrInvalidEmail)
	}

	if !isValidEmail(email) {
		return WrapError(ErrValidation, ErrInvalidEmail)
	}

	if password == "" {
		return WrapError(ErrValidation, ErrNoPassword)
	}

	if len(password) > 96 {
		return WrapError(ErrValidation, ErrPasswordTooLong)
	}

	return nil
}

func (u User) ValidateUser() error {
	if err := ValidateEmailAndPassword(u.Email, u.Password); err != nil {
		return err
	}

	if len(u.Username) > 32 || len(u.Username) < 2 {
		return WrapError(ErrValidation, ErrInvalidUsername)
	}

	if u.Birthday.After(time.Now()) || time.Since(u.Birthday) > 150*365*24*time.Hour {
		return WrapError(ErrValidation, ErrInvalidBirthday)
	}

	return nil
}

func WrapError(base error, err error) error {
	return fmt.Errorf("%w: %w", base, err)
}

func isValidEmail(email string) bool {
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	return emailRegex.MatchString(email)
}
