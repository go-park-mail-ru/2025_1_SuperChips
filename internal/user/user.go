package user

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/security"
)

type User struct {
	id       uint64    `json:"-"`
	username string    `json:username`
	password string    `json:"-"`
	email    string    `json:email`
	avatar   string    `json:avatar,omitempty`
	birthday time.Time `json:birthday`
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
)

var users []User = make([]User, 1)
var id uint64 = 0

func containsUsername(username string) bool {
	for _, v := range users {
		if v.username == username {
			return true
		}
	}

	return false
}

func containsEmail(email string) bool {
	for _, v := range users {
		if v.email == email {
			return true
		}
	}

	return false
}

func findUserByMail(email string) (User, bool) {
	for _, v := range users {
		if v.email == email {
			return v, true
		}
	}

	return User{}, false
}

func (u User) ValidateUser() error {
	if len(u.email) > 32 || len(u.email) < 4 {
		return ErrInvalidEmail
	}

	if len(u.username) > 32 || len(u.username) < 2 {
		return ErrInvalidUsername
	}

	if u.password == "" {
		return ErrNoPassword
	}

	if len(u.password) > 64 {
		return ErrPasswordTooLong
	}

	if u.birthday.IsZero() {
		return ErrInvalidBirthday
	}

	return nil
}

func AddUser(user User) error {
	if err := user.ValidateUser(); err != nil {
		return err
	}

	if containsEmail(user.email) {
		return ErrEmailAlreadyTaken
	}

	if containsUsername(user.username) {
		return ErrUsernameAlreadyTaken
	}

	user.id = id
	id++

	hashPassword, err := security.HashPassword(user.password)
	if err != nil {
		return fmt.Errorf("internal error: %w", err)
	}

	user.password = hashPassword
	users = append(users, user)

	return nil
}

func LoginUser(email, password string) error {
	user, found := findUserByMail(email)
	if !found {
		return ErrInvalidCredentials
	}

	if !security.CheckPassword(password, user.password) {
		return ErrInvalidCredentials
	}

	return nil
}
