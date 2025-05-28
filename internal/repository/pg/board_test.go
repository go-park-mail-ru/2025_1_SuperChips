package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/stretchr/testify/assert"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return db, mock
}

func TestGetUsernameID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	storage := NewBoardStorage(db)
	ctx := context.Background()
	username := "testuser"
	userID := 123

	mock.ExpectQuery(`SELECT id FROM flow_user WHERE username = \$1`).
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))

	result, err := storage.GetUsernameID(ctx, username, userID)
	assert.NoError(t, err)
	assert.Equal(t, userID, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUsernameID_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	storage := NewBoardStorage(db)
	ctx := context.Background()
	username := "nonexistentuser"

	mock.ExpectQuery(`SELECT id FROM flow_user WHERE username = \$1`).
		WithArgs(username).
		WillReturnError(sql.ErrNoRows)

	result, err := storage.GetUsernameID(ctx, username, 0)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Zero(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateBoard_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	storage := NewBoardStorage(db)
	ctx := context.Background()
	board := &domain.Board{
		Name:      "Test Board",
		IsPrivate: true,
	}
	userID := 123

	mock.ExpectQuery(`INSERT INTO board \(author_id, board_name, is_private\) VALUES \(\$1, \$2, \$3\) ON CONFLICT \(author_id, board_name\) DO NOTHING RETURNING id`).
		WithArgs(userID, board.Name, board.IsPrivate).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := storage.CreateBoard(ctx, board, "", userID)
	assert.NoError(t, err)
	assert.Equal(t, 1, board.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateBoard_Conflict(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	storage := NewBoardStorage(db)
	ctx := context.Background()
	board := &domain.Board{
		Name:      "Test Board",
		IsPrivate: true,
	}
	userID := 123

	mock.ExpectQuery(`INSERT INTO board \(author_id, board_name, is_private\) VALUES \(\$1, \$2, \$3\) ON CONFLICT \(author_id, board_name\) DO NOTHING RETURNING id`).
		WithArgs(userID, board.Name, board.IsPrivate).
		WillReturnError(sql.ErrNoRows)

	err := storage.CreateBoard(ctx, board, "", userID)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrConflict, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteBoard_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	storage := NewBoardStorage(db)
	ctx := context.Background()
	boardID := 1
	userID := 123

	mock.ExpectQuery(`DELETE FROM board WHERE id = \$1 AND author_id = \$2 RETURNING id`).
		WithArgs(boardID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(boardID))

	err := storage.DeleteBoard(ctx, boardID, userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteBoard_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	storage := NewBoardStorage(db)
	ctx := context.Background()
	boardID := 1
	userID := 123

	mock.ExpectQuery(`DELETE FROM board WHERE id = \$1 AND author_id = \$2 RETURNING id`).
		WithArgs(boardID, userID).
		WillReturnError(sql.ErrNoRows)

	err := storage.DeleteBoard(ctx, boardID, userID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddToBoard_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	storage := NewBoardStorage(db)
	ctx := context.Background()
	boardID := 1
	userID := 123
	flowID := 456

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE board SET flow_count = flow_count \+ 1 WHERE id = \$1 AND author_id = \$2`).
		WithArgs(boardID, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`INSERT INTO board_post \(board_id, flow_id\) VALUES \(\$1, \$2\) RETURNING board_id`).
		WithArgs(boardID, flowID).
		WillReturnRows(sqlmock.NewRows([]string{"board_id"}).AddRow(boardID))
	mock.ExpectCommit()

	err := storage.AddToBoard(ctx, boardID, userID, flowID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddToBoard_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	storage := NewBoardStorage(db)
	ctx := context.Background()
	boardID := 1
	userID := 123
	flowID := 456

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE board SET flow_count = flow_count \+ 1 WHERE id = \$1 AND author_id = \$2`).
		WithArgs(boardID, userID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := storage.AddToBoard(ctx, boardID, userID, flowID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateBoard_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	storage := NewBoardStorage(db)
	ctx := context.Background()
	boardID := 1
	userID := 123
	newName := "Updated Board"
	isPrivate := true

	mock.ExpectExec(`UPDATE board SET board_name = COALESCE\(\$1, board_name\), is_private = COALESCE\(\$2, is_private\) WHERE id = \$3 AND author_id = \$4`).
		WithArgs(newName, isPrivate, boardID, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := storage.UpdateBoard(ctx, boardID, userID, newName, isPrivate)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateBoard_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	storage := NewBoardStorage(db)
	ctx := context.Background()
	boardID := 1
	userID := 123
	newName := "Updated Board"
	isPrivate := true

	mock.ExpectExec(`UPDATE board SET board_name = COALESCE\(\$1, board_name\), is_private = COALESCE\(\$2, is_private\) WHERE id = \$3 AND author_id = \$4`).
		WithArgs(newName, isPrivate, boardID, userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := storage.UpdateBoard(ctx, boardID, userID, newName, isPrivate)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBoard_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	storage := NewBoardStorage(db)
	ctx := context.Background()
	boardID := 1
	userID := 123
	previewNum := 5
	previewStart := 0

	mock.ExpectQuery(`SELECT board\.id, board\.author_id, board\.board_name, board\.created_at, board\.is_private, board\.flow_count, flow_user\.username FROM board INNER JOIN flow_user ON board\.author_id = flow_user\.id WHERE board\.id = \$1`).
		WithArgs(boardID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "author_id", "board_name", "created_at", "is_private", "flow_count", "username"}).
			AddRow(1, 123, "Test Board", time.Now(), false, 10, "testuser"))

	mock.ExpectQuery(`SELECT f\.id, f\.title, f\.description, f\.author_id, f\.created_at, f\.updated_at, f\.is_private, f\.media_url, f\.like_count, f\.width, f\.height, f\.is_nsfw FROM flow f JOIN board_post bp ON f\.id = bp\.flow_id WHERE bp\.board_id = \$1 AND \(f\.is_private = false OR f\.author_id = \$2\) ORDER BY bp\.saved_at DESC LIMIT \$3 OFFSET \$4`).
		WithArgs(boardID, userID, previewNum, previewStart).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "author_id", "created_at", "updated_at", "is_private", "media_url", "like_count", "width", "height", "is_nsfw"}).
			AddRow(1, "Flow Title", "Flow Description", 123, time.Now(), time.Now(), false, "http://example.com/media.jpg", 5, 800, 600, false))

	mock.ExpectQuery(`SELECT c\.color_hex FROM color c JOIN flow f ON c\.flow_id = f\.id JOIN board_post bp ON f\.id = bp\.flow_id WHERE bp\.board_id = \$1 AND \(f\.is_private = false OR f\.author_id = \$2\) ORDER BY bp\.saved_at DESC LIMIT \$3`).
		WithArgs(boardID, userID, 20).
		WillReturnRows(sqlmock.NewRows([]string{"color_hex"}).AddRow("#FFFFFF"))

	board, colors, err := storage.GetBoard(ctx, boardID, userID, previewNum, previewStart)
	assert.NoError(t, err)
	assert.NotNil(t, board)
	assert.NotEmpty(t, colors)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBoard_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	storage := NewBoardStorage(db)
	ctx := context.Background()
	boardID := 1
	userID := 123
	previewNum := 5
	previewStart := 0

	mock.ExpectQuery(`SELECT board\.id, board\.author_id, board\.board_name, board\.created_at, board\.is_private, board\.flow_count, flow_user\.username FROM board INNER JOIN flow_user ON board\.author_id = flow_user\.id WHERE board\.id = \$1`).
		WithArgs(boardID).
		WillReturnError(sql.ErrNoRows)

	board, colors, err := storage.GetBoard(ctx, boardID, userID, previewNum, previewStart)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Empty(t, board)
	assert.Empty(t, colors)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserPublicBoards_Success(t *testing.T) {
    db, mock := setupMockDB(t)
    defer db.Close()

    storage := NewBoardStorage(db)
    ctx := context.Background()
    username := "testuser"
    previewNum := 5
    previewStart := 0

    mock.ExpectQuery(`SELECT b\.id, b\.author_id, b\.board_name, b\.created_at, b\.is_private, b\.flow_count FROM flow_user u JOIN board b ON u\.id = b\.author_id WHERE u\.username = \$1 AND b\.is_private = false`).
        WithArgs(username).
        WillReturnRows(sqlmock.NewRows([]string{"id", "author_id", "board_name", "created_at", "is_private", "flow_count"}).
            AddRow(1, 0, "Public Board", time.Now(), false, 10))

    mock.ExpectQuery(`SELECT f\.id, f\.title, f\.description, f\.author_id, f\.created_at, f\.updated_at, f\.is_private, f\.media_url, f\.like_count, f\.width, f\.height, f\.is_nsfw FROM flow f JOIN board_post bp ON f\.id = bp\.flow_id WHERE bp\.board_id = \$1 AND \(f\.is_private = false OR f\.author_id = \$2\) ORDER BY bp\.saved_at DESC LIMIT \$3 OFFSET \$4`).
        WithArgs(1, 0, previewNum, previewStart).
        WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "author_id", "created_at", "updated_at", "is_private", "media_url", "like_count", "width", "height", "is_nsfw"}).
            AddRow(1, "Flow Title", "Flow Description", 123, time.Now(), time.Now(), false, "http://example.com/media.jpg", 5, 800, 600, false))

    boards, err := storage.GetUserPublicBoards(ctx, username, previewNum, previewStart)
    assert.NoError(t, err)
    assert.NotEmpty(t, boards)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserAllBoards_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	storage := NewBoardStorage(db)
	ctx := context.Background()
	userID := 123
	previewNum := 5
	previewStart := 0

	mock.ExpectQuery(`SELECT id, author_id, board_name, created_at, is_private, flow_count FROM board WHERE author_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "author_id", "board_name", "created_at", "is_private", "flow_count"}).
			AddRow(1, 123, "Test Board", time.Now(), false, 10))

	mock.ExpectQuery(`SELECT f\.id, f\.title, f\.description, f\.author_id, f\.created_at, f\.updated_at, f\.is_private, f\.media_url, f\.like_count, f\.width, f\.height, f\.is_nsfw FROM flow f JOIN board_post bp ON f\.id = bp\.flow_id WHERE bp\.board_id = \$1 AND \(f\.is_private = false OR f\.author_id = \$2\) ORDER BY bp\.saved_at DESC LIMIT \$3 OFFSET \$4`).
		WithArgs(1, userID, previewNum, previewStart).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "author_id", "created_at", "updated_at", "is_private", "media_url", "like_count", "width", "height", "is_nsfw"}).
			AddRow(1, "Flow Title", "Flow Description", 123, time.Now(), time.Now(), false, "http://example.com/media.jpg", 5, 800, 600, false))

	boards, err := storage.GetUserAllBoards(ctx, userID, previewNum, previewStart)
	assert.NoError(t, err)
	assert.NotEmpty(t, boards)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBoardFlow_Success(t *testing.T) {
    db, mock := setupMockDB(t)
    defer db.Close()

    storage := NewBoardStorage(db)
    ctx := context.Background()
    boardID := 1
    userID := 123
    page := 1
    pageSize := 10

    mock.ExpectQuery(`SELECT id FROM board WHERE id = \$1 AND \(is_private = false OR author_id = \$2\)`).
        WithArgs(boardID, userID).
        WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(boardID))

    mock.ExpectQuery(`SELECT f\.id, f\.title, f\.description, f\.author_id, f\.created_at, f\.updated_at, f\.is_private, f\.media_url, f\.like_count, f\.width, f\.height, f\.is_nsfw FROM flow f JOIN board_post bp ON f\.id = bp\.flow_id WHERE bp\.board_id = \$1 AND \(f\.is_private = false OR f\.author_id = \$2\) ORDER BY bp\.saved_at DESC LIMIT \$3 OFFSET \$4`).
        WithArgs(boardID, userID, pageSize, (page-1)*pageSize).
        WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "author_id", "created_at", "updated_at", "is_private", "media_url", "like_count", "width", "height", "is_nsfw"}).
            AddRow(1, "Flow Title", "Flow Description", 123, time.Now(), time.Now(), false, "http://example.com/media.jpg", 5, 800, 600, false))

    flows, err := storage.GetBoardFlow(ctx, boardID, userID, page, pageSize)
    assert.NoError(t, err)
    assert.NotEmpty(t, flows)
    assert.NoError(t, mock.ExpectationsWereMet())
}

