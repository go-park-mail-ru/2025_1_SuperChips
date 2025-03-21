package repository

import (
	"fmt"
	"sync"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	security "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/security"
)

type MapUserStorage struct {
	users map[string]domain.User
	id    uint
}

func NewMapUserStorage() MapUserStorage {
	strg := MapUserStorage{}
	strg.initialize()

	return strg
}

func (storage *MapUserStorage) AddUser(user domain.User) error {
	if err := user.ValidateUser(); err != nil {
		return err
	}

	if storage.containsEmail(user.Email) {
		return wrapError(domain.ErrConflict, domain.ErrEmailAlreadyTaken)
	}

	if storage.containsUsername(user.Username) {
		return wrapError(domain.ErrConflict, domain.ErrUsernameAlreadyTaken)
	}

	user.Id = uint64(storage.id)
	storage.id++

	hashPassword, err := security.HashPassword(user.Password)
	if err != nil {
		return wrapError(domain.ErrInternal, nil)
	}

	user.Password = hashPassword
	storage.addUserToBase(user)

	return nil
}

func (storage *MapUserStorage) LoginUser(email, password string) error {
	user, found := storage.findUserByMail(email)
	if !found {
		return wrapError(domain.ErrUnauthorized, domain.ErrInvalidCredentials)
	}

	if !security.ComparePassword(password, user.Password) {
		return wrapError(domain.ErrUnauthorized, domain.ErrInvalidCredentials)
	}

	return nil
}

func (storage *MapUserStorage) GetUserPublicInfo(email string) (domain.PublicUser, error) {
	user, found := storage.findUserByMail(email)
	if !found {
		return domain.PublicUser{}, wrapError(domain.ErrNotFound, domain.ErrUserNotFound)
	}

	publicUser := domain.PublicUser{
		Username: user.Username,
		Email:    user.Email,
		Birthday: user.Birthday,
		Avatar:   user.Avatar,
	}

	return publicUser, nil
}

func (storage *MapUserStorage) GetUserId(email string) uint64 {
	user, found := storage.findUserByMail(email)
	if !found {
		return 0
	}

	return user.Id
}

func (u *MapUserStorage) initialize() {
	u.users = make(map[string]domain.User, 0)
	u.id = 1
}

func (u MapUserStorage) containsUsername(username string) bool {
	for _, v := range u.users {
		if v.Username == username {
			return true
		}
	}

	return false
}

func (u MapUserStorage) containsEmail(email string) bool {
	for _, v := range u.users {
		if v.Email == email {
			return true
		}
	}

	return false
}

func (u MapUserStorage) findUserByMail(email string) (domain.User, bool) {
	for _, v := range u.users {
		if v.Email == email {
			return v, true
		}
	}

	return domain.User{}, false
}

func (u *MapUserStorage) addUserToBase(user domain.User) {
	m := sync.RWMutex{}

	m.Lock()
	u.users[user.Email] = user
	m.Unlock()
}

func wrapError(base error, err error) error {
	return fmt.Errorf("%w: %w", base, err)
}

