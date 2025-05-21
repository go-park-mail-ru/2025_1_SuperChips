package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

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
	var isExternalAvatar sql.NullBool
	var additionalByte []byte
	
	rows, err := r.db.QueryContext(ctx, `
	SELECT 
		n.id,
		fu.username AS author_username, 
		fu.avatar AS author_avatar, 
		fu.is_external_avatar AS author_is_external, 
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
			&isExternalAvatar,
			&notification.ReceiverUsername,
			&notification.Type,
			&notification.IsRead,
			&notification.CreatedAt,
			&additionalByte,
		); err != nil {
			return nil, err
		}

		notification.SenderExternalAvatar = isExternalAvatar.Bool

		// unmarshall byte into something
		var additional map[string]interface{}
		if err := json.Unmarshal(additionalByte, &additional); err != nil {
			return nil, err
		}

		notification.AdditionalData = additional

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *NotificationRepository) AddNotification(ctx context.Context, notification domain.Notification) (uint, time.Time, error) {
	var authorID int
	err := r.db.QueryRowContext(ctx, `
		SELECT id FROM flow_user WHERE username = $1
	`, notification.SenderUsername).Scan(&authorID)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("get notification sender err: %v", err)
	}

	var receiverID int
	err = r.db.QueryRowContext(ctx, `
		SELECT id FROM flow_user WHERE username = $1
	`, notification.ReceiverUsername).Scan(&receiverID)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("get notification receiver err: %v", err)
	}

	rawAdditional, err := json.Marshal(notification.AdditionalData)
	if err != nil {
		return 0, time.Time{}, err
	}

	err = r.db.QueryRowContext(ctx, `
		INSERT INTO notification (author_id, receiver_id, notification_type, is_read, additional)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, timestamp
	`, authorID, receiverID, notification.Type, notification.IsRead, rawAdditional).Scan()
	if err != nil {
		return 0, time.Time{}, err
	}

	return nil
}
