package user

import (
	"errors"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/security"
	"regexp"
	"time"
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


var userBase = initUserStorage(&mapUserStorage{})
var id uint64 = 1

func isValidEmail(email string) bool {
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	return emailRegex.MatchString(email)
}

func (u User) ValidateUser() error {
	if len(u.Email) > 64 || len(u.Email) < 3 {
		return ErrInvalidEmail
	}

	if !isValidEmail(u.Email) {
		return ErrInvalidEmail
	}

	if len(u.Username) > 32 || len(u.Username) < 2 {
		return ErrInvalidUsername
	}

	if u.Password == "" {
		return ErrNoPassword
	}

	if len(u.Password) > 96 {
		return ErrPasswordTooLong
	}

	if u.Birthday.After(time.Now()) || time.Since(u.Birthday) > 150*365*24*time.Hour {
		return ErrInvalidBirthday
	}

	return nil
}

func AddUser(user User) error {
	if err := user.ValidateUser(); err != nil {
		return err
	}

	if userBase.repo.containsEmail(user.Email) {
		return ErrEmailAlreadyTaken
	}

	if userBase.repo.containsUsername(user.Username) {
		return ErrUsernameAlreadyTaken
	}

	user.Id = id
	id++

	hashPassword, err := security.HashPassword(user.Password)
	if err != nil {
		return ErrInternalError
	}

	user.Password = hashPassword
	userBase.repo.addUserToBase(user)

	return nil
}

func LoginUser(email, password string) error {
	user, found := userBase.repo.findUserByMail(email)
	if !found {
		return ErrInvalidCredentials
	}

	if !security.ComparePassword(password, user.Password) {
		return ErrInvalidCredentials
	}

	return nil
}

func GetUserPublicInfo(email string) (PublicUser, error) {
	user, found := userBase.repo.findUserByMail(email)
	if !found {
		return PublicUser{}, ErrUserNotFound
	}

	publicUser := PublicUser{
		Username: user.Username,
		Email:    user.Email,
		Birthday: user.Birthday,
		Avatar:   user.Avatar,
	}

	return publicUser, nil
}

func GetUserId(email string) uint64 {
	user, found := userBase.repo.findUserByMail(email)
	if !found {
		return 0
	}

	return user.Id
}

