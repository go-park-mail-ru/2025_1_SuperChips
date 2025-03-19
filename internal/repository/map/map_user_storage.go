package repository

import (
	"fmt"
	"sync"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/adapter/security"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/entity"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/errs"
)

type MapUserStorage struct {
	users map[string]entity.User
	id    uint
}

func (storage *MapUserStorage) NewStorage() {
	storage.initialize()
}

func (storage *MapUserStorage) AddUser(user entity.User) error {
	if err := user.ValidateUser(); err != nil {
		return err
	}

	if storage.containsEmail(user.Email) {
		return wrapError(errs.ErrConflict, entity.ErrEmailAlreadyTaken)
	}

	if storage.containsUsername(user.Username) {
		return wrapError(errs.ErrConflict, entity.ErrUsernameAlreadyTaken)
	}

	user.Id = uint64(storage.id)
	storage.id++

	hashPassword, err := security.HashPassword(user.Password)
	if err != nil {
		return wrapError(errs.ErrInternal, nil)
	}

	user.Password = hashPassword
	storage.addUserToBase(user)

	return nil
}

func (storage *MapUserStorage) LoginUser(email, password string) error {
	user, found := storage.findUserByMail(email)
	if !found {
		return wrapError(errs.ErrUnauthorized, entity.ErrInvalidCredentials)
	}

	if !security.ComparePassword(password, user.Password) {
		return wrapError(errs.ErrUnauthorized, entity.ErrInvalidCredentials)
	}

	return nil
}

func (storage *MapUserStorage) GetUserPublicInfo(email string) (entity.PublicUser, error) {
	user, found := storage.findUserByMail(email)
	if !found {
		return entity.PublicUser{}, wrapError(errs.ErrNotFound, entity.ErrUserNotFound)
	}

	publicUser := entity.PublicUser{
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
	u.users = make(map[string]entity.User, 0)
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

func (u MapUserStorage) findUserByMail(email string) (entity.User, bool) {
	for _, v := range u.users {
		if v.Email == email {
			return v, true
		}
	}

	return entity.User{}, false
}

func (u *MapUserStorage) addUserToBase(user entity.User) {
	m := sync.RWMutex{}

	m.Lock()
	u.users[user.Email] = user
	m.Unlock()
}

func wrapError(base error, err error) error {
	return fmt.Errorf("%w: %w", base, err)
}

