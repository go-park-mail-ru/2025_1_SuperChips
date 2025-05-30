package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	pincrudService "github.com/go-park-mail-ru/2025_1_SuperChips/pincrud"
	"github.com/pkg/errors"
)

func (p *pgPinStorage) GetPin(ctx context.Context, pinID, userID uint64) (domain.PinData, uint64, error) {
	row := p.db.QueryRowContext(ctx, `
        SELECT 
            f.id, 
            f.title, 
            f.description, 
            f.author_id, 
            f.is_private, 
            f.media_url,
            fu.username,
            f.like_count,
			f.width,
			f.height,
            CASE 
                WHEN fl.user_id IS NOT NULL THEN true
                ELSE false
            END AS is_liked
        FROM flow f
        JOIN flow_user fu ON f.author_id = fu.id
        LEFT JOIN flow_like fl ON fl.flow_id = f.id AND fl.user_id = $2
        WHERE f.id = $1;
    `, pinID, userID)

	var isLiked bool
	var flowDBRow flowDBSchema
	err := row.Scan(&flowDBRow.ID, &flowDBRow.Title, &flowDBRow.Description,
		&flowDBRow.AuthorId, &flowDBRow.IsPrivate, &flowDBRow.MediaURL,
		&flowDBRow.AuthorUsername, &flowDBRow.LikeCount, &flowDBRow.Width, &flowDBRow.Height, &isLiked)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PinData{}, 0, pincrudService.ErrPinNotFound
	}
	if err != nil {
		return domain.PinData{}, 0, pincrudService.ErrUntracked
	}

	pin := domain.PinData{
		FlowID:         flowDBRow.ID,
		Header:         flowDBRow.Title.String,
		AuthorUsername: flowDBRow.AuthorUsername,
		Description:    flowDBRow.Description.String,
		MediaURL:       p.assembleMediaURL(flowDBRow.MediaURL),
		IsPrivate:      flowDBRow.IsPrivate,
		LikeCount:      flowDBRow.LikeCount,
		IsLiked:        isLiked,
		Width:          int(flowDBRow.Width.Int64),
		Height:         int(flowDBRow.Height.Int64),
	}

	return pin, flowDBRow.AuthorId, nil
}

func (p *pgPinStorage) GetPinCleanMediaURL(ctx context.Context, pinID uint64) (string, uint64, error) {
	var mediaURL string
	var authorID uint64

	err := p.db.QueryRowContext(ctx, `
        SELECT 
            f.media_url,
			f.author_id
        FROM flow f
        JOIN flow_user fu ON f.author_id = fu.id
        WHERE f.id = $1
    `, pinID).Scan(&mediaURL, &authorID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, domain.ErrConflict
	}
	if err != nil {
		return "", 0, err
	}

	return mediaURL, authorID, nil
}

func (p *pgPinStorage) GetFromBoard(ctx context.Context, boardID, userID, flowID int) (domain.PinData, int, error) {
	row := p.db.QueryRowContext(ctx, `
		SELECT DISTINCT
            f.id, 
            f.title, 
            f.description, 
            f.author_id, 
            f.is_private, 
            f.media_url,
            fu.username,
            f.like_count,
			f.width,
			f.height,
            CASE 
                WHEN fl.user_id IS NOT NULL THEN true
                ELSE false
            END AS is_liked
        FROM flow AS f
		JOIN board_post AS bp 
			ON f.id = bp.flow_id
		LEFT JOIN board_coauthor AS bc 
			ON bp.board_id = bc.board_id
		LEFT JOIN flow_user AS fu 
			ON f.author_id = fu.id
		LEFT JOIN flow_like AS fl 
			ON fl.flow_id = f.id AND fl.user_id = $2
		WHERE 
			f.id = $3
			AND bp.board_id = $1
        	AND (f.is_private = false OR f.author_id = $2 OR bc.coauthor_id = $2)
    `, boardID, userID, flowID)
	
	var isLiked bool
	var flowDBRow flowDBSchema
	err := row.Scan(
		&flowDBRow.ID,
		&flowDBRow.Title,
		&flowDBRow.Description,
		&flowDBRow.AuthorId,
		&flowDBRow.IsPrivate,
		&flowDBRow.MediaURL,
		&flowDBRow.AuthorUsername,
		&flowDBRow.LikeCount,
		&flowDBRow.Width,
		&flowDBRow.Height,
		&isLiked)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PinData{}, 0, pincrudService.ErrPinNotFound
	}
	if err != nil {
		return domain.PinData{}, 0, pincrudService.ErrUntracked
	}

	pin := domain.PinData{
		FlowID:         flowDBRow.ID,
		Header:         flowDBRow.Title.String,
		AuthorUsername: flowDBRow.AuthorUsername,
		Description:    flowDBRow.Description.String,
		MediaURL:       p.assembleMediaURL(flowDBRow.MediaURL),
		IsPrivate:      flowDBRow.IsPrivate,
		LikeCount:      flowDBRow.LikeCount,
		IsLiked:        isLiked,
		Width:          int(flowDBRow.Width.Int64),
		Height:         int(flowDBRow.Height.Int64),
	}

	return pin, int(flowDBRow.AuthorId), nil
}

func (p *pgPinStorage) DeletePin(ctx context.Context, pinID uint64, userID uint64) error {
	res, err := p.db.ExecContext(ctx, `
		DELETE 
		FROM flow
		WHERE id=$1 AND author_id=$2
	`, pinID, userID)
	if err != nil {
		return pincrudService.ErrUntracked
	}

	count, err := res.RowsAffected()
	if err != nil || count < 1 {
		return pincrudService.ErrUntracked
	}

	return nil
}

func (p *pgPinStorage) UpdatePin(ctx context.Context, patch domain.PinDataUpdate, userID uint64) error {
	var fields []string
	var values []any
	paramCounter := 1

	if patch.Header != nil {
		fields = append(fields, fmt.Sprintf("%v = $%d", "title", paramCounter))
		values = append(values, *patch.Header)
		paramCounter++
	}

	if patch.Description != nil {
		fields = append(fields, fmt.Sprintf("%v = $%d", "description", paramCounter))
		values = append(values, *patch.Description)
		paramCounter++
	}

	if patch.IsPrivate != nil {
		fields = append(fields, fmt.Sprintf("%v = $%d", "is_private", paramCounter))
		values = append(values, *patch.IsPrivate)
		paramCounter++
	}

	if len(fields) == 0 {
		return pincrudService.ErrNoFieldsToUpdate
	}

	sqlQuery := fmt.Sprintf(
		"UPDATE flow SET %s WHERE id = $%d AND author_id = $%d",
		strings.Join(fields, ", "),
		paramCounter,
		paramCounter+1,
	)

	values = append(values, patch.FlowID)
	values = append(values, userID)

	res, err := p.db.ExecContext(ctx, sqlQuery, values...)
	if err != nil {
		return pincrudService.ErrUntracked
	}

	count, err := res.RowsAffected()
	if err != nil {
		return pincrudService.ErrUntracked
	}
	if count < 1 {
		return pincrudService.ErrPinNotFound
	}

	return nil
}

func (p *pgPinStorage) CreatePin(ctx context.Context, data domain.PinDataCreate, imgName string, userID uint64) (uint64, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx, `
        INSERT INTO flow (title, description, author_id, is_private, media_url, width, height)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
    `, data.Header, data.Description, userID, data.IsPrivate, imgName, data.Width, data.Height)

	var pinID uint64
	err = row.Scan(&pinID)
	if err != nil {
		return 0, err
	}

	log.Println("db colors len: %v", len(data.Colors))

	for i := range data.Colors {
		_, err := tx.ExecContext(ctx, `
		INSERT INTO color
		(flow_id, color_hex)
		VALUES ($1, $2)
		`, pinID, data.Colors[i])
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return pinID, nil
}
