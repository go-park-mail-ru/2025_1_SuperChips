package user

import (
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/security"
)

type UserRepository interface {
	AddUser(user domain.User) (uint64, error)
	GetHash(email, password string) (uint64, string, error)
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

func (u *UserService) AddUser(user domain.User) (uint64, error) {
	if err := user.ValidateUser(); err != nil {
		return 0, err
	}

	hashed, err := security.HashPassword(user.Password)
	if err != nil {
		return 0, err
	}

	user.Password = hashed

	id, err := u.repo.AddUser(user)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (u *UserService) LoginUser(email, password string) (uint64, error) {
	if err := domain.ValidateEmailAndPassword(email, password); err != nil {
		return 0, err
	}

	id, pswd, err := u.repo.GetHash(email, password)
	if err != nil {
		return 0, err
	}

	if !security.ComparePassword(password, pswd) {
		return 0, domain.ErrInvalidCredentials
	}

	return id, nil
}

func (u *UserService) GetUserPublicInfo(email string) (domain.PublicUser, error) {
	return u.repo.GetUserPublicInfo(email)
}

func (u *UserService) GetUserId(email string) (uint64, error) {
	return u.repo.GetUserId(email)
}

