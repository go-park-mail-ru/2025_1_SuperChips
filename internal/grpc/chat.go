package grpc

import (
	"context"
	"errors"
	"log"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	gen "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/chat"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type ChatUsecase interface {
	GetChats(ctx context.Context, username string) ([]domain.Chat, error)
	CreateChat(ctx context.Context, username, targetUsername string) (domain.Chat, error)
	GetContacts(ctx context.Context, username string) ([]domain.Contact, error)
	CreateContact(ctx context.Context, username, targetUsername string) (domain.Chat, error)
	GetChat(ctx context.Context, id uint64, username string) (domain.Chat, error)
	GetChatMessages(ctx context.Context, id, page uint64) ([]domain.Message, error)
}

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
	chats, err := h.usecase.GetChats(ctx, in.Username)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &gen.ChatsStruct{
		Chats: chatsToGrpc(chats),
	}, nil
}

func (h *GrpcChatHandler) CreateChat(ctx context.Context, in *gen.CreateChatRequest) (*gen.CreateChatResponse, error) {
	chat, err := h.usecase.CreateChat(ctx, in.Username, in.TargetUsername)
	if err != nil {
		log.Println(err)
		return nil, mapChatErrToGrpc(err)
	}

	grpcChats := chatsToGrpc([]domain.Chat{chat})
	return &gen.CreateChatResponse{
		Chat: grpcChats[0],
	}, nil
}

func (h *GrpcChatHandler) GetContacts(ctx context.Context, in *gen.GetContactsRequest) (*gen.ContactsStruct, error) {
	contacts, err := h.usecase.GetContacts(ctx, in.Username)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &gen.ContactsStruct{
		Contacts: contactsToGrpc(contacts),
	}, nil
}

func (h *GrpcChatHandler) CreateContact(ctx context.Context, in *gen.CreateContactRequest) (*gen.CreateContactResponse, error) {
	resp, err := h.usecase.CreateContact(ctx, in.Username, in.TargetUsername)
	if err != nil {
		log.Println(err)
		return nil, mapChatErrToGrpc(err)
	}

	return &gen.CreateContactResponse{
		ChatID:     uint64(resp.ChatID),
		Avatar:     resp.Avatar,
		PublicName: resp.PublicName,
	}, nil
}

func (h *GrpcChatHandler) GetChat(ctx context.Context, in *gen.GetChatRequest) (*gen.Chat, error) {
	chat, err := h.usecase.GetChat(ctx, in.ChatID, in.Username)
	if err != nil {
		log.Println(err)
		return nil, mapChatErrToGrpc(err)
	}

	grpcChat := chatsToGrpc([]domain.Chat{chat})

	return grpcChat[0], nil
}

func (h *GrpcChatHandler) GetChatMessages(ctx context.Context, in *gen.GetChatMessagesRequest) (*gen.MessagesStruct, error) {
	return nil, nil
}

func chatsToGrpc(chats []domain.Chat) []*gen.Chat {
	var grpc []*gen.Chat

	for i := range chats {
		chat := chats[i]
		grpc = append(grpc, &gen.Chat{
			ChatID:       uint64(chat.ChatID),
			Username:     chat.Username,
			Avatar:       chat.Avatar,
			PublicName:   chat.PublicName,
			MessageCount: uint64(chat.MessageCount),
			Messages: &gen.MessagesStruct{
				Messages: messagesToGrpc(chat.Messages),
			},
		})
	}

	return grpc
}

func messagesToGrpc(messages []domain.Message) []*gen.Message {
	var grpc []*gen.Message

	for i := range messages {
		message := messages[i]
		grpc = append(grpc, &gen.Message{
			MessageID: uint64(message.MessageID),
			Content:   message.Content,
			Sender:    message.Sender,
			Timestamp: timestamppb.New(message.Timestamp),
			IsRead:    message.IsRead,
			Recipient: message.Recipient,
			ChatID:    message.ChatID,
		})
	}

	return grpc
}

func contactsToGrpc(contacts []domain.Contact) []*gen.Contact {
	var grpc []*gen.Contact

	for i := range contacts {
		contact := contacts[i]
		grpc = append(grpc, &gen.Contact{
			Username:       contact.Username,
			PublicUsername: contact.Username,
			Avatar:         contact.Avatar,
		})
	}

	return grpc
}

func mapChatErrToGrpc(err error) error {
	switch {
	case errors.Is(err, domain.ErrConflict):
		return status.Errorf(codes.AlreadyExists, "conflict")
	case errors.Is(err, domain.ErrForbidden):
		return status.Error(codes.PermissionDenied, "forbidden")
	}

	return err
}