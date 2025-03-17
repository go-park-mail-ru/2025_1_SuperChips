package user

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/errs"
)

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
		return wrapError(errs.ErrValidation, ErrInvalidEmail)
	}

	if !isValidEmail(email) {
		return wrapError(errs.ErrValidation, ErrInvalidEmail)
	}

	if password == "" {
		return wrapError(errs.ErrValidation, ErrNoPassword)
	}

	if len(password) > 96 {
		return wrapError(errs.ErrValidation, ErrPasswordTooLong)
	}

	return nil
}

func (u User) ValidateUser() error {
	if err := ValidateEmailAndPassword(u.Email, u.Password); err != nil {
		return err
	}

	if len(u.Username) > 32 || len(u.Username) < 2 {
		return wrapError(errs.ErrValidation, ErrInvalidUsername)
	}

	if u.Birthday.After(time.Now()) || time.Since(u.Birthday) > 150*365*24*time.Hour {
		return wrapError(errs.ErrValidation, ErrInvalidBirthday)
	}

	return nil
}

func wrapError(base error, err error) error {
	return fmt.Errorf("%w: %w", base, err)
}

func isValidEmail(email string) bool {
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	return emailRegex.MatchString(email)
}

