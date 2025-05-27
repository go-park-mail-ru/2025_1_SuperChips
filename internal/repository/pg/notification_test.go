package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/stretchr/testify/assert"
)

func setupNotificationTest(t *testing.T) (*NotificationRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}

	repo := NewNotificationRepository(db)
	return repo, mock, func() { db.Close() }
}

func TestGetNewNotifications_Success(t *testing.T) {
	repo, mock, closeFn := setupNotificationTest(t)
	defer closeFn()

	ctx := context.Background()
	userID := uint64(1)
	now := time.Now()

	additionalData := map[string]interface{}{"key": "value"}
	additionalBytes, _ := json.Marshal(additionalData)

	rows := sqlmock.NewRows([]string{
		"id", "author_username", "author_avatar", "author_is_external",
		"receiver_username", "notification_type", "is_read", "created_at", "additional",
	}).AddRow(
		1, "sender1", "avatar1.jpg", true,
		"receiver1", "friend_request", false, now, additionalBytes,
	).AddRow(
		2, "sender2", "avatar2.jpg", false,
		"receiver1", "message", true, now.Add(-time.Hour), []byte("{}"),
	)

	mock.ExpectQuery(`
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
	WHERE n.receiver_id = \$1
	ORDER BY n.created_at DESC;
	`).WithArgs(userID).WillReturnRows(rows)

	notifications, err := repo.GetNewNotifications(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, notifications, 2)

	assert.Equal(t, uint(1), notifications[0].ID)
	assert.Equal(t, "sender1", notifications[0].SenderUsername)
	assert.Equal(t, "avatar1.jpg", notifications[0].SenderAvatar)
	assert.True(t, notifications[0].SenderExternalAvatar)
	assert.Equal(t, "receiver1", notifications[0].ReceiverUsername)
	assert.Equal(t, "friend_request", notifications[0].Type)
	assert.False(t, notifications[0].IsRead)
	assert.Equal(t, additionalData, notifications[0].AdditionalData)

	assert.Equal(t, uint(2), notifications[1].ID)
	assert.Equal(t, "sender2", notifications[1].SenderUsername)
	assert.Equal(t, "avatar2.jpg", notifications[1].SenderAvatar)
	assert.False(t, notifications[1].SenderExternalAvatar)
	assert.Equal(t, "receiver1", notifications[1].ReceiverUsername)
	assert.Equal(t, "message", notifications[1].Type)
	assert.True(t, notifications[1].IsRead)
	assert.Equal(t, map[string]interface{}{}, notifications[1].AdditionalData)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNewNotifications_Empty(t *testing.T) {
	repo, mock, closeFn := setupNotificationTest(t)
	defer closeFn()

	ctx := context.Background()
	userID := uint64(1)

	rows := sqlmock.NewRows([]string{
		"id", "author_username", "author_avatar", "author_is_external",
		"receiver_username", "notification_type", "is_read", "created_at", "additional",
	})

	mock.ExpectQuery(`
	SELECT .*
	`).WithArgs(userID).WillReturnRows(rows)

	notifications, err := repo.GetNewNotifications(ctx, userID)
	assert.NoError(t, err)
	assert.Empty(t, notifications)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNewNotifications_DBError(t *testing.T) {
	repo, mock, closeFn := setupNotificationTest(t)
	defer closeFn()

	ctx := context.Background()
	userID := uint64(1)

	mock.ExpectQuery(`
	SELECT .*
	`).WithArgs(userID).WillReturnError(errors.New("database error"))

	_, err := repo.GetNewNotifications(ctx, userID)
	assert.Error(t, err)
	assert.EqualError(t, err, "database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddNotification_Success(t *testing.T) {
	repo, mock, closeFn := setupNotificationTest(t)
	defer closeFn()

	ctx := context.Background()
	now := time.Now()
	additionalData := map[string]interface{}{"request_id": 123}
	notification := domain.Notification{
		SenderUsername:   "sender1",
		ReceiverUsername: "receiver1",
		Type:             "friend_request",
		IsRead:           false,
		AdditionalData:   additionalData,
	}

	senderRow := sqlmock.NewRows([]string{"id", "avatar", "is_external_avatar"}).
		AddRow(1, "avatar1.jpg", true)
	mock.ExpectQuery(`
		SELECT id, avatar, is_external_avatar FROM flow_user WHERE username = \$1
	`).WithArgs("sender1").WillReturnRows(senderRow)

	receiverRow := sqlmock.NewRows([]string{"id"}).AddRow(2)
	mock.ExpectQuery(`
		SELECT id FROM flow_user WHERE username = \$1
	`).WithArgs("receiver1").WillReturnRows(receiverRow)

	additionalBytes, _ := json.Marshal(additionalData)
	insertRow := sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now)
	mock.ExpectQuery(`
	INSERT INTO notification \(author_id, receiver_id, notification_type, is_read, additional\)
	VALUES \(\$1, \$2, \$3, \$4, \$5\)
	RETURNING id, created_at
	`).WithArgs(1, 2, "friend_request", false, additionalBytes).
		WillReturnRows(insertRow)

	result, err := repo.AddNotification(ctx, notification)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), result.ID)
	assert.Equal(t, now, result.Timestamp)
	assert.Equal(t, "avatar1.jpg", result.Avatar)
	assert.True(t, result.IsExternalAvatar)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddNotification_SenderNotFound(t *testing.T) {
	repo, mock, closeFn := setupNotificationTest(t)
	defer closeFn()

	ctx := context.Background()
	notification := domain.Notification{
		SenderUsername: "unknown",
	}

	mock.ExpectQuery(`
		SELECT id, avatar, is_external_avatar FROM flow_user WHERE username = \$1
	`).WithArgs("unknown").WillReturnError(sql.ErrNoRows)

	_, err := repo.AddNotification(ctx, notification)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get notification sender err")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddNotification_ReceiverNotFound(t *testing.T) {
	repo, mock, closeFn := setupNotificationTest(t)
	defer closeFn()

	ctx := context.Background()
	notification := domain.Notification{
		SenderUsername:   "sender1",
		ReceiverUsername: "unknown",
	}

	senderRow := sqlmock.NewRows([]string{"id", "avatar", "is_external_avatar"}).
		AddRow(1, "avatar1.jpg", true)
	mock.ExpectQuery(`
		SELECT id, avatar, is_external_avatar FROM flow_user WHERE username = \$1
	`).WithArgs("sender1").WillReturnRows(senderRow)

	mock.ExpectQuery(`
		SELECT id FROM flow_user WHERE username = \$1
	`).WithArgs("unknown").WillReturnError(sql.ErrNoRows)

	_, err := repo.AddNotification(ctx, notification)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get notification receiver err")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddNotification_InsertError(t *testing.T) {
	repo, mock, closeFn := setupNotificationTest(t)
	defer closeFn()

	ctx := context.Background()
	notification := domain.Notification{
		SenderUsername:   "sender1",
		ReceiverUsername: "receiver1",
		Type:             "friend_request",
	}

	senderRow := sqlmock.NewRows([]string{"id", "avatar", "is_external_avatar"}).
		AddRow(1, "avatar1.jpg", true)
	mock.ExpectQuery(`
		SELECT id, avatar, is_external_avatar FROM flow_user WHERE username = \$1
	`).WithArgs("sender1").WillReturnRows(senderRow)

	receiverRow := sqlmock.NewRows([]string{"id"}).AddRow(2)
	mock.ExpectQuery(`
		SELECT id FROM flow_user WHERE username = \$1
	`).WithArgs("receiver1").WillReturnRows(receiverRow)

	additionalBytes, _ := json.Marshal(notification.AdditionalData)
	mock.ExpectQuery(`
	INSERT INTO notification .*
	`).WithArgs(1, 2, "friend_request", false, additionalBytes).
		WillReturnError(errors.New("insert error"))

	_, err := repo.AddNotification(ctx, notification)
	assert.Error(t, err)
	assert.EqualError(t, err, "insert error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteNotification_Success(t *testing.T) {
	repo, mock, closeFn := setupNotificationTest(t)
	defer closeFn()

	ctx := context.Background()
	id := uint64(1)
	usernameID := uint64(2)

	mock.ExpectExec(`
	DELETE FROM notification
	WHERE id = \$1
	AND receiver_id = \$2
	`).WithArgs(id, usernameID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteNotification(ctx, id, usernameID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteNotification_NotFound(t *testing.T) {
	repo, mock, closeFn := setupNotificationTest(t)
	defer closeFn()

	ctx := context.Background()
	id := uint64(1)
	usernameID := uint64(2)

	mock.ExpectExec(`
	DELETE FROM notification
	WHERE id = \$1
	AND receiver_id = \$2
	`).WithArgs(id, usernameID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.DeleteNotification(ctx, id, usernameID)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteNotification_DBError(t *testing.T) {
	repo, mock, closeFn := setupNotificationTest(t)
	defer closeFn()

	ctx := context.Background()
	id := uint64(1)
	usernameID := uint64(2)

	mock.ExpectExec(`
	DELETE FROM notification
	WHERE id = \$1
	AND receiver_id = \$2
	`).WithArgs(id, usernameID).
		WillReturnError(errors.New("database error"))

	err := repo.DeleteNotification(ctx, id, usernameID)
	assert.Error(t, err)
	assert.EqualError(t, err, "database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}
