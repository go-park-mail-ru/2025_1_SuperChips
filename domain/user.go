package domain

import (
	"errors"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/validator"
	"github.com/go-park-mail-ru/2025_1_SuperChips/utils/wrapper"
)

type User struct {
	Id         uint64    `json:"user_id,omitempty"`
	Username   string    `json:"username"`
	Password   string    `json:"password,omitempty"`
	Email      string    `json:"email"`
	Avatar     string    `json:"avatar,omitempty"`
	Birthday   time.Time `json:"birthday,omitempty"`
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
	v := validator.New()

	if err := ValidateEmail(u.Email); err != nil {
		return wrapper.WrapError(ErrValidation, err)
	}

	if err := ValidateUsername(u.Username); err != nil {
		return wrapper.WrapError(ErrValidation, err)
	}

	if v.Check(u.Birthday.After(time.Now()) || time.Since(u.Birthday) > 150*365*24*time.Hour, "birthday", "cannot be too old") {
		return wrapper.WrapError(ErrValidation, v.GetError("birthday"))
	}

	return nil
}

func isValidEmail(email string) bool {
	return validator.Matches(email, validator.EmailRX)
}
