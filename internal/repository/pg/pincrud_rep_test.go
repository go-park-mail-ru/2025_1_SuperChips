package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	pincrudService "github.com/go-park-mail-ru/2025_1_SuperChips/pincrud"
	"github.com/stretchr/testify/assert"
)

func setupPinMock(t *testing.T) (sqlmock.Sqlmock, *pgPinStorage) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}

	storage := &pgPinStorage{
		db: db,
	}

	return mock, storage
}

func TestGetPin_Success(t *testing.T) {
    mock, storage := setupPinMock(t)
    defer mock.ExpectClose()

    ctx := context.Background()
    pinID := uint64(1)
    userID := uint64(2)

    mock.ExpectQuery(
        `SELECT f\.id, f\.title, f\.description, f\.author_id, f\.is_private, f\.media_url, fu\.username, f\.like_count, f\.width, f\.height, f\.is_nsfw, CASE WHEN fl\.user_id IS NOT NULL THEN true ELSE false END AS is_liked ` +
            `FROM flow f JOIN flow_user fu ON f\.author_id = fu\.id LEFT JOIN flow_like fl ON fl\.flow_id = f\.id AND fl\.user_id = \$2 WHERE f\.id = \$1;`,
    ).WithArgs(pinID, userID).
        WillReturnRows(sqlmock.NewRows([]string{
            "id", "title", "description", "author_id", "is_private", "media_url",
            "username", "like_count", "width", "height", "is_nsfw", "is_liked",
        }).AddRow(1, "Test Title", "Test Description", 2, false, "media.jpg", "user1", 10, 400, 400, false, true))

    pin, authorID, err := storage.GetPin(ctx, pinID, userID)

    assert.NoError(t, err)
    assert.Equal(t, uint64(2), authorID)
    assert.Equal(t, "Test Title", pin.Header)
    assert.Equal(t, "user1", pin.AuthorUsername)
    assert.True(t, pin.IsLiked)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPin_NotFound(t *testing.T) {
	mock, storage := setupPinMock(t)
	defer mock.ExpectClose()

	ctx := context.Background()
	pinID := uint64(999)
	userID := uint64(2)

	mock.ExpectQuery("SELECT .* FROM flow f").
		WithArgs(pinID, userID).
		WillReturnError(sql.ErrNoRows)

	_, _, err := storage.GetPin(ctx, pinID, userID)
	assert.Equal(t, pincrudService.ErrPinNotFound, err)
}

func TestGetPinCleanMediaURL_Success(t *testing.T) {
	mock, storage := setupPinMock(t)
	defer mock.ExpectClose()

	ctx := context.Background()
	pinID := uint64(1)

	rows := sqlmock.NewRows([]string{"media_url", "author_id"}).
		AddRow("media.jpg", 2)

	mock.ExpectQuery("SELECT f.media_url, f.author_id FROM flow f").
		WithArgs(pinID).
		WillReturnRows(rows)

	url, authorID, err := storage.GetPinCleanMediaURL(ctx, pinID)
	assert.NoError(t, err)
	assert.Equal(t, "media.jpg", url)
	assert.Equal(t, uint64(2), authorID)
}

func TestDeletePin_Success(t *testing.T) {
	mock, storage := setupPinMock(t)
	defer mock.ExpectClose()

	ctx := context.Background()
	pinID := uint64(1)
	userID := uint64(2)

	mock.ExpectExec("DELETE FROM flow").
		WithArgs(pinID, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := storage.DeletePin(ctx, pinID, userID)
	assert.NoError(t, err)
}

func TestDeletePin_Unauthorized(t *testing.T) {
	mock, storage := setupPinMock(t)
	defer mock.ExpectClose()

	ctx := context.Background()
	pinID := uint64(1)
	userID := uint64(999)

	mock.ExpectExec("DELETE FROM flow").
		WithArgs(pinID, userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := storage.DeletePin(ctx, pinID, userID)
	assert.Equal(t, pincrudService.ErrUntracked, err)
}

func TestUpdatePin_Success(t *testing.T) {
	mock, storage := setupPinMock(t)
	defer mock.ExpectClose()

	var num uint64 = 1

	ctx := context.Background()
	patch := domain.PinDataUpdate{
		FlowID:      &num,
		Header:      ptrString("New Title"),
		Description: ptrString("New Description"),
		IsPrivate:   ptrBool(true),
	}
	userID := uint64(2)

	mock.ExpectExec("UPDATE flow SET title = \\$1, description = \\$2, is_private = \\$3 WHERE id = \\$4 AND author_id = \\$5").
		WithArgs("New Title", "New Description", true, uint64(1), userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := storage.UpdatePin(ctx, patch, userID)
	assert.NoError(t, err)
}

func TestUpdatePin_NoFields(t *testing.T) {
	mock, storage := setupPinMock(t)
	defer mock.ExpectClose()

	ctx := context.Background()

	var num uint64 = 1

	patch := domain.PinDataUpdate{
		FlowID: &num,
	}
	userID := uint64(2)

	err := storage.UpdatePin(ctx, patch, userID)
	assert.Equal(t, pincrudService.ErrNoFieldsToUpdate, err)
}

func TestCreatePin_Success(t *testing.T) {
    mock, storage := setupPinMock(t)
    defer mock.ExpectClose()

    ctx := context.Background()
    data := domain.PinDataCreate{
        Header:      "Test Pin",
        Description: "Test Description",
        IsPrivate:   false,
        Width:       400,
        Height:      400,
        Colors:      []string{"#FF0000", "#00FF00"},
    }
    imgName := "test.jpg"
    userID := uint64(2)

    mock.ExpectBegin()

    mock.ExpectQuery("INSERT INTO flow").
        WithArgs("Test Pin", "Test Description", userID, false, imgName, 400, 400).
        WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

    for _, color := range data.Colors {
        mock.ExpectExec("INSERT INTO color").
            WithArgs(1, color).
            WillReturnResult(sqlmock.NewResult(1, 1))
    }

    mock.ExpectCommit()

    pinID, err := storage.CreatePin(ctx, data, imgName, userID)
    assert.NoError(t, err)
    assert.Equal(t, uint64(1), pinID)

    assert.NoError(t, mock.ExpectationsWereMet())
}

func ptrString(s string) *string {
	return &s
}

func ptrBool(b bool) *bool {
	return &b
}
