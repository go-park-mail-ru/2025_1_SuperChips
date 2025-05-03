package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	gen "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/chat"
)

type ChatUsecase interface {
	GetChats(ctx context.Context, username string) ([]domain.Chat, error)
	CreateChat(ctx context.Context, targetUsername string) (int, string, error)
	GetContacts(ctx context.Context, username string) ([]domain.Contact, error)
	GetChat(ctx context.Context, id uint64) ([]domain.Chat, error)
	GetChatMessages(ctx context.Context, id, page uint64) ([]domain.Message, error)
}

// type ChatServiceClient interface {
// 	GetChats(ctx context.Context, in *GetChatsRequest, opts ...grpc.CallOption) (*ChatsStruct, error)
// 	CreateChat(ctx context.Context, in *CreateChatRequest, opts ...grpc.CallOption) (*CreateChatResponse, error)
// 	GetContacts(ctx context.Context, in *GetContactsRequest, opts ...grpc.CallOption) (*ContactsStruct, error)
// 	GetChat(ctx context.Context, in *GetChatRequest, opts ...grpc.CallOption) (*Chat, error)
// 	GetChatMessages(ctx context.Context, in *GetChatMessagesRequest, opts ...grpc.CallOption) (*MessagesStruct, error)
// }


type GrpcChatHandler struct {
	usecase ChatUsecase
	gen.UnimplementedChatServiceServer
}

func NewGrpcChatHandler(usecase ChatUsecase) *GrpcChatHandler {
	return &GrpcChatHandler{
		usecase: usecase,
	}
}

// todo
func (h *GrpcChatHandler) GetChats(ctx context.Context, in *gen.GetChatsRequest) (*gen.ChatsStruct, error) {
	return nil, nil
}

func (h *GrpcChatHandler) CreateChat(ctx context.Context, in *gen.CreateChatRequest) (*gen.CreateChatResponse, error) {
	return nil, nil
}

func (h *GrpcChatHandler) GetContacts(ctx context.Context, in *gen.GetContactsRequest) (*gen.ContactsStruct, error) {
	return nil, nil
}

func (h *GrpcChatHandler) GetChat(ctx context.Context, in *gen.GetChatRequest) (*gen.Chat, error) {
	return nil, nil
}

func (h *GrpcChatHandler) GetChatMessages(ctx context.Context, in *gen.GetChatMessagesRequest) (*gen.MessagesStruct, error) {
	return nil, nil
}
