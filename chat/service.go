package chat

import (
	"context"
	"path/filepath"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type ChatRepository interface {
	GetChats(ctx context.Context, username string) ([]domain.Chat, error)
	CreateChat(ctx context.Context, username, targetUsername string) (domain.Chat, error)
	GetContacts(ctx context.Context, username string) ([]domain.Contact, error)
	CreateContact(ctx context.Context, username, targetUsername string) (domain.Chat, error)
	GetChat(ctx context.Context, id uint64, username string) (domain.Chat, error)
}

type ChatService struct {
	repo      ChatRepository
	baseURL   string
	staticDir string
	avatarDir string
}

func NewChatService(repo ChatRepository, baseURL, staticDir, avatarDir string) *ChatService {
	return &ChatService{
		repo: repo,
		baseURL: baseURL,
		staticDir: staticDir,
		avatarDir: avatarDir,
	}
}

func (service *ChatService) GetChats(ctx context.Context, username string) ([]domain.Chat, error) {
	chats, err := service.repo.GetChats(ctx, username)
	if err != nil {
		return nil, err
	}

	for i := range chats {
		if !chats[i].IsExternalAvatar {
			chats[i].Avatar = service.generateAvatarURL(chats[i].Avatar)
		}
	}

	return chats, nil
}

func (service *ChatService) CreateChat(ctx context.Context, username, targetUsername string) (domain.Chat, error) {
	chat, err := service.repo.CreateChat(ctx, username, targetUsername)
	if err != nil {
		return domain.Chat{}, err
	}

	if !chat.IsExternalAvatar {
		chat.Avatar = service.generateAvatarURL(chat.Avatar)
	}

	return chat, nil
}

func (service *ChatService) GetContacts(ctx context.Context, username string) ([]domain.Contact, error) {
	contacts, err := service.repo.GetContacts(ctx, username)
	if err != nil {
		return nil, err
	}

	for i := range contacts {
		if !contacts[i].IsExternalAvatar {
			contacts[i].Avatar = service.generateAvatarURL(contacts[i].Avatar)
		}
	}

	return contacts, nil
}

func (service *ChatService) CreateContact(ctx context.Context, username, targetUsername string) (domain.Chat, error) {
	chat, err := service.repo.CreateContact(ctx, username, targetUsername)
	if err != nil {
		return domain.Chat{}, err
	}

	if !chat.IsExternalAvatar {
		chat.Avatar = service.generateAvatarURL(chat.Avatar)
	}

	return chat, nil
}

func (service *ChatService) GetChat(ctx context.Context, id uint64, username string) (domain.Chat, error) {
	chat, err := service.repo.GetChat(ctx, id, username)
	if err != nil {
		return domain.Chat{}, err
	}

	if !chat.IsExternalAvatar {
		chat.Avatar = service.generateAvatarURL(chat.Avatar)
	}

	return chat, nil
}

func (s *ChatService) generateAvatarURL(filename string) string {
	if filename == "" {
		return ""
	}

	return s.baseURL + filepath.Join(s.staticDir, s.avatarDir, filename)
}
