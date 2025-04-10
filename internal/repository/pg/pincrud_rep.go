package repository

import (
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	pincrudService "github.com/go-park-mail-ru/2025_1_SuperChips/pincrud"
	"github.com/go-park-mail-ru/2025_1_SuperChips/utils/image"
	"github.com/pkg/errors"
)

func (p *pgPinStorage) GetPin(pinID uint64) (domain.PinData, error) {
	row := p.db.QueryRow(`
		SELECT id, title, description, author_id, is_private, media_url 
		FROM flow
		WHERE id=$1
	`, pinID)

	var flowPinDB flowPinDB
	err := row.Scan(&flowPinDB.Id, &flowPinDB.Title, &flowPinDB.Description, &flowPinDB.AuthorId, &flowPinDB.IsPrivate, &flowPinDB.MediaURL)
	if err != nil {
		var errToService error
		switch {
		case errors.Is(err, sql.ErrNoRows):
			errToService = pincrudService.ErrPinNotFound
		default:
			errToService = domain.WrapError(pincrudService.ErrUntracked, err)
		}
		return domain.PinData{}, errToService
	}

	pin := domain.PinData{
		FlowID:      flowPinDB.Id,
		Header:      flowPinDB.Title.String,
		AuthorID:    flowPinDB.AuthorId,
		Description: flowPinDB.Description.String,
		MediaURL:    p.pinDir + flowPinDB.MediaURL,
		IsPrivate:   flowPinDB.IsPrivate,
	}

	return pin, nil
}

func (p *pgPinStorage) DeletePinByID(pinID uint64, userID uint64) error {
	row := p.db.QueryRow(`
		SELECT media_url
		FROM flow
		WHERE id = $1 AND author_id = $2
	`, pinID, userID)

	var imgName string
	err := row.Scan(&imgName)
	if err != nil {
		var errToService error
		switch {
		case errors.Is(err, sql.ErrNoRows):
			errToService = pincrudService.ErrPinNotFound
		default:
			errToService = domain.WrapError(pincrudService.ErrUntracked, err)
		}
		return errToService
	}

	imgPath := filepath.Join(p.pinDir, imgName)
	_, err = os.Stat(imgPath)
	if os.IsNotExist(err) {
		return domain.WrapError(pincrudService.ErrUntracked, err)
	}
	err = os.Remove(imgPath)
	if err != nil {
		return domain.WrapError(pincrudService.ErrUntracked, err)
	}

	res, err := p.db.Exec(`
		DELETE 
		FROM flow
		WHERE id=$1 AND author_id=$2
	`, pinID, userID)
	if err != nil {
		return domain.WrapError(pincrudService.ErrUntracked, err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return domain.WrapError(pincrudService.ErrUntracked, err)
	}
	if count < 1 {
		return pincrudService.ErrPinNotFound
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
		return domain.WrapError(pincrudService.ErrUntracked, err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return domain.WrapError(pincrudService.ErrUntracked, err)
	}
	if count < 1 {
		return pincrudService.ErrPinNotFound
	}

	return nil
}

func (p *pgPinStorage) CreatePin(data domain.PinDataCreate, file multipart.File, header *multipart.FileHeader, userID uint64) error {
	if !image.IsImageFile(header.Filename) || filepath.Ext(header.Filename) == "" {
		return pincrudService.ErrInvalidImageExt
	}

	var cancel bool = true
	var pinID uint64
	row := p.db.QueryRow(`
        INSERT INTO flow (title, description, author_id, is_private)
        VALUES ($1, $2, $3, $4)
		RETURNING id
    `, data.Header, data.Description, userID, data.IsPrivate)

	err := row.Scan(&pinID)
	if err != nil {
		return domain.WrapError(pincrudService.ErrUntracked, err)
	}
	defer func() {
		if !cancel {
			return
		}
		p.db.Exec(`
			DELETE 
			FROM flow
			WHERE id=$1 AND author_id=$2
		`, pinID, userID)
	}()

	imgName := strconv.FormatUint(pinID, 10) + filepath.Ext(header.Filename)
	imgPath := filepath.Join(p.pinDir, imgName)
	dst, err := os.Create(imgPath)
	if err != nil {
		return domain.WrapError(pincrudService.ErrUntracked, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return domain.WrapError(pincrudService.ErrUntracked, err)
	}

	res, err := p.db.Exec(`
		UPDATE flow 
		SET media_url = $1 
		WHERE id = $2 AND author_id = $3
	`, imgName, pinID, userID)
	if err != nil {
		return domain.WrapError(pincrudService.ErrUntracked, err)
	}
	count, err := res.RowsAffected()
	if err != nil || count < 1 {
		return domain.WrapError(pincrudService.ErrUntracked, err)
	}

	cancel = false
	return nil
}
