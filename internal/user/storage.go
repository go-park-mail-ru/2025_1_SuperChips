package user

import (
	"sync"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/security"
)


type MapUserStorage struct {
	users map[string]User
	id    uint
}

func NewMapUserStorage() MapUserStorage {
	newMap := MapUserStorage{}
	newMap.initialize()

	return newMap
}

func (storage *MapUserStorage) AddUser(user User) error {
	if err := user.ValidateUser(); err != nil {
		return err
	}

	if storage.containsEmail(user.Email) {
		return wrapError(ErrConflict, ErrEmailAlreadyTaken)
	}

	if storage.containsUsername(user.Username) {
		return wrapError(ErrConflict, ErrUsernameAlreadyTaken)
	}

	user.Id = uint64(storage.id)
	storage.id++

	hashPassword, err := security.HashPassword(user.Password)
	if err != nil {
		return wrapError(ErrInternal, nil)
	}

	user.Password = hashPassword
	storage.addUserToBase(user)

	return nil
}

func (storage MapUserStorage) LoginUser(email, password string) error {
	user, found := storage.findUserByMail(email)
	if !found {
		return wrapError(ErrForbidden, ErrInvalidCredentials)
	}

	if !security.ComparePassword(password, user.Password) {
		return wrapError(ErrForbidden, ErrInvalidCredentials)
	}

	return nil
}

func (storage MapUserStorage) GetUserPublicInfo(email string) (PublicUser, error) {
	user, found := storage.findUserByMail(email)
	if !found {
		return PublicUser{}, wrapError(ErrNotFound, ErrUserNotFound)
	}

	publicUser := PublicUser{
		Username: user.Username,
		Email:    user.Email,
		Birthday: user.Birthday,
		Avatar:   user.Avatar,
	}

	return publicUser, nil
}

func (storage MapUserStorage) GetUserId(email string) uint64 {
	user, found := storage.findUserByMail(email)
	if !found {
		return 0
	}

	return user.Id
}


func (u *MapUserStorage) initialize() {
	u.users = make(map[string]User, 0)
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

func (u MapUserStorage) findUserByMail(email string) (User, bool) {
	for _, v := range u.users {
		if v.Email == email {
			return v, true
		}
	}

	return User{}, false
}

func (u *MapUserStorage) addUserToBase(user User) {
	m := sync.RWMutex{}

	m.Lock()
	u.users[user.Email] = user
	m.Unlock()
}

