package usecase

import (
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/entity"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository"
)

type UserUsecase struct {
	repo repository.UserStorage
}

func NewUserService(repo repository.UserStorage) *UserUsecase {
    return &UserUsecase{repo: repo}
}

func (u *UserUsecase) AddUser(user entity.User) error {
	if err := user.ValidateUser(); err != nil {
		return err
	}

	err := u.repo.AddUser(user)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserUsecase) CheckCredentials(email, password string) error {
	if err := entity.ValidateEmailAndPassword(email, password); err != nil {
		return err
	}

	if err := u.repo.LoginUser(email, password); err != nil {
		return err
	}

	return nil
}

func (u *UserUsecase) GetUserId(email string) uint64 {
	return u.repo.GetUserId(email)
}

func (u *UserUsecase) GetUserPublicInfo(email string) (entity.PublicUser, error) {
	return u.repo.GetUserPublicInfo(email)
}