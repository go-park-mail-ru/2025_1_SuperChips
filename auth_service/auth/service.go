package auth

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2025_1_SuperChips/auth_service"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/security"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserRepository interface {
	AddUser(ctx context.Context, user models.User) (uint64, error)
	GetHash(ctx context.Context, email, password string) (uint64, string, error)
	GetUserPublicInfo(ctx context.Context, email string) (models.PublicUser, error)
	GetUserId(ctx context.Context, email string) (uint64, error)
	FindExternalServiceUser(ctx context.Context, email string, externalID string) (int, string, error)
	AddExternalUser(ctx context.Context, email, username, password string, externalID string) (uint64, error)
}

type BoardRepository interface {
	CreateBoard(ctx context.Context, board *models.Board, username string, userID int) error
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

func (u *UserService) AddUser(ctx context.Context, user models.User) (uint64, error) {
	if err := user.ValidateUser(); err != nil {
		return 0, mapToGrpcError(err)
	}

	hashed, err := security.HashPassword(user.Password)
	if err != nil {
		return 0, err
	}

	user.Password = hashed

	id, err := u.userRepo.AddUser(ctx, user)
	if err != nil {
		return 0, mapToGrpcError(err)
	}

	if err := u.createUserBoards(ctx, user.Username, int(id)); err != nil {
		return 0, mapToGrpcError(err)
	}

	return id, nil
}

func (u *UserService) LoginUser(ctx context.Context, email, password string) (uint64, error) {
	if err := models.ValidateEmailAndPassword(email, password); err != nil {
		return 0, mapToGrpcError(err)
	}

	id, pswd, err := u.userRepo.GetHash(ctx, email, password)
	if err != nil {
		return 0, err
	}

	if !security.ComparePassword(password, pswd) {
		return 0, mapToGrpcError(models.ErrInvalidCredentials)
	}

	return id, nil
}

func (u *UserService) LoginExternalUser(ctx context.Context, email string, externalID string) (int, string, error) {
	id, gotEmail, err := u.userRepo.FindExternalServiceUser(ctx, email, externalID)
	if err != nil {
		return 0, "", err
	}

	// this error shouldn't happen ever
	if gotEmail != email {
		return 0, "", mapToGrpcError(models.ErrForbidden)
	}

	return id, gotEmail, nil
}

func (u *UserService) AddExternalUser(ctx context.Context, email, username string, externalID string) (uint64, error) {
	dummyPassword, err := security.GenerateRandomHash()
	if err != nil {
		return 0, err
	}

	dummyPassword, err = security.HashPassword(dummyPassword)
	if err != nil {
		return 0, err
	}

	id, err := u.userRepo.AddExternalUser(ctx, email, username, dummyPassword, externalID)
	if err != nil {
		return 0, err
	}

	if err := u.createUserBoards(ctx, username, int(id)); err != nil {
		return 0, err
	}

	return id, nil
}

func (u *UserService) GetUserPublicInfo(ctx context.Context, email string) (models.PublicUser, error) {
	return u.userRepo.GetUserPublicInfo(ctx, email)
}

func (u *UserService) GetUserId(ctx context.Context, email string) (uint64, error) {
	return u.userRepo.GetUserId(ctx, email)
}

func (u *UserService) createUserBoards(ctx context.Context, username string, id int) error {
	if err := u.boardRepo.CreateBoard(ctx, &models.Board{
		Name: "Созданные вами",
	}, username, int(id)); err != nil {
		return err
	}

	if err := u.boardRepo.CreateBoard(ctx, &models.Board{
		Name: "Сохраненные",
	}, username, int(id)); err != nil {
		return err
	}

	return nil
}

func mapToGrpcError(err error) error {
    switch {
    case errors.Is(err, models.ErrInvalidCredentials):
        return status.Errorf(codes.Unauthenticated, "invalid credentials")
    case errors.Is(err, models.ErrUserNotFound):
        return status.Errorf(codes.NotFound, "user not found")
	case errors.Is(err, models.ErrForbidden):
		return status.Errorf(codes.PermissionDenied, "forbidden")
	case errors.Is(err, models.ErrConflict):
		return status.Errorf(codes.AlreadyExists, "conflict")
	case errors.Is(err, models.ErrValidation):
		return status.Errorf(codes.InvalidArgument, "validation error")
    default:
        return status.Errorf(codes.Internal, "internal server error")
    }
}