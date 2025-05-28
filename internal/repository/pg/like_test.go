package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/stretchr/testify/assert"
)

func setupLikeMock(t *testing.T) (sqlmock.Sqlmock, *pgLikeStorage) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	return mock, NewPgLikeStorage(db)
}

const expected = `
		WITH deleted AS \(
			DELETE FROM flow_like
			WHERE user_id = \$1 AND flow_id = \$2
			RETURNING 'delete' AS action
		\),
		inserted AS \(
			INSERT INTO flow_like \(user_id, flow_id\)
			SELECT \$1, \$2
			WHERE NOT EXISTS \(SELECT 1 FROM deleted\)
			RETURNING 'insert' AS action
		\),
		update_like_count AS \(
			UPDATE flow
			SET like_count = like_count \+ CASE
				WHEN EXISTS \(SELECT 1 FROM inserted\) THEN 1
				WHEN EXISTS \(SELECT 1 FROM deleted\) THEN -1
				ELSE 0
			END
			WHERE id = \$2
		\)
		SELECT 
			author_username, 
			COALESCE\(\(SELECT action FROM inserted\), \(SELECT action FROM deleted\)\) AS action`

func TestLikeFlow_InsertLike(t *testing.T) {
	mock, storage := setupLikeMock(t)
	defer mock.ExpectClose()

	ctx := context.Background()
	pinID := 1
	userID := 2

	mock.ExpectQuery(expected).
		WithArgs(userID, pinID).
		WillReturnRows(sqlmock.NewRows([]string{"author_username", "action"}).AddRow("test_author", "insert"))

	action, author, err := storage.LikeFlow(ctx, pinID, userID)
	assert.NoError(t, err)
	assert.Equal(t, "insert", action)
	assert.Equal(t, "test_author", author)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLikeFlow_DeleteLike(t *testing.T) {
	mock, storage := setupLikeMock(t)
	defer mock.ExpectClose()

	ctx := context.Background()
	pinID := 1
	userID := 2

	mock.ExpectQuery(expected).
		WithArgs(userID, pinID).
		WillReturnRows(sqlmock.NewRows([]string{"author_username", "action"}).AddRow("test_author", "delete"))

	action, author, err := storage.LikeFlow(ctx, pinID, userID)
	assert.NoError(t, err)
	assert.Equal(t, "delete", action)
	assert.Equal(t, "test_author", author)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLikeFlow_NoRows(t *testing.T) {
	mock, storage := setupLikeMock(t)
	defer mock.ExpectClose()

	ctx := context.Background()
	pinID := 1
	userID := 2

	mock.ExpectQuery(expected).
		WithArgs(userID, pinID).
		WillReturnError(sql.ErrNoRows)

	action, author, err := storage.LikeFlow(ctx, pinID, userID)
	assert.Equal(t, "", action)
	assert.Equal(t, "", author)
	assert.ErrorIs(t, err, domain.ErrForbidden)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLikeFlow_DBError(t *testing.T) {
	mock, storage := setupLikeMock(t)
	defer mock.ExpectClose()

	ctx := context.Background()
	pinID := 1
	userID := 2

	mock.ExpectQuery(expected).
		WithArgs(userID, pinID).
		WillReturnError(errors.New("database error"))

	action, author, err := storage.LikeFlow(ctx, pinID, userID)
	assert.Equal(t, "", action)
	assert.Equal(t, "", author)
	assert.ErrorContains(t, err, "database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

