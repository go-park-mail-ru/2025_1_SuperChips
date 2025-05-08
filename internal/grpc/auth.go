package grpc

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	gen "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserUsecase interface {
	AddUser(ctx context.Context, user domain.User) (uint64, error)
	LoginUser(ctx context.Context, email, password string) (uint64, string, error)
	LoginExternalUser(ctx context.Context, email string, externalID string) (int, string, string, error)
	AddExternalUser(ctx context.Context, email, username, avatarURL string, externalID string) (uint64, error)
	CheckImgPermission(ctx context.Context, imageName string, userID int) (bool, error)
}

type GrpcAuthHandler struct {
	gen.UnimplementedAuthServer
	usecase UserUsecase
}

// mustEmbedUnimplementedAuthServer implements gen.AuthServer.
func (h *GrpcAuthHandler) mustEmbedUnimplementedAuthServer() {
	panic("unimplemented")
}

func NewGrpcAuthHandler(usecase UserUsecase) *GrpcAuthHandler {
	return &GrpcAuthHandler{
		usecase: usecase,
	}
}

func (h *GrpcAuthHandler) AddUser(ctx context.Context, in *gen.AddUserRequest) (*gen.AddUserResponse, error) {
	userData := domain.User{
		Email:    in.Email,
		Username: in.Username,
		Password: in.Password,
	}

	id, err := h.usecase.AddUser(ctx, userData)
	if err != nil {
		return nil, mapToGrpcError(err)
	}

	return &gen.AddUserResponse{
		ID: int64(id),
	}, nil
}

func (h *GrpcAuthHandler) LoginUser(ctx context.Context, in *gen.LoginUserRequest) (*gen.LoginUserResponse, error) {
	id, username, err := h.usecase.LoginUser(ctx, in.Email, in.Password)
	if err != nil { 
		return nil, mapToGrpcError(err)
	}

	return &gen.LoginUserResponse{
		ID: int64(id),
		Username: username,
	}, nil
}

func (h *GrpcAuthHandler) LoginExternalUser(ctx context.Context, in *gen.LoginExternalUserRequest) (*gen.LoginExternalUserResponse, error) {
	id, email, username, err := h.usecase.LoginExternalUser(ctx, in.Email, in.ExternalID)
	if err != nil {
		return nil, mapToGrpcError(err)
	}

	return &gen.LoginExternalUserResponse{
		ID:    int64(id),
		Email: email,
		Username: username,
	}, nil
}

func (h *GrpcAuthHandler) AddExternalUser(ctx context.Context, in *gen.AddExternalUserRequest) (*gen.AddExternalUserResponse, error) {
	id, err := h.usecase.AddExternalUser(ctx, in.Email, in.Username, in.Avatar, in.ExternalID)
	if err != nil {
		return nil, mapToGrpcError(err)
	}

	return &gen.AddExternalUserResponse{
		ID: int64(id),
	}, nil
}

func (h *GrpcAuthHandler) CheckImgPermission(ctx context.Context, in *gen.CheckImgPermissionRequest) (*gen.CheckImgPermissionResponse, error) {
	hasAccess, err := h.usecase.CheckImgPermission(ctx, in.ImageName, int(in.ID))
	if err != nil {
		return nil, err
	}

	return &gen.CheckImgPermissionResponse{
		HasAccess: hasAccess,
	}, nil
}

func mapToGrpcError(err error) error {
    switch {
    case errors.Is(err, domain.ErrInvalidCredentials):
        return status.Errorf(codes.Unauthenticated, "invalid credentials")
    case errors.Is(err, domain.ErrUserNotFound), errors.Is(err, domain.ErrNotFound):
        return status.Errorf(codes.NotFound, "user not found")
	case errors.Is(err, domain.ErrForbidden):
		return status.Errorf(codes.PermissionDenied, "forbidden")
	case errors.Is(err, domain.ErrConflict):
		return status.Errorf(codes.AlreadyExists, "conflict")
	case errors.Is(err, domain.ErrValidation):
		return status.Errorf(codes.InvalidArgument, "validation error")
    default:
        return status.Errorf(codes.Internal, "internal server error")
    }
}
