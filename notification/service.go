package notification

import (
	"context"
	"path/filepath"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type NotificationRepository interface {
	GetNewNotifications(ctx context.Context, userID uint64) ([]domain.Notification, error)
}

type NotificationService struct {
	repo      NotificationRepository
	baseURL   string
	staticDir string
	avatarDir string
}

func NewNotificationService(repo NotificationRepository, baseURL, staticDir, avatarDir string) *NotificationService {
	return &NotificationService{
		repo: repo,
		baseURL:   baseURL,
		staticDir: staticDir,
		avatarDir: avatarDir,
	}
}

func (s *NotificationService) GetNotifications(ctx context.Context, userID uint) ([]domain.Notification, error) {
	notifications, err := s.repo.GetNewNotifications(ctx, uint64(userID))
	if err != nil {
		return nil, err
	}

	for i := range notifications {
		if notifications[i].SenderExternalAvatar {
			notifications[i].SenderAvatar = s.generateAvatarURL(notifications[i].SenderAvatar)
		}
	}

	return notifications, nil
}

func (s *NotificationService) generateAvatarURL(filename string) string {
	if filename == "" {
		return ""
	}

	return s.baseURL + filepath.Join(s.staticDir, s.avatarDir, filename)
}

