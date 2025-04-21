package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

// type SearchRepository interface {
// 	SearchPins(query string, page, pageSize int) ([]domain.PinData, error)
// 	SearchUsers(query string, page, pageSize int) ([]domain.PublicUser, error)
// 	SearchBoards(query string, page, pageSize int) ([]domain.Board, error)
// }

type SearchRepository struct {
	db *sql.DB
}

type UserNullable struct {
	About sql.NullString
}

type PinNullable struct {
	Header      sql.NullString
	Description sql.NullString
}

type BoardNullable struct {
	BoardName sql.NullString
}

func NewSearchRepository(db *sql.DB) *SearchRepository {
	return &SearchRepository{
		db: db,
	}
}

func (s *SearchRepository) SearchPins(ctx context.Context, query string, page, pageSize int) ([]domain.PinData, error) {
	offset := (page - 1) * pageSize

	queryString := `
    SELECT 
        f.id, 
        f.title, 
        f.description, 
        f.author_id, 
        f.is_private, 
        f.media_url,
        fu.username
    FROM flow f
    JOIN flow_user fu ON f.author_id = fu.id
    WHERE f.is_private = false
    AND (to_tsvector(f.title || ' ' || f.description) @@ plainto_tsquery($1) OR
	f.title ILIKE '%' || $1 || '%' OR
	f.description ILIKE '%' || $1 || '%')
    ORDER BY f.like_count DESC
    LIMIT $2
    OFFSET $3
    `

	rows, err := s.db.QueryContext(ctx, queryString, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var pins []domain.PinData
	for rows.Next() {
		var pin domain.PinData
		var pinNullable PinNullable
		if err := rows.Scan(
			&pin.FlowID,
			&pinNullable.Header,
			&pinNullable.Description,
			&pin.AuthorID,
			&pin.IsPrivate,
			&pin.MediaURL,
			&pin.AuthorUsername,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		pin.Header = pinNullable.Header.String
		pin.Description = pinNullable.Description.String

		pins = append(pins, pin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return pins, nil
}

func (s *SearchRepository) SearchUsers(ctx context.Context, query string, page, pageSize int) ([]domain.PublicUser, error) {
	offset := (page - 1) * pageSize

	rows, err := s.db.QueryContext(ctx, `
    SELECT username, email, avatar, birthday, about, public_name
    FROM flow_user
    WHERE to_tsvector(username) @@ plainto_tsquery($1) OR 
	username ILIKE '%' || $1 || '%'
	LIMIT $2
	OFFSET $3
	`, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var users []domain.PublicUser
	for rows.Next() {
		var user domain.PublicUser
		var userNullable UserNullable
		if err := rows.Scan(
			&user.Username,
			&user.Email,
			&user.Avatar,
			&user.Birthday,
			&userNullable.About,
			&user.PublicName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row %w", err)
		}

		user.About = userNullable.About.String

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return users, nil
}

func (s *SearchRepository) SearchBoards(ctx context.Context, query string, page, pageSize int) ([]domain.Board, error) {
	offset := (page - 1) * pageSize

	rows, err := s.db.QueryContext(ctx, `
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
	WHERE board.is_private = false AND
    (
        board.board_name ILIKE '%' || $1 || '%' OR
        to_tsvector(board.board_name) @@ plainto_tsquery($1)
    )
	LIMIT $2
	OFFSET $3`, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var boards []domain.Board
	for rows.Next() {
		var board domain.Board
		var boardNullable BoardNullable
		if err := rows.Scan(
			&board.ID,
			&board.AuthorID,
			&boardNullable.BoardName,
			&board.CreatedAt,
			&board.IsPrivate,
			&board.FlowCount,
			&board.AuthorUsername,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row %w", err)
		}

		board.Name = boardNullable.BoardName.String

		boards = append(boards, board)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return boards, nil
}
