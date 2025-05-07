package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSubscriptionStorage(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    repo := NewSubscriptionStorage(db)

    t.Run("GetUserFollowers", func(t *testing.T) { testGetUserFollowers(t, repo, mock) })
    t.Run("GetUserFollowing", func(t *testing.T) { testGetUserFollowing(t, repo, mock) })
    t.Run("CreateSubscription", func(t *testing.T) { testCreateSubscription(t, repo, mock) })
    t.Run("DeleteSubscription", func(t *testing.T) { testDeleteSubscription(t, repo, mock) })
}

func testGetUserFollowers(t *testing.T, repo *SubscriptionStorage, mock sqlmock.Sqlmock) {
    ctx := context.Background()

    id := 1
    page := 2
    size := 10
    offset := (page - 1) * size

    mock.ExpectQuery(regexp.QuoteMeta(
        `SELECT u.username, u.avatar, u.birthday, u.about, u.public_name, u.subscriber_count, u.is_external_avatar
		FROM subscription
		LEFT JOIN flow_user u ON
		subscription.user_id = u.id 
		WHERE subscription.target_id = $1 
		ORDER BY
		CASE WHEN subscription.created_at IS NULL THEN 1 ELSE 0 END,
		subscription.created_at DESC OFFSET $2 LIMIT $3`)).WithArgs(id, offset, size).
        WillReturnRows(sqlmock.NewRows([]string{"username", "avatar", "birthday", "about", "public_name", "subscriber_count", "is_external_avatar"}).
            AddRow("user1", "avatar1", time.Now(), "about1", "Public Name 1", int64(100), true).
            AddRow("user2", "avatar2", time.Now(), "about2", "Public Name 2", int64(200), false))

    users, err := repo.GetUserFollowers(ctx, id, page, size)
    assert.NoError(t, err)
    assert.Len(t, users, 2)

    assert.Equal(t, "user1", users[0].Username)
    assert.Equal(t, "avatar1", users[0].Avatar)
    assert.Equal(t, "Public Name 1", users[0].PublicName)
    assert.Equal(t, 100, users[0].SubscriberCount)
    assert.True(t, users[0].IsExternalAvatar)

    assert.Equal(t, "user2", users[1].Username)
    assert.Equal(t, "avatar2", users[1].Avatar)
    assert.Equal(t, "Public Name 2", users[1].PublicName)
    assert.Equal(t, 200, users[1].SubscriberCount)
    assert.False(t, users[1].IsExternalAvatar)

    assert.NoError(t, mock.ExpectationsWereMet())
}

func testGetUserFollowing(t *testing.T, repo *SubscriptionStorage, mock sqlmock.Sqlmock) {
    ctx := context.Background()

    id := 1
    page := 2
    size := 10
    offset := (page - 1) * size

    mock.ExpectQuery(regexp.QuoteMeta(
        `SELECT u.username, u.avatar, u.birthday, u.about, u.public_name, u.subscriber_count, u.is_external_avatar
		FROM subscription 
		LEFT JOIN flow_user u 
		ON subscription.target_id = u.id 
		WHERE subscription.user_id = $1 
		ORDER BY CASE WHEN subscription.created_at IS NULL THEN 1 ELSE 0 END, subscription.created_at DESC OFFSET $2 LIMIT $3`,)).WithArgs(id, offset, size).
        WillReturnRows(sqlmock.NewRows([]string{"username", "avatar", "birthday", "about", "public_name", "subscriber_count", "is_external_avatar"}).
            AddRow("user1", "avatar1", time.Now(), "about1", "Public Name 1", int64(100), true).
            AddRow("user2", "avatar2", time.Now(), "about2", "Public Name 2", int64(200), false))

    users, err := repo.GetUserFollowing(ctx, id, page, size)
    assert.NoError(t, err)
    assert.Len(t, users, 2)

    assert.Equal(t, "user1", users[0].Username)
    assert.Equal(t, "avatar1", users[0].Avatar)
    assert.Equal(t, "Public Name 1", users[0].PublicName)
    assert.Equal(t, 100, users[0].SubscriberCount)
    assert.True(t, users[0].IsExternalAvatar)

    assert.Equal(t, "user2", users[1].Username)
    assert.Equal(t, "avatar2", users[1].Avatar)
    assert.Equal(t, "Public Name 2", users[1].PublicName)
    assert.Equal(t, 200, users[1].SubscriberCount)
    assert.False(t, users[1].IsExternalAvatar)

    assert.NoError(t, mock.ExpectationsWereMet())
}

func testCreateSubscription(t *testing.T, repo *SubscriptionStorage, mock sqlmock.Sqlmock) {
    ctx := context.Background()

    targetUsername := "target_user"
    currentID := 1

	mock.ExpectBegin()

    mock.ExpectQuery(regexp.QuoteMeta(
        `SELECT id FROM flow_user WHERE username = $1`,
	)).WithArgs(targetUsername).
        WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

    mock.ExpectExec(regexp.QuoteMeta(
        `INSERT INTO subscription (user_id, target_id) VALUES ($1, $2) ON CONFLICT (user_id, target_id) DO NOTHING`,
    )).WithArgs(currentID, 2).
        WillReturnResult(sqlmock.NewResult(1, 1))

    mock.ExpectExec(regexp.QuoteMeta(
        `UPDATE flow_user SET subscriber_count = subscriber_count + 1 WHERE id = $1`,
    )).WithArgs(2).
        WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

    err := repo.CreateSubscription(ctx, targetUsername, currentID)
    assert.NoError(t, err)

    assert.NoError(t, mock.ExpectationsWereMet())
}

func testDeleteSubscription(t *testing.T, repo *SubscriptionStorage, mock sqlmock.Sqlmock) {
    ctx := context.Background()

    targetUsername := "target_user"
    currentID := 1

	mock.ExpectBegin()
    mock.ExpectQuery(regexp.QuoteMeta(
        `SELECT id FROM flow_user WHERE username = $1`,
	)).WithArgs(targetUsername).
        WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

    mock.ExpectExec(regexp.QuoteMeta(
        `DELETE FROM subscription WHERE user_id = $1 AND target_id = $2`,
	)).WithArgs(currentID, 2).
        WillReturnResult(sqlmock.NewResult(1, 1))

    mock.ExpectExec(regexp.QuoteMeta(
        `UPDATE flow_user SET subscriber_count = subscriber_count - 1 WHERE id = $1`,
    )).WithArgs(2).
        WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

    err := repo.DeleteSubscription(ctx, targetUsername, currentID)
    assert.NoError(t, err)

    assert.NoError(t, mock.ExpectationsWereMet())
}

