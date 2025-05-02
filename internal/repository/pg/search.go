package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type SearchRepository struct {
	db *sql.DB
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
		f.width,
		f.height,
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
	var header sql.NullString
	var description sql.NullString

	for rows.Next() {
		var pin domain.PinData
		if err := rows.Scan(
			&pin.FlowID,
			&header,
			&description,
			&pin.AuthorID,
			&pin.IsPrivate,
			&pin.MediaURL,
			&pin.Width,
			&pin.Height,
			&pin.AuthorUsername,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		pin.Header = header.String
		pin.Description = description.String

		pins = append(pins, pin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return pins, nil
}

func (s *SearchRepository) SearchUsers(ctx context.Context, query string, page, pageSize int) ([]domain.PublicUser, error) {
	offset := (page - 1) * pageSize

	var isExternalAvatar sql.NullBool

	rows, err := s.db.QueryContext(ctx, `
    SELECT username, email, avatar, birthday, about, public_name, is_external_avatar
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
		var about sql.NullString
		if err := rows.Scan(
			&user.Username,
			&user.Email,
			&user.Avatar,
			&user.Birthday,
			&about,
			&user.PublicName,
			&isExternalAvatar,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row %w", err)
		}

		user.About = about.String
		user.IsExternalAvatar = isExternalAvatar.Bool

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return users, nil
}

func (s *SearchRepository) SearchBoards(ctx context.Context, query string, page, pageSize, previewNum, previewStart int) ([]domain.Board, error) {
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
		var name sql.NullString
		if err := rows.Scan(
			&board.ID,
			&board.AuthorID,
			&name,
			&board.CreatedAt,
			&board.IsPrivate,
			&board.FlowCount,
			&board.AuthorUsername,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row %w", err)
		}

		board.Name = name.String

		boards = append(boards, board)
	}

	for i := range boards {
		preview, err := s.fetchFirstNFlowsForBoard(ctx, boards[i].ID, 0, previewNum, previewStart)
		if err != nil {
			// наверное, если произошла ошибка
			// при получении превью, имеет смысл
			// просто не отображать его, а не выдавать
			// хттп ошибку
			continue
		}

		boards[i].Preview = preview
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return boards, nil
}

func (p *SearchRepository) fetchFirstNFlowsForBoard(ctx context.Context, boardID, userID, pageSize, offset int) ([]domain.PinData, error) {
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

		flow.Header = middlePin.Header.String
		flow.Description = middlePin.Description.String

		flows = append(flows, flow)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during flow iteration: %w", err)
	}

	return flows, nil
}

