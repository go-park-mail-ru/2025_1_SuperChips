package repository

import "github.com/go-park-mail-ru/2025_1_SuperChips/internal/entity"

type UserStorage interface {
	NewStorage()
	AddUser(user entity.User) error
	LoginUser(email, password string) error
	GetUserPublicInfo(email string) (entity.PublicUser, error)
	GetUserId(email string) uint64
}

