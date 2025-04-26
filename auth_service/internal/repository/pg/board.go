package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-park-mail-ru/2025_1_SuperChips/auth_service"
)

var (
	ErrNotFound  = errors.New("board not found")
	ErrForbidden = errors.New("forbidden")
)

type pgBoardStorage struct {
	db       *sql.DB
}

func NewBoardStorage(db *sql.DB) *pgBoardStorage {
	return &pgBoardStorage{db: db}
}

func (p *pgBoardStorage) CreateBoard(ctx context.Context, board *models.Board, username string, userID int) error {
	var id int
	err := p.db.QueryRowContext(ctx, `
        INSERT INTO board (author_id, board_name, is_private)
        VALUES ($1, $2, $3)
        ON CONFLICT (author_id, board_name) DO NOTHING
        RETURNING id
    `, userID, board.Name, board.IsPrivate).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return models.ErrConflict
	}
	if err != nil {
		return err
	}

	board.ID = id
	return nil
}

