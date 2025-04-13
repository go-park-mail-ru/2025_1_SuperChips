package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	boardService "github.com/go-park-mail-ru/2025_1_SuperChips/board"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

var (
	ErrNotFound  = errors.New("board not found")
	ErrForbidden = errors.New("forbidden")
)

const (
	previewNum   = 3 // количество изображений для превью
	previewStart = 0 // смещение от начала доски для превью
)

type pgBoardStorage struct {
	db       *sql.DB
	baseURL  string
	imageDir string
}

func NewBoardStorage(db *sql.DB, imageDir string, baseURL string) *pgBoardStorage {
	return &pgBoardStorage{db: db, baseURL: baseURL, imageDir: imageDir}
}

func (p *pgBoardStorage) CreateBoard(ctx context.Context, board *domain.Board, username string, userID int) error {
	var userCheckID int
	err := p.db.QueryRowContext(ctx, `
        SELECT id 
        FROM flow_user 
        WHERE username = $1
    `, username).Scan(&userCheckID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	if userCheckID != userID {
		return boardService.ErrForbidden
	}

	var id int
	err = p.db.QueryRowContext(ctx, `
        INSERT INTO board (author_id, board_name, is_private)
        VALUES ($1, $2, $3)
        ON CONFLICT (author_id, board_name) DO NOTHING
        RETURNING id
    `, userID, board.Name, board.IsPrivate).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrConflict
	}
	if err != nil {
		return err
	}

	board.ID = id
	return nil
}

func (p *pgBoardStorage) DeleteBoard(ctx context.Context, boardID, userID int) error {
	var id int
	err := p.db.QueryRowContext(ctx, `
	DELETE FROM board 
	WHERE id = $1
	AND
	author_id = $2
	RETURNING id`, boardID, userID).
		Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	return nil
}

func (p *pgBoardStorage) AddToBoard(ctx context.Context, boardID, userID, flowID int) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
        UPDATE board
        SET flow_count = flow_count + 1
        WHERE id = $1 AND author_id = $2
    `, boardID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	var insertedID int
	err = tx.QueryRowContext(ctx, `
        INSERT INTO board_post (board_id, flow_id)
        VALUES ($1, $2)
        RETURNING board_id
    `, boardID, flowID).Scan(&insertedID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (p *pgBoardStorage) DeleteFromBoard(ctx context.Context, boardID, userID, flowID int) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
        DELETE FROM board_post
        USING board
        WHERE board_post.board_id = $1
        AND board_post.flow_id = $3
        AND board.id = $1
        AND board.author_id = $2
    `, boardID, userID, flowID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	_, err = tx.ExecContext(ctx, `
        UPDATE board
        SET flow_count = flow_count - 1
        WHERE id = $1
    `, boardID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (p *pgBoardStorage) UpdateBoard(ctx context.Context, boardID, userID int, newName string, isPrivate bool) error {
	query := `
        UPDATE board
        SET board_name = COALESCE($1, board_name),
            is_private = COALESCE($2, is_private)
        WHERE id = $3 AND author_id = $4
    `

	result, err := p.db.ExecContext(ctx, query, newName, isPrivate, boardID, userID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (p *pgBoardStorage) GetBoard(ctx context.Context, boardID int) (domain.Board, error) {
	var board domain.Board
	err := p.db.QueryRowContext(ctx, `
        SELECT id, author_id, board_name, created_at, is_private, flow_count 
        FROM board 
        WHERE id = $1
    `, boardID).Scan(
		&board.ID,
		&board.AuthorID,
		&board.Name,
		&board.CreatedAt,
		&board.IsPrivate,
		&board.FlowCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Board{}, ErrNotFound
	}
	if err != nil {
		return domain.Board{}, err
	}

	flows, err := p.fetchFirstNFlowsForBoard(ctx, board.ID, board.AuthorID, previewNum, previewStart)
	if err != nil {
		return domain.Board{}, err
	}
	board.Preview = flows

	return board, nil
}

func (p *pgBoardStorage) GetUserPublicBoards(ctx context.Context, username string) ([]domain.Board, error) {
	var userID int
	rows, err := p.db.QueryContext(ctx, `
    SELECT b.id, b.author_id, b.board_name, b.created_at, b.is_private, b.flow_count
    FROM flow_user u
    JOIN board b ON u.id = b.author_id
    WHERE u.username = $1 
    AND b.is_private = false
	`, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boards []domain.Board
	for rows.Next() {
		var board domain.Board
		err := rows.Scan(
			&board.ID,
			&board.AuthorID,
			&board.Name,
			&board.CreatedAt,
			&board.IsPrivate,
			&board.FlowCount,
		)
		if err != nil {
			return nil, err
		}

		flows, err := p.fetchFirstNFlowsForBoard(ctx, board.ID, userID, previewNum, previewStart)
		if err != nil {
			return nil, err
		}
		board.Preview = flows

		boards = append(boards, board)
	}
	return boards, nil
}

func (p *pgBoardStorage) GetUserAllBoards(ctx context.Context, userID int) ([]domain.Board, error) {
	rows, err := p.db.QueryContext(ctx, `
        SELECT id, author_id, board_name, created_at, is_private, flow_count 
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
			&board.ID,
			&board.AuthorID,
			&board.Name,
			&board.CreatedAt,
			&board.IsPrivate,
			&board.FlowCount,
		)
		if err != nil {
			return nil, err
		}

		flows, err := p.fetchFirstNFlowsForBoard(ctx, board.ID, userID, previewNum, previewStart)
		if err != nil {
			return nil, err
		}
		board.Preview = flows

		boards = append(boards, board)
	}

	return boards, nil
}

func (p *pgBoardStorage) GetBoardFlow(ctx context.Context, boardID, userID, page, pageSize int) ([]domain.PinData, error) {
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	var scanID int

	err := p.db.QueryRowContext(ctx, `
	SELECT id
	FROM board
	WHERE id = $1 AND (is_private = false
	OR author_id = $2)`, boardID, userID).Scan(&scanID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, boardService.ErrForbidden
	}
	if err != nil {
		return nil, err
	}

	return p.fetchFirstNFlowsForBoard(ctx, boardID, userID, pageSize, offset)
}

func (p *pgBoardStorage) fetchFirstNFlowsForBoard(ctx context.Context, boardID, userID, pageSize, offset int) ([]domain.PinData, error) {
	rows, err := p.db.QueryContext(ctx, `
        SELECT f.id, f.title, f.description, f.author_id, f.created_at, 
               f.updated_at, f.is_private, f.media_url, f.like_count
        FROM flow f
        JOIN board_post bp ON f.id = bp.flow_id
        WHERE bp.board_id = $1
          AND (f.is_private = false OR f.author_id = $2)
        ORDER BY bp.saved_at DESC
        LIMIT $3 OFFSET $4
    `, boardID, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch flows: %w", err)
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
		if err != nil {
			return nil, fmt.Errorf("failed to scan flow: %w", err)
		}

		flow.MediaURL = p.generateImageURL(flow.MediaURL)
		flow.Header = middlePin.Header.String
		flow.Description = middlePin.Description.String

		flows = append(flows, flow)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during flow iteration: %w", err)
	}

	return flows, nil
}

func (p *pgBoardStorage) generateImageURL(filename string) string {
	return p.baseURL + filepath.Join(strings.ReplaceAll(p.imageDir, ".", ""), filename)
}

