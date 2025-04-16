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

    storage := NewPgLikeStorage(db)
    return mock, storage
}

func TestLikeFlow_SuccessfulLike(t *testing.T) {
    mock, storage := setupLikeMock(t)
    defer mock.ExpectClose()

    ctx := context.Background()
    pinID := 1
    userID := 2

    mock.ExpectQuery(`WITH access_check AS \(.*\)`).
        WithArgs(userID, pinID).
        WillReturnRows(sqlmock.NewRows([]string{"action"}).AddRow("insert"))

    action, err := storage.LikeFlow(ctx, pinID, userID)
    assert.NoError(t, err)
    assert.Equal(t, "insert", action)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLikeFlow_SuccessfulUnlike(t *testing.T) {
    mock, storage := setupLikeMock(t)
    defer mock.ExpectClose()

    ctx := context.Background()
    pinID := 1
    userID := 2

    mock.ExpectQuery(`WITH access_check AS \(.*\)`).
        WithArgs(userID, pinID).
        WillReturnRows(sqlmock.NewRows([]string{"action"}).AddRow("delete"))

    action, err := storage.LikeFlow(ctx, pinID, userID)
    assert.NoError(t, err)
    assert.Equal(t, "delete", action)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLikeFlow_AccessDenied(t *testing.T) {
    mock, storage := setupLikeMock(t)
    defer mock.ExpectClose()

    ctx := context.Background()
    pinID := 1
    userID := 2

    mock.ExpectQuery(`WITH access_check AS \(.*\)`).
        WithArgs(userID, pinID).
        WillReturnError(sql.ErrNoRows)

    action, err := storage.LikeFlow(ctx, pinID, userID)
    assert.Equal(t, "", action)
    assert.Equal(t, domain.ErrForbidden, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLikeFlow_DatabaseError(t *testing.T) {
    mock, storage := setupLikeMock(t)
    defer mock.ExpectClose()

    ctx := context.Background()
    pinID := 1
    userID := 2

    mock.ExpectQuery(`WITH access_check AS \(.*\)`).
        WithArgs(userID, pinID).
        WillReturnError(errors.New("database error"))

    action, err := storage.LikeFlow(ctx, pinID, userID)
    assert.Equal(t, "", action)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "database error")
    assert.NoError(t, mock.ExpectationsWereMet())
}