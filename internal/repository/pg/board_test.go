package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	boardService "github.com/go-park-mail-ru/2025_1_SuperChips/board"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/stretchr/testify/assert"
)

func setupMockDB(t *testing.T) (*pgBoardStorage, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}

	storage := NewBoardStorage(db)
	return storage, mock, func() { db.Close() }
}

func TestGetUsernameID_Success(t *testing.T) {
	storage, mock, closeFn := setupMockDB(t)
	defer closeFn()

	username := "testuser"
	expectedID := 1

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id 
        FROM flow_user 
        WHERE username = $1
    `)).
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedID))

	id, err := storage.GetUsernameID(context.Background(), username, 0)
	assert.NoError(t, err)
	assert.Equal(t, expectedID, id)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUsernameID_NotFound(t *testing.T) {
	storage, mock, closeFn := setupMockDB(t)
	defer closeFn()

	username := "nonexistent"

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id 
        FROM flow_user 
        WHERE username = $1
    `)).
		WithArgs(username).
		WillReturnError(sql.ErrNoRows)

	id, err := storage.GetUsernameID(context.Background(), username, 0)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Equal(t, 0, id)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateBoard_Success(t *testing.T) {
	storage, mock, closeFn := setupMockDB(t)
	defer closeFn()

	testBoard := &domain.Board{
		Name:      "Test Board",
		IsPrivate: false,
	}
	userID := 1

	mock.ExpectQuery(regexp.QuoteMeta(`
        INSERT INTO board (author_id, board_name, is_private)
        VALUES ($1, $2, $3)
        ON CONFLICT (author_id, board_name) DO NOTHING
        RETURNING id
    `)).
		WithArgs(userID, testBoard.Name, testBoard.IsPrivate).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(100))

	err := storage.CreateBoard(context.Background(), testBoard, "testuser", userID)
	assert.NoError(t, err)
	assert.Equal(t, 100, testBoard.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateBoard_Conflict(t *testing.T) {
	storage, mock, closeFn := setupMockDB(t)
	defer closeFn()

	testBoard := &domain.Board{
		Name:      "Conflict Board",
		IsPrivate: false,
	}
	userID := 1

	mock.ExpectQuery(regexp.QuoteMeta(`
        INSERT INTO board (author_id, board_name, is_private)
        VALUES ($1, $2, $3)
        ON CONFLICT (author_id, board_name) DO NOTHING
        RETURNING id
    `)).
		WithArgs(userID, testBoard.Name, testBoard.IsPrivate).
		WillReturnError(sql.ErrNoRows)

	err := storage.CreateBoard(context.Background(), testBoard, "testuser", userID)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrConflict, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteBoard_Success(t *testing.T) {
	storage, mock, closeFn := setupMockDB(t)
	defer closeFn()

	boardID := 101
	userID := 1

	mock.ExpectQuery(regexp.QuoteMeta(`
	DELETE FROM board 
	WHERE id = $1
	AND
	author_id = $2
	RETURNING id`)).
		WithArgs(boardID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(boardID))

	err := storage.DeleteBoard(context.Background(), boardID, userID)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteBoard_NotFound(t *testing.T) {
	storage, mock, closeFn := setupMockDB(t)
	defer closeFn()

	boardID := 999
	userID := 1

	mock.ExpectQuery(regexp.QuoteMeta(`
	DELETE FROM board 
	WHERE id = $1
	AND
	author_id = $2
	RETURNING id`)).
		WithArgs(boardID, userID).
		WillReturnError(sql.ErrNoRows)

	err := storage.DeleteBoard(context.Background(), boardID, userID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddToBoard_Success(t *testing.T) {
	storage, mock, closeFn := setupMockDB(t)
	defer closeFn()

	boardID, userID, flowID := 200, 1, 300

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta(`
        UPDATE board
        SET flow_count = flow_count + 1
        WHERE id = $1 AND author_id = $2
    `)).
		WithArgs(boardID, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`
        INSERT INTO board_post (board_id, flow_id)
        VALUES ($1, $2)
        RETURNING board_id
    `)).
		WithArgs(boardID, flowID).
		WillReturnRows(sqlmock.NewRows([]string{"board_id"}).AddRow(boardID))

	mock.ExpectCommit()

	err := storage.AddToBoard(context.Background(), boardID, userID, flowID)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteFromBoard_Success(t *testing.T) {
	storage, mock, closeFn := setupMockDB(t)
	defer closeFn()

	boardID, userID, flowID := 200, 1, 300

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta(`
        DELETE FROM board_post
        WHERE board_id = $1
        AND flow_id = $3
        AND EXISTS (
            SELECT 1 FROM board
            WHERE board.id = $1
            AND board.author_id = $2
			FOR UPDATE
        )
    `)).
		WithArgs(boardID, userID, flowID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(regexp.QuoteMeta(`
        UPDATE board
        SET flow_count = flow_count - 1
        WHERE id = $1
        AND flow_count > 0
    `)).
		WithArgs(boardID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err := storage.DeleteFromBoard(context.Background(), boardID, userID, flowID)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateBoard_Success(t *testing.T) {
	storage, mock, closeFn := setupMockDB(t)
	defer closeFn()

	boardID := 10
	userID := 1
	newName := "Updated Board"
	isPrivate := true

	mock.ExpectExec(regexp.QuoteMeta(`
        UPDATE board
        SET board_name = COALESCE($1, board_name),
            is_private = COALESCE($2, is_private)
        WHERE id = $3 AND author_id = $4
    `)).
		WithArgs(newName, isPrivate, boardID, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := storage.UpdateBoard(context.Background(), boardID, userID, newName, isPrivate)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateBoard_NotFound(t *testing.T) {
	storage, mock, closeFn := setupMockDB(t)
	defer closeFn()

	boardID := 999
	userID := 1
	newName := "Name"
	isPrivate := false

	mock.ExpectExec(regexp.QuoteMeta(`
        UPDATE board
        SET board_name = COALESCE($1, board_name),
            is_private = COALESCE($2, is_private)
        WHERE id = $3 AND author_id = $4
    `)).
		WithArgs(newName, isPrivate, boardID, userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := storage.UpdateBoard(context.Background(), boardID, userID, newName, isPrivate)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBoardFlow_Success(t *testing.T) {
	storage, mock, closeFn := setupMockDB(t)
	defer closeFn()

	boardID := 50
	userID := 1
	page, pageSize := 1, 2
	offset := 0

	mock.ExpectQuery(regexp.QuoteMeta(`
	SELECT id
	FROM board
	WHERE id = $1 AND (is_private = false
	OR author_id = $2)`)).
		WithArgs(boardID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(boardID))

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT f.id, f.title, f.description, f.author_id, f.created_at, 
               f.updated_at, f.is_private, f.media_url, f.like_count
        FROM flow f
        JOIN board_post bp ON f.id = bp.flow_id
        WHERE bp.board_id = $1
          AND (f.is_private = false OR f.author_id = $2)
        ORDER BY bp.saved_at DESC
        LIMIT $3 OFFSET $4
    `)).
		WithArgs(boardID, userID, pageSize, offset).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "author_id", "created_at", "updated_at",
			"is_private", "media_url", "like_count",
		}).
			AddRow(1, "First Flow", "Description 1", 2, time.Now(), time.Now(), false, "url1", 10).
			AddRow(2, "Second Flow", "Description 2", 3, time.Now(), time.Now(), false, "url2", 5))

	flows, err := storage.GetBoardFlow(context.Background(), boardID, userID, page, pageSize)
	assert.NoError(t, err)
	assert.Len(t, flows, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBoardFlow_Forbidden(t *testing.T) {
	storage, mock, closeFn := setupMockDB(t)
	defer closeFn()

	boardID := 50
	userID := 1

	mock.ExpectQuery(regexp.QuoteMeta(`
	SELECT id
	FROM board
	WHERE id = $1 AND (is_private = false
	OR author_id = $2)`)).
		WithArgs(boardID, userID).
		WillReturnError(sql.ErrNoRows)

	flows, err := storage.GetBoardFlow(context.Background(), boardID, userID, 1, 2)
	assert.Error(t, err)
	assert.Nil(t, flows)
	assert.True(t, errors.Is(err, boardService.ErrForbidden))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func createFakeFlowRows() *sqlmock.Rows {
	cols := []string{
		"id", "title", "description", "author_id", "created_at", "updated_at",
		"is_private", "media_url", "like_count",
	}
	now := time.Now()
	rows := sqlmock.NewRows(cols).
		AddRow(1, "Flow Title 1", "Flow Desc 1", 10, now, now, false, "media1", 5).
		AddRow(2, "Flow Title 2", "Flow Desc 2", 11, now, now, false, "media2", 3)
	return rows
}

func TestGetBoard_Success(t *testing.T) {
	storage, mock, closeFn := setupMockDB(t)
	defer closeFn()

	boardID := 42
	userID := 99
	previewNum := 2
	previewStart := 0

	boardQuery := regexp.QuoteMeta(`
	SELECT 
		board.id, 
		board.author_id, 
		board.board_name, 
		board.created_at, 
		board.is_private, 
		board.flow_count,
		flow_user.username
	FROM
		board
	INNER JOIN 
		flow_user
	ON 
		board.author_id = flow_user.id
	WHERE 
    board.id = $1`)
	now := time.Now()
	mock.ExpectQuery(boardQuery).
		WithArgs(boardID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "author_id", "board_name", "created_at", "is_private", "flow_count", "username",
		}).AddRow(boardID, 55, "Test Board", now, false, 10, "author_user"))

	flowsQuery := regexp.QuoteMeta(`
        SELECT f.id, f.title, f.description, f.author_id, f.created_at, 
               f.updated_at, f.is_private, f.media_url, f.like_count
        FROM flow f
        JOIN board_post bp ON f.id = bp.flow_id
        WHERE bp.board_id = $1
          AND (f.is_private = false OR f.author_id = $2)
        ORDER BY bp.saved_at DESC
        LIMIT $3 OFFSET $4
    `)
	mock.ExpectQuery(flowsQuery).
		WithArgs(boardID, userID, previewNum, previewStart).
		WillReturnRows(createFakeFlowRows())

	board, err := storage.GetBoard(context.Background(), boardID, userID, previewNum, previewStart)
	assert.NoError(t, err)
	assert.Equal(t, boardID, board.ID)
	assert.Equal(t, "Test Board", board.Name)
	assert.Len(t, board.Preview, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBoard_BoardNotFound(t *testing.T) {
	storage, mock, closeFn := setupMockDB(t)
	defer closeFn()

	boardID := 123
	userID := 1
	previewNum := 1
	previewStart := 0

	boardQuery := regexp.QuoteMeta(`
	SELECT 
		board.id, 
		board.author_id, 
		board.board_name, 
		board.created_at, 
		board.is_private, 
		board.flow_count,
		flow_user.username
	FROM
		board
	INNER JOIN 
		flow_user
	ON 
		board.author_id = flow_user.id
	WHERE 
    board.id = $1`)
	mock.ExpectQuery(boardQuery).
		WithArgs(boardID).
		WillReturnError(sql.ErrNoRows)

	board, err := storage.GetBoard(context.Background(), boardID, userID, previewNum, previewStart)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound))
	assert.Equal(t, domain.Board{}, board)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserPublicBoards_Success(t *testing.T) {
	storage, mock, closeFn := setupMockDB(t)
	defer closeFn()

	username := "publicUser"
	previewNum := 2
	previewStart := 0

	boardsQuery := regexp.QuoteMeta(`
    SELECT b.id, b.author_id, b.board_name, b.created_at, b.is_private, b.flow_count
    FROM flow_user u
    JOIN board b ON u.id = b.author_id
    WHERE u.username = $1 
    AND b.is_private = false
	`)
	now := time.Now()
	boardRows := sqlmock.NewRows([]string{
		"id", "author_id", "board_name", "created_at", "is_private", "flow_count",
	}).
		AddRow(1, 55, "Board 1", now, false, 5).
		AddRow(2, 66, "Board 2", now, false, 8)
	mock.ExpectQuery(boardsQuery).
		WithArgs(username).
		WillReturnRows(boardRows)

	flowsQuery := regexp.QuoteMeta(`
        SELECT f.id, f.title, f.description, f.author_id, f.created_at, 
               f.updated_at, f.is_private, f.media_url, f.like_count
        FROM flow f
        JOIN board_post bp ON f.id = bp.flow_id
        WHERE bp.board_id = $1
          AND (f.is_private = false OR f.author_id = $2)
        ORDER BY bp.saved_at DESC
        LIMIT $3 OFFSET $4
    `)
	mock.ExpectQuery(flowsQuery).
		WithArgs(1, 0, previewNum, previewStart).
		WillReturnRows(createFakeFlowRows())
	mock.ExpectQuery(flowsQuery).
		WithArgs(2, 0, previewNum, previewStart).
		WillReturnRows(createFakeFlowRows())

	boards, err := storage.GetUserPublicBoards(context.Background(), username, previewNum, previewStart)
	assert.NoError(t, err)
	assert.Len(t, boards, 2)
	for _, b := range boards {
		assert.Len(t, b.Preview, 2)
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserAllBoards_Success(t *testing.T) {
	storage, mock, closeFn := setupMockDB(t)
	defer closeFn()

	userID := 77
	previewNum := 3
	previewStart := 0

	boardsQuery := regexp.QuoteMeta(`
        SELECT id, author_id, board_name, created_at, is_private, flow_count 
        FROM board 
        WHERE author_id = $1
    `)
	now := time.Now()
	boardRows := sqlmock.NewRows([]string{
		"id", "author_id", "board_name", "created_at", "is_private", "flow_count",
	}).
		AddRow(10, userID, "User Board 1", now, false, 15).
		AddRow(20, userID, "User Board 2", now, true, 7)
	mock.ExpectQuery(boardsQuery).
		WithArgs(userID).
		WillReturnRows(boardRows)

	flowsQuery := regexp.QuoteMeta(`
        SELECT f.id, f.title, f.description, f.author_id, f.created_at, 
               f.updated_at, f.is_private, f.media_url, f.like_count
        FROM flow f
        JOIN board_post bp ON f.id = bp.flow_id
        WHERE bp.board_id = $1
          AND (f.is_private = false OR f.author_id = $2)
        ORDER BY bp.saved_at DESC
        LIMIT $3 OFFSET $4
    `)
	mock.ExpectQuery(flowsQuery).
		WithArgs(10, userID, previewNum, previewStart).
		WillReturnRows(createFakeFlowRows())
	mock.ExpectQuery(flowsQuery).
		WithArgs(20, userID, previewNum, previewStart).
		WillReturnRows(createFakeFlowRows())

	boards, err := storage.GetUserAllBoards(context.Background(), userID, previewNum, previewStart)
	assert.NoError(t, err)
	assert.Len(t, boards, 2)
	for _, b := range boards {
		assert.Len(t, b.Preview, 2)
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

