package repository

import (
	"context"
	"database/sql"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{
		db: db,
	}
}

func (r *NotificationRepository) GetNewNotifications(ctx context.Context, userID uint64) ([]domain.Notification, error) {
	rows, err := r.db.QueryContext(ctx, `
	SELECT 
		n.id,
		fu.username AS author_username, 
		fu.avatar AS author_avatar, 
		fu.is_external AS author_is_external, 
		ru.username AS receiver_username,
		n.notification_type, 
		n.is_read, 
		n.created_at, 
		n.additional
	FROM notification n
	LEFT JOIN flow_user fu ON n.author_id = fu.id
	LEFT JOIN flow_user ru ON n.receiver_id = ru.id
	WHERE n.receiver_id = $1
	ORDER BY n.created_at DESC;
	`, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var notifications []domain.Notification

	for rows.Next() {
		var notification domain.Notification
		if err := rows.Scan(
			&notification.ID,
			&notification.SenderUsername,
			&notification.SenderAvatar,
			&notification.SenderExternalAvatar,
			&notification.ReceiverUsername,
			&notification.Type,
			&notification.IsRead,
			&notification.CreatedAt,
			&notification.AdditionalData,
		); err != nil {
			return nil, err
		}

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *NotificationRepository) AddNotification(ctx context.Context, notification domain.Notification) error {
	var authorID int
	err := r.db.QueryRowContext(ctx, `
		SELECT id FROM flow_user WHERE username = $1
	`, notification.SenderUsername).Scan(&authorID)
	if err != nil {
		return err
	}

	var receiverID int
	err = r.db.QueryRowContext(ctx, `
		SELECT id FROM flow_user WHERE username = $1
	`, notification.ReceiverUsername).Scan(&receiverID)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO notification (author_id, receiver_id, notification_type, is_read, additional)
		VALUES ($1, $2, $3, $4, $5);
	`, authorID, receiverID, notification.Type, notification.IsRead, notification.AdditionalData)
	if err != nil {
		return err
	}

	return nil
}
