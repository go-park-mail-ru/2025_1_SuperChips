package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

var (
	ErrNotFound = errors.New("board not found")
)

type pgBoardStorage struct {
	db       *sql.DB
	pageSize int
}

func NewBoardStorage(db *sql.DB, pageSize int) *pgBoardStorage {
	return &pgBoardStorage{db: db, pageSize: pageSize}
}

func (p *pgBoardStorage) CreateBoard(board *domain.Board, username string) (int, error) {
	var userID int
	err := p.db.QueryRow(`
        SELECT id 
        FROM flow_user 
        WHERE username = $1
    `, username).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, err
	}

	var id int
	err = p.db.QueryRow(`
        INSERT INTO board (author_id, board_name, is_private)
        VALUES ($1, $2, $3)
        ON CONFLICT (author_id, board_name) DO NOTHING
        RETURNING id
    `, userID, board.Name, board.IsPrivate).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, domain.ErrConflict
	} else if err != nil {
		return 0, err
	}

	board.Id = id
	return userID, nil
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
        INSERT INTO board_post (board_id, flow_id)
        SELECT $1, $3
        WHERE EXISTS (SELECT 1 FROM board_check)
        RETURNING *
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
        DELETE FROM board_post
        WHERE board_id = $1 AND flow_id = $3
          AND EXISTS (SELECT 1 FROM board_check)
		RETURNING board_id
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

func (p *pgBoardStorage) UpdateBoard(boardID, userID int, newName string, isPrivate bool) error {
	query := `
        UPDATE board
        SET board_name = COALESCE($1, board_name),
            is_private = COALESCE($2, is_private)
        WHERE id = $3 AND author_id = $4
    `

	result, err := p.db.Exec(query, newName, isPrivate, boardID, userID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (p *pgBoardStorage) GetBoard(boardID int) (domain.Board, error) {
	var board domain.Board
	err := p.db.QueryRow(`
        SELECT id, author_id, board_name, created_at, is_private 
        FROM board 
        WHERE id = $1
    `, boardID).Scan(
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

func (p *pgBoardStorage) GetUserPublicBoards(username string) ([]domain.Board, error) {
	var userID int
	err := p.db.QueryRow(`
        SELECT id 
        FROM flow_user 
        WHERE username = $1
    `, username).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

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

func (p *pgBoardStorage) GetBoardFlow(boardID, userID, page int) ([]domain.PinData, error) {
	offset := (page - 1) * p.pageSize
	if offset < 0 {
		offset = 0
	}

	query := `
        SELECT f.id, f.title, f.description, f.author_id, f.created_at, 
               f.updated_at, f.is_private, f.media_url, f.like_count
        FROM flow f
        JOIN board_post bp ON f.id = bp.flow_id
        WHERE bp.board_id = $1
          AND (f.is_private = false OR f.author_id = $2)
        ORDER BY bp.saved_at DESC
        LIMIT $3 OFFSET $4
    `

	rows, err := p.db.Query(query, boardID, userID, p.pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	type middlePinData struct {
		Header      sql.NullString
		Description sql.NullString
	}

	middlePin := middlePinData{}

	var flows []domain.PinData
	for rows.Next() {
		var flow domain.PinData
		err := rows.Scan(
			&flow.FlowID,
			&middlePin.Header,
			&middlePin.Description,
			&flow.AuthorID,
			&flow.CreatedAt,
			&flow.UpdatedAt,
			&flow.IsPrivate,
			&flow.MediaURL,
			&flow.LikeCount,
		)

		flow.Header = middlePin.Header.String
		flow.Description = middlePin.Description.String

		if err != nil {
			return nil, fmt.Errorf("failed to scan flow: %w", err)
		}
		flows = append(flows, flow)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return flows, nil
}
