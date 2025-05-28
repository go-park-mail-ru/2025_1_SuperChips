package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/stretchr/testify/assert"
)

func setupCommentMock(t *testing.T) (*CommentRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}

	repo := NewCommentRepository(db)
	return repo, mock, func() { db.Close() }
}

func TestGetComments_Success(t *testing.T) {
	repo, mock, closeFn := setupCommentMock(t)
	defer closeFn()

	ctx := context.Background()
	flowID := 1
	userID := 2
	page := 1
	size := 10
	offset := (page - 1) * size

	rows := sqlmock.NewRows([]string{
		"id", "author_id", "flow_id", "contents", "like_count", "created_at",
		"username", "avatar", "is_external_avatar", "is_liked",
	}).
		AddRow(1, 10, 1, "First comment", 5, time.Now(), "user1", "avatar1.jpg", true, true).
		AddRow(2, 11, 1, "Second comment", 3, time.Now(), "user2", "avatar2.jpg", false, false)

	mock.ExpectQuery(regexp.QuoteMeta(`
    SELECT 
        c.id, 
        c.author_id, 
        c.flow_id, 
        c.contents, 
        c.like_count, 
        c.created_at, 
        fu.username, 
        fu.avatar, 
        fu.is_external_avatar,
        EXISTS (
            SELECT 1 FROM comment_like cl 
            WHERE cl.comment_id = c.id AND cl.user_id = $2
        ) AS is_liked
    FROM comment c
    JOIN flow_user fu ON fu.id = c.author_id
    LEFT JOIN flow f ON f.id = c.flow_id
    WHERE c.flow_id = $1
    AND (f.is_private = false OR f.author_id = $2)
    ORDER BY c.created_at DESC
    OFFSET $3
    LIMIT $4
	`)).WithArgs(flowID, userID, offset, size).WillReturnRows(rows)

	comments, err := repo.GetComments(ctx, flowID, userID, page, size)
	assert.NoError(t, err)
	assert.Len(t, comments, 2)

	assert.Equal(t, 1, comments[0].ID)
	assert.Equal(t, 10, comments[0].AuthorID)
	assert.Equal(t, "First comment", comments[0].Content)
	assert.Equal(t, 5, comments[0].LikeCount)
	assert.Equal(t, "user1", comments[0].AuthorUsername)
	assert.Equal(t, "avatar1.jpg", comments[0].AuthorAvatar)
	assert.True(t, comments[0].AuthorIsExternalAvatar)
	assert.True(t, comments[0].IsLiked)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetComments_EmptyResult(t *testing.T) {
	repo, mock, closeFn := setupCommentMock(t)
	defer closeFn()

	ctx := context.Background()
	flowID := 1
	userID := 2
	page := 1
	size := 10
	offset := (page - 1) * size

	mock.ExpectQuery(regexp.QuoteMeta(`
    SELECT 
        c.id, 
        c.author_id, 
        c.flow_id, 
        c.contents, 
        c.like_count, 
        c.created_at, 
        fu.username, 
        fu.avatar, 
        fu.is_external_avatar,
        EXISTS (
            SELECT 1 FROM comment_like cl 
            WHERE cl.comment_id = c.id AND cl.user_id = $2
        ) AS is_liked
    FROM comment c
    JOIN flow_user fu ON fu.id = c.author_id
    LEFT JOIN flow f ON f.id = c.flow_id
    WHERE c.flow_id = $1
    AND (f.is_private = false OR f.author_id = $2)
    ORDER BY c.created_at DESC
    OFFSET $3
    LIMIT $4
	`)).WithArgs(flowID, userID, offset, size).WillReturnRows(sqlmock.NewRows([]string{
		"id", "author_id", "flow_id", "contents", "like_count", "created_at",
		"username", "avatar", "is_external_avatar", "is_liked",
	}))

	comments, err := repo.GetComments(ctx, flowID, userID, page, size)
	assert.NoError(t, err)
	assert.Empty(t, comments)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetComments_DatabaseError(t *testing.T) {
	repo, mock, closeFn := setupCommentMock(t)
	defer closeFn()

	ctx := context.Background()
	flowID := 1
	userID := 2
	page := 1
	size := 10
	offset := (page - 1) * size

	mock.ExpectQuery(regexp.QuoteMeta(`
    SELECT 
        c.id, 
        c.author_id, 
        c.flow_id, 
        c.contents, 
        c.like_count, 
        c.created_at, 
        fu.username, 
        fu.avatar, 
        fu.is_external_avatar,
        EXISTS (
            SELECT 1 FROM comment_like cl 
            WHERE cl.comment_id = c.id AND cl.user_id = $2
        ) AS is_liked
    FROM comment c
    JOIN flow_user fu ON fu.id = c.author_id
    LEFT JOIN flow f ON f.id = c.flow_id
    WHERE c.flow_id = $1
    AND (f.is_private = false OR f.author_id = $2)
    ORDER BY c.created_at DESC
    OFFSET $3
    LIMIT $4
	`)).WithArgs(flowID, userID, offset, size).WillReturnError(errors.New("database error"))

	comments, err := repo.GetComments(ctx, flowID, userID, page, size)
	assert.Error(t, err)
	assert.Nil(t, comments)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLikeComment_InsertLike(t *testing.T) {
	repo, mock, closeFn := setupCommentMock(t)
	defer closeFn()

	ctx := context.Background()
	commentID := 1
	userID := 2

	mock.ExpectQuery(regexp.QuoteMeta(`
	WITH deleted AS (
		DELETE FROM comment_like
		WHERE user_id = $1 AND comment_id = $2
		RETURNING 'delete' AS action
	),
	inserted AS (
		INSERT INTO comment_like (user_id, comment_id)
		SELECT $1, $2
		WHERE NOT EXISTS (SELECT 1 FROM deleted)
		RETURNING 'insert' AS action
	),
	update_like_count AS (
		UPDATE comment
		SET like_count = like_count + CASE
			WHEN EXISTS (SELECT 1 FROM inserted) THEN 1
			WHEN EXISTS (SELECT 1 FROM deleted) THEN -1
			ELSE 0
		END
		WHERE id = $2
	)
	SELECT COALESCE((SELECT action FROM inserted), (SELECT action FROM deleted)) AS action
    `)).WithArgs(userID, commentID).WillReturnRows(sqlmock.NewRows([]string{"action"}).AddRow("insert"))

	action, err := repo.LikeComment(ctx, commentID, userID)
	assert.NoError(t, err)
	assert.Equal(t, "insert", action)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLikeComment_DeleteLike(t *testing.T) {
	repo, mock, closeFn := setupCommentMock(t)
	defer closeFn()

	ctx := context.Background()
	commentID := 1
	userID := 2

	mock.ExpectQuery(regexp.QuoteMeta(`
	WITH deleted AS (
		DELETE FROM comment_like
		WHERE user_id = $1 AND comment_id = $2
		RETURNING 'delete' AS action
	),
	inserted AS (
		INSERT INTO comment_like (user_id, comment_id)
		SELECT $1, $2
		WHERE NOT EXISTS (SELECT 1 FROM deleted)
		RETURNING 'insert' AS action
	),
	update_like_count AS (
		UPDATE comment
		SET like_count = like_count + CASE
			WHEN EXISTS (SELECT 1 FROM inserted) THEN 1
			WHEN EXISTS (SELECT 1 FROM deleted) THEN -1
			ELSE 0
		END
		WHERE id = $2
	)
	SELECT COALESCE((SELECT action FROM inserted), (SELECT action FROM deleted)) AS action
    `)).WithArgs(userID, commentID).WillReturnRows(sqlmock.NewRows([]string{"action"}).AddRow("delete"))

	action, err := repo.LikeComment(ctx, commentID, userID)
	assert.NoError(t, err)
	assert.Equal(t, "delete", action)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLikeComment_Forbidden(t *testing.T) {
	repo, mock, closeFn := setupCommentMock(t)
	defer closeFn()

	ctx := context.Background()
	commentID := 1
	userID := 2

	mock.ExpectQuery(regexp.QuoteMeta(`
	WITH deleted AS (
		DELETE FROM comment_like
		WHERE user_id = $1 AND comment_id = $2
		RETURNING 'delete' AS action
	),
	inserted AS (
		INSERT INTO comment_like (user_id, comment_id)
		SELECT $1, $2
		WHERE NOT EXISTS (SELECT 1 FROM deleted)
		RETURNING 'insert' AS action
	),
	update_like_count AS (
		UPDATE comment
		SET like_count = like_count + CASE
			WHEN EXISTS (SELECT 1 FROM inserted) THEN 1
			WHEN EXISTS (SELECT 1 FROM deleted) THEN -1
			ELSE 0
		END
		WHERE id = $2
	)
	SELECT COALESCE((SELECT action FROM inserted), (SELECT action FROM deleted)) AS action
    `)).WithArgs(userID, commentID).WillReturnError(sql.ErrNoRows)

	_, err := repo.LikeComment(ctx, commentID, userID)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrForbidden, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddComment_Success(t *testing.T) {
	repo, mock, closeFn := setupCommentMock(t)
	defer closeFn()

	ctx := context.Background()
	flowID := 1
	userID := 2
	content := "Test comment"

	mock.ExpectQuery(regexp.QuoteMeta(`
	INSERT INTO comment (author_id, flow_id, contents)
	SELECT $1, $2, $3
	RETURNING id;
	`)).WithArgs(userID, flowID, content).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := repo.AddComment(ctx, flowID, userID, content)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddComment_Forbidden(t *testing.T) {
	repo, mock, closeFn := setupCommentMock(t)
	defer closeFn()

	ctx := context.Background()
	flowID := 1
	userID := 2
	content := "Test comment"

	mock.ExpectQuery(regexp.QuoteMeta(`
	INSERT INTO comment (author_id, flow_id, contents)
	SELECT $1, $2, $3
	RETURNING id;
	`)).WithArgs(userID, flowID, content).WillReturnError(sql.ErrNoRows)

	err := repo.AddComment(ctx, flowID, userID, content)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrForbidden, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteComment_Success(t *testing.T) {
	repo, mock, closeFn := setupCommentMock(t)
	defer closeFn()

	ctx := context.Background()
	commentID := 1
	userID := 2

	mock.ExpectQuery(regexp.QuoteMeta(`
	DELETE FROM comment
	WHERE id = $1 AND author_id = $2
	RETURNING id
	`)).WithArgs(commentID, userID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := repo.DeleteComment(ctx, commentID, userID)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteComment_Forbidden(t *testing.T) {
	repo, mock, closeFn := setupCommentMock(t)
	defer closeFn()

	ctx := context.Background()
	commentID := 1
	userID := 2

	mock.ExpectQuery(regexp.QuoteMeta(`
	DELETE FROM comment
	WHERE id = $1 AND author_id = $2
	RETURNING id
	`)).WithArgs(commentID, userID).WillReturnError(sql.ErrNoRows)

	err := repo.DeleteComment(ctx, commentID, userID)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrForbidden, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}