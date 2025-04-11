package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	pincrudService "github.com/go-park-mail-ru/2025_1_SuperChips/pincrud"
	"github.com/pkg/errors"
)

func (p *pgPinStorage) GetPin(pinID uint64) (domain.PinData, error) {
	row := p.db.QueryRow(`
		SELECT id, title, description, author_id, is_private, media_url 
		FROM flow
		WHERE id=$1
	`, pinID)

	var flowDBRow flowDBSchema
	err := row.Scan(&flowDBRow.Id, &flowDBRow.Title, &flowDBRow.Description, &flowDBRow.AuthorId, &flowDBRow.IsPrivate, &flowDBRow.MediaURL)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PinData{}, pincrudService.ErrPinNotFound
	}
	if err != nil {
		return domain.PinData{}, pincrudService.ErrUntracked
	}

	pin := domain.PinData{
		FlowID:      flowDBRow.Id,
		Header:      flowDBRow.Title.String,
		AuthorID:    flowDBRow.AuthorId,
		Description: flowDBRow.Description.String,
		MediaURL:    p.assembleMediaURL(flowDBRow.MediaURL),
		IsPrivate:   flowDBRow.IsPrivate,
	}

	return pin, nil
}

func (p *pgPinStorage) DeletePin(pinID uint64, userID uint64) error {
	res, err := p.db.Exec(`
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

func (p *pgPinStorage) UpdatePin(patch domain.PinDataUpdate, userID uint64) error {
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

	res, err := p.db.Exec(sqlQuery, values...)
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

func (p *pgPinStorage) CreatePin(data domain.PinDataCreate, imgName string, userID uint64) (uint64, error) {
	row := p.db.QueryRow(`
        INSERT INTO flow (title, description, author_id, is_private, media_url)
        VALUES ($1, $2, $3, $4, $5)
		RETURNING id
    `, data.Header, data.Description, userID, data.IsPrivate, imgName)

	var pinID uint64
	err := row.Scan(&pinID)
	if err != nil {
		return 0, pincrudService.ErrUntracked
	}

	return pinID, nil
}
