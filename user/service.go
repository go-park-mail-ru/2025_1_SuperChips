package user

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/security"
)

type UserRepository interface {
	AddUser(ctx context.Context, user domain.User) (uint64, error)
	GetHash(ctx context.Context, email, password string) (uint64, string, error)
	GetUserPublicInfo(ctx context.Context, email string) (domain.PublicUser, error)
	GetUserId(ctx context.Context, email string) (uint64, error)
}

type BoardRepository interface {
	CreateBoard(ctx context.Context, board *domain.Board, username string, userID int) error
}

type UserService struct {
	userRepo  UserRepository
	boardRepo BoardRepository
}

func NewUserService(u UserRepository, b BoardRepository) *UserService {
	return &UserService{
		userRepo: u,
		boardRepo: b,
	}
}

func (u *UserService) AddUser(ctx context.Context, user domain.User) (uint64, error) {
	if err := user.ValidateUser(); err != nil {
		return 0, err
	}

	hashed, err := security.HashPassword(user.Password)
	if err != nil {
		return 0, err
	}

	user.Password = hashed

	id, err := u.userRepo.AddUser(ctx, user)
	if err != nil {
		return 0, err
	}

	if err := u.boardRepo.CreateBoard(ctx, &domain.Board{
		Name: "Созданные вами",
	}, user.Username, int(id)); err != nil {
		return 0, err
	}

	if err := u.boardRepo.CreateBoard(ctx, &domain.Board{
		Name: "Сохраненные",
	}, user.Username, int(id)); err != nil {
		return 0, err
	}

	return id, nil
}

func (u *UserService) LoginUser(ctx context.Context, email, password string) (uint64, error) {
	if err := domain.ValidateEmailAndPassword(email, password); err != nil {
		return 0, err
	}

	id, pswd, err := u.userRepo.GetHash(ctx, email, password)
	if err != nil {
		return 0, err
	}

	if !security.ComparePassword(password, pswd) {
		return 0, domain.ErrInvalidCredentials
	}

	return id, nil
}

func (u *UserService) GetUserPublicInfo(ctx context.Context, email string) (domain.PublicUser, error) {
	return u.userRepo.GetUserPublicInfo(ctx, email)
}

func (u *UserService) GetUserId(ctx context.Context, email string) (uint64, error) {
	return u.userRepo.GetUserId(ctx, email)
}
