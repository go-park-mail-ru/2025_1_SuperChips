package api

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_SuperChips/auth_service"
	gen "github.com/go-park-mail-ru/2025_1_SuperChips/auth_service/internal/proto/gen/auth"
)

type UserUsecaseInterface interface {
	AddUser(ctx context.Context, user models.User) (uint64, error)
	LoginUser(ctx context.Context, email, password string) (uint64, error)
	LoginExternalUser(ctx context.Context, email string, externalID string) (int, string, error)
	AddExternalUser(ctx context.Context, email, username string, externalID string) (uint64, error)
}

type GrpcAuthHandler struct {
	gen.UnimplementedAuthServer
	usecase UserUsecaseInterface
}

// mustEmbedUnimplementedAuthServer implements gen.AuthServer.
func (h *GrpcAuthHandler) mustEmbedUnimplementedAuthServer() {
	panic("unimplemented")
}

func NewGrpcAuthHandler(usecase UserUsecaseInterface) *GrpcAuthHandler {
	return &GrpcAuthHandler{
		usecase: usecase,
	}
}

func (h *GrpcAuthHandler) AddUser(ctx context.Context, in *gen.AddUserRequest) (*gen.AddUserResponse, error) {
	userData := models.User{
		Email:    in.Email,
		Username: in.Username,
		Password: in.Password,
	}

	id, err := h.usecase.AddUser(ctx, userData)
	if err != nil {
		return nil, err
	}

	return &gen.AddUserResponse{
		ID: int64(id),
	}, nil
}

func (h *GrpcAuthHandler) LoginUser(ctx context.Context, in *gen.LoginUserRequest) (*gen.LoginUserResponse, error) {
	id, err := h.usecase.LoginUser(ctx, in.Email, in.Password)
	if err != nil { 
		return nil, err
	}

	return &gen.LoginUserResponse{
		ID: int64(id),
	}, nil
}

func (h *GrpcAuthHandler) LoginExternalUser(ctx context.Context, in *gen.LoginExternalUserRequest) (*gen.LoginExternalUserResponse, error) {
	id, email, err := h.usecase.LoginExternalUser(ctx, in.Email, in.ExternalID)
	if err != nil {
		return nil, err
	}

	return &gen.LoginExternalUserResponse{
		ID:    int64(id),
		Email: email,
	}, nil
}

func (h *GrpcAuthHandler) AddExternalUser(ctx context.Context, in *gen.AddExternalUserRequest) (*gen.AddExternalUserResponse, error) {
	id, err := h.usecase.AddExternalUser(ctx, in.Email, in.Username, in.ExternalID)
	if err != nil {
		return nil, err
	}

	return &gen.AddExternalUserResponse{
		ID: int64(id),
	}, nil
}
