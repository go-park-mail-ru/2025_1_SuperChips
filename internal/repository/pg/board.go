package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

var (
	ErrNotFound = errors.New("board not found")
)

type pgBoardStorage struct {
	db *sql.DB
}

func NewBoardStorage(db *sql.DB) *pgBoardStorage {
	return &pgBoardStorage{db: db}
}

func (p *pgBoardStorage) CreateBoard(board domain.Board) error {
	var id int
	err := p.db.QueryRow(`
        INSERT INTO board (author_id, board_name, is_private)
        VALUES ($1, $2, $3)
        ON CONFLICT (author_id, board_name) DO NOTHING
        RETURNING id
    `, board.AuthorID, board.Name, board.IsPrivate).Scan(&id)

	if err == sql.ErrNoRows {
		return domain.ErrConflict
	} else if err != nil {
		return err
	}

	board.Id = id
	return nil
}

func (p *pgBoardStorage) DeleteBoard(boardID, userID int) error {
	var id int
	err := p.db.QueryRow(`DELETE FROM board 
	WHERE id = $1
	AND
	author_id = $2
	RETURNING id`, boardID, userID).
		Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func (p *pgBoardStorage) AddToBoard(boardID, userID, flowID int) error {
	var insertedID int
	err := p.db.QueryRow(`
        WITH board_check AS (
            SELECT id
            FROM board
            WHERE id = $1 AND author_id = $2
        )
        INSERT INTO board_flow (board_id, flow_id)
        SELECT $1, $3
        WHERE EXISTS (SELECT 1 FROM board_check)
        RETURNING board_id
    `, boardID, userID, flowID).Scan(&insertedID)
	if err == sql.ErrNoRows {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	return err
}

func (p *pgBoardStorage) DeleteFromBoard(boardID, userID, flowID int) error {
	var checkID int
	err := p.db.QueryRow(`
        WITH board_check AS (
            SELECT id
            FROM board
            WHERE id = $1 AND author_id = $2
        )
        DELETE FROM board_flow
        WHERE board_id = $1 AND flow_id = $3
          AND EXISTS (SELECT 1 FROM board_check)
    `, boardID, userID, flowID).
		Scan(&checkID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}

		return err
	}

	return nil
}

func (p *pgBoardStorage) UpdateBoard(board domain.Board, userID int, newName *string, isPrivate *bool) error {
	query := "UPDATE board SET "
	var params []interface{}
	setClauses := []string{}
	paramCount := 1

	if newName != nil {
		setClauses = append(setClauses, fmt.Sprintf("board_name = $%d", paramCount))
		params = append(params, *newName)
		paramCount++
	}

	if isPrivate != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_private = $%d", paramCount))
		params = append(params, *isPrivate)
		paramCount++
	}

	if len(setClauses) == 0 {
		return errors.New("no fields to update")
	}

	query += strings.Join(setClauses, ", ") + " WHERE id = $" + strconv.Itoa(paramCount) + " AND author_id = $" + strconv.Itoa(paramCount+1)
	params = append(params, board.Id, userID)

	result, err := p.db.Exec(query, params...)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
func (p *pgBoardStorage) GetBoard(name string, authorID int) (domain.Board, error) {
	var board domain.Board
	err := p.db.QueryRow(`
        SELECT id, author_id, board_name, created_at, is_private 
        FROM board 
        WHERE author_id = $1 AND board_name = $2
    `, authorID, name).Scan(
		&board.Id,
		&board.AuthorID,
		&board.Name,
		&board.CreatedAt,
		&board.IsPrivate,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Board{}, ErrNotFound
		}
		return domain.Board{}, err
	}
	return board, nil
}

func (p *pgBoardStorage) GetBoardByID(boardID int) (domain.Board, error) {
	var board domain.Board
	err := p.db.QueryRow(`
	SELECT id, author_id, board_name, created_at, is_private
	FROM board
	WHERE id = $1`,
		boardID).Scan(
		&boardID,
		&board.AuthorID,
		&board.Name,
		&board.CreatedAt,
		&board.IsPrivate,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Board{}, ErrNotFound
		}
		return domain.Board{}, err
	}

	return board, nil
}

func (p *pgBoardStorage) GetUserPublicBoards(userID int) ([]domain.Board, error) {
	rows, err := p.db.Query(`
        SELECT id, author_id, board_name, created_at, is_private 
        FROM board 
        WHERE author_id = $1 AND is_private = false
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boards []domain.Board
	for rows.Next() {
		var board domain.Board
		err := rows.Scan(
			&board.Id,
			&board.AuthorID,
			&board.Name,
			&board.CreatedAt,
			&board.IsPrivate,
		)
		if err != nil {
			return nil, err
		}
		boards = append(boards, board)
	}
	return boards, nil
}

func (p *pgBoardStorage) GetUserAllBoards(userID int) ([]domain.Board, error) {
	rows, err := p.db.Query(`
        SELECT id, author_id, board_name, created_at, is_private 
        FROM board 
        WHERE author_id = $1
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boards []domain.Board
	for rows.Next() {
		var board domain.Board
		err := rows.Scan(
			&board.Id,
			&board.AuthorID,
			&board.Name,
			&board.CreatedAt,
			&board.IsPrivate,
		)
		if err != nil {
			return nil, err
		}
		boards = append(boards, board)
	}
	return boards, nil
}
