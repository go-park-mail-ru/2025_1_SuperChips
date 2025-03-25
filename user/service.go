package user

import (
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/security"
)

type UserRepository interface {
	AddUser(user domain.User) error
	LoginUser(email, password string) (string, error)
	GetUserPublicInfo(email string) (domain.PublicUser, error)
	GetUserId(email string) (uint64, error)	
}

type UserService struct {
	repo UserRepository
}

func NewUserService(u UserRepository) *UserService {
	return &UserService{
		repo: u,
	}
}

func (u *UserService) AddUser(user domain.User) error {
	if err := user.ValidateUser(); err != nil {
		return err
	}

	if err := u.repo.AddUser(user); err != nil {
		return err
	}

	return nil
}

func (u *UserService) LoginUser(email, password string) error {
	if err := domain.ValidateEmailAndPassword(email, password); err != nil {
		return err
	}

	pswd, err := u.repo.LoginUser(email, password)
	if err != nil {
		return err
	}

	if !security.ComparePassword(password, pswd) {
		return domain.ErrInvalidCredentials
	}

	return nil
}

func (u *UserService) GetUserPublicInfo(email string) (domain.PublicUser, error) {
	return u.repo.GetUserPublicInfo(email)
}

func (u *UserService) GetUserId(email string) (uint64, error) {
	return u.repo.GetUserId(email)
}

