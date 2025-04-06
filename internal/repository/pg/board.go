package repository

import (
	"database/sql"
	"errors"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

// type BoardRepository interface {
// 	CreateBoard(name string, authorID int) error
// 	DeleteBoard(name string, authorID int) error
// 	AddToBoard(name string, authorID, flowID int) error       // == update board
// 	DeleteFromBoard(name string, authorID, flowID int) error  // == update board
// 	GetBoard(name string, authorID int) (domain.Board, error) // == get board
// 	GetUserPublicBoards(userID int) ([]domain.Board, error)   // == get board
// 	GetUserAllBoards(userID int) ([]domain.Board, error)
// }

type pgBoardStorage struct {
	db *sql.DB
}

func NewPGBoardStorage(db *sql.DB) *pgBoardStorage {
	return &pgBoardStorage{
		db: db,
	}
}

func (p pgBoardStorage) CreateBoard(name string, authorID int, isPrivate bool) error {
	_, err := p.db.Exec(`
	INSERT INTO board (author_id, board_name)
	VALUES ($1, $2)
	ON CONFLICT (author_id, board_name) DO NOTHING
	RETURNING id
	`)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrConflict
	} else if err != nil {
		return err
	}

	return nil
}
