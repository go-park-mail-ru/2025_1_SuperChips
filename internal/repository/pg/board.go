package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	boardService "github.com/go-park-mail-ru/2025_1_SuperChips/board"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

var (
	ErrNotFound  = errors.New("board not found")
	ErrForbidden = errors.New("forbidden")
)

const (
	countColorLimit = 20
)

type pgBoardStorage struct {
	db       *sql.DB
}

func NewBoardStorage(db *sql.DB) *pgBoardStorage {
	return &pgBoardStorage{db: db}
}

func (p *pgBoardStorage) GetUsernameID(ctx context.Context, username string, userID int) (int, error) {
	var userCheckID int
	err := p.db.QueryRowContext(ctx, `
        SELECT id 
        FROM flow_user 
        WHERE username = $1
    `, username).Scan(&userCheckID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}

	return userCheckID, nil
}

func (p *pgBoardStorage) CreateBoard(ctx context.Context, board *domain.Board, username string, userID int) error {
	var id int
	err := p.db.QueryRowContext(ctx, `
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

	row := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM board
				WHERE id = $1 AND author_id = $2
			UNION
			SELECT 1 FROM board_coauthor
				WHERE board_id = $1 AND coauthor_id = $2
		) AS is_editor
	`, boardID, userID)

	var isEditor bool
	err = row.Scan(&isEditor)
	if err != nil {
		return err
	}
	if !isEditor {
		return boardService.ErrForbidden
	}

	result, err := tx.ExecContext(ctx, `
        UPDATE board
        SET flow_count = flow_count + 1
        WHERE id = $1
    `, boardID)
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

func (p *pgBoardStorage) AddToSavedBoard(ctx context.Context, userID, flowID int) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var boardID int
	err = tx.QueryRowContext(ctx, `
	SELECT id
	FROM board
	WHERE board_name = 'Созданные вами' AND
	author_id = $1
	`, userID).Scan(&boardID)
	if err != nil {
		return err
	}

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
        WHERE board_id = $1
			AND flow_id = $3
			AND EXISTS (
				SELECT 1 FROM board
				WHERE id = $1 AND author_id = $2
				UNION
				SELECT 1 FROM board_coauthor
				WHERE board_id = $1 AND coauthor_id = $2
			)
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

    updateResult, err := tx.ExecContext(ctx, `
        UPDATE board
        SET flow_count = flow_count - 1
        WHERE id = $1
        AND flow_count > 0
    `, boardID)
    if err != nil {
        return err
    }

    updateRows, err := updateResult.RowsAffected()
    if err != nil {
        return err
    }
    if updateRows == 0 {
        return errors.New("failed to update flow_count: possible inconsistency")
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

func (p *pgBoardStorage) GetBoard(ctx context.Context, boardID, userID, previewNum, previewStart int) (domain.Board, []string, error) {
	var board domain.Board
	err := p.db.QueryRowContext(ctx, `
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
    		board.id = $1
	`, boardID).Scan(
		&board.ID,
		&board.AuthorID,
		&board.Name,
		&board.CreatedAt,
		&board.IsPrivate,
		&board.FlowCount,
		&board.AuthorUsername,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Board{}, nil, ErrNotFound
	}
	if err != nil {
		return domain.Board{}, nil, err
	}

	flows, colors, err := p.fetchFlowsAndColors(ctx, board.ID, userID, previewNum, previewStart, board.FlowCount)
	if err != nil {
		return domain.Board{}, nil, err
	}
	board.Preview = flows

	return board, colors, nil
}

func (p *pgBoardStorage) GetUserPublicBoards(ctx context.Context, username string, previewNum, previewStart int) ([]domain.Board, error) {
	var userID int
	rows, err := p.db.QueryContext(ctx, `
		SELECT DISTINCT b.id, b.author_id, b.board_name, b.created_at, b.is_private, b.flow_count
    	FROM board AS b
		LEFT JOIN board_coauthor AS bc
			ON b.id = bc.board_id
    	LEFT JOIN flow_user AS bu
			ON bu.id = b.author_id
		LEFT JOIN flow_user AS bcu
			ON bcu.id = bc.coauthor_id
    	WHERE b.is_private = false
    		AND (bu.username = $1 OR bcu.username = $1)
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

func (p *pgBoardStorage) GetUserAllBoards(ctx context.Context, userID, previewNum, previewStart int) ([]domain.Board, error) {
	rows, err := p.db.QueryContext(ctx, `
        SELECT DISTINCT b.id, b.author_id, b.board_name, b.created_at, b.is_private, b.flow_count 
		FROM board AS b
		WHERE b.author_id = $1 
			OR EXISTS (
				SELECT 1 FROM board_coauthor AS bc 
				WHERE bc.board_id = b.id AND bc.coauthor_id = $1
			)
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
		SELECT DISTINCT b.id
		FROM board AS b
		LEFT JOIN board_coauthor AS bc
			ON b.id = bc.board_id
		WHERE b.id = $1 AND (b.is_private = false OR b.author_id = $2 OR bc.coauthor_id = $2)
	`, boardID, userID).Scan(&scanID)
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
        SELECT DISTINCT 
			f.id, 
			f.title, 
			f.description, 
			f.author_id, 
			f.created_at, 
            f.updated_at, 
			f.is_private, 
			f.media_url, 
			f.like_count, 
			f.width, 
			f.height,
			bp.saved_at
        FROM flow f
        JOIN board_post bp 
			ON f.id = bp.flow_id
		LEFT JOIN board_coauthor bc 
			ON bp.board_id = bc.board_id
        WHERE bp.board_id = $1
        	AND (f.is_private = false OR f.author_id = $2 OR bc.coauthor_id = $2)
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
	var savedAt string

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
			&flow.Width,
			&flow.Height,
			&savedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan flow: %w", err)
		}

		flow.Header = middlePin.Header.String
		flow.Description = middlePin.Description.String

		flows = append(flows, flow)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during flow iteration: %w", err)
	}

	return flows, nil
}

// this function fetches first N flows with starting with offset
// also fetches colors of the first 20 flows, if possible.
func (p *pgBoardStorage) fetchFlowsAndColors(ctx context.Context, boardID, userID, pageSize, offset, pinCount int) ([]domain.PinData, []string, error) {
	rows, err := p.db.QueryContext(ctx, `
        SELECT DISTINCT
			f.id,
			f.title,
			f.description,
			f.author_id,
			f.created_at, 
            f.updated_at,
			f.is_private,
			f.media_url,
			f.like_count,
			f.width,
			f.height,
			bp.saved_at
        FROM flow f
        JOIN board_post bp
			ON f.id = bp.flow_id
		LEFT JOIN board_coauthor bc
			ON bp.board_id = bc.board_id
        WHERE bp.board_id = $1
			AND (f.is_private = false OR f.author_id = $2 OR bc.coauthor_id = $2)
        ORDER BY bp.saved_at DESC
        LIMIT $3 OFFSET $4
    `, boardID, userID, pageSize, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch flows: %w", err)
	}
	defer rows.Close()

	type middlePinData struct {
		Header      sql.NullString
		Description sql.NullString
	}

	middlePin := middlePinData{}
	var flows []domain.PinData
	var savedAt string

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
			&flow.Width,
			&flow.Height,
			&savedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan flow: %w", err)
		}

		flow.Header = middlePin.Header.String
		flow.Description = middlePin.Description.String

		flows = append(flows, flow)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error during flow iteration: %w", err)
	}

	var colors []string

	// if flow count is >= 20
	// get the colors
	if pinCount >= 20 {
		colors, err = p.fetchColors(ctx, boardID, userID)
		if err != nil {
			return nil, nil, err
		}
	}

	return flows, colors, nil
}

func (p *pgBoardStorage) fetchColors(ctx context.Context, boardID, userID int) ([]string, error) {
    query := `
		SELECT DISTINCT 
			c.color_hex,
			bp.saved_at
		FROM color c
		JOIN flow f 
			ON c.flow_id = f.id
		JOIN board_post bp 
			ON f.id = bp.flow_id
		LEFT JOIN board_coauthor bc 
			ON bc.board_id = bp.board_id
		WHERE bp.board_id = $1
			AND (f.is_private = false OR f.author_id = $2 OR bc.coauthor_id = $2)
		ORDER BY bp.saved_at DESC
		LIMIT $3
    `

    rows, err := p.db.QueryContext(ctx, query, boardID, userID, countColorLimit)
    if err != nil {
        return nil, fmt.Errorf("failed to query colors: %w", err)
    }
    defer rows.Close()

    var colors []string
	var savedAt string

    for rows.Next() {
        var color sql.NullString
        if err := rows.Scan(&color, &savedAt); err != nil {
            return nil, fmt.Errorf("failed to scan color: %w", err)
        }
		if color.String != "" {
			colors = append(colors, color.String)
		}
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("rows error: %w", err)
    }

    return colors, nil
}

