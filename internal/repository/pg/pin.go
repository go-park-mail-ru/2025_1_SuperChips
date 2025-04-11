package repository

import (
	"database/sql"

	pin "github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type flowDBSchema struct {
	Id          uint64
	Title       sql.NullString
	Description sql.NullString
	AuthorId    uint64
	CreatedAt   sql.NullTime
	UpdatedAt   sql.NullTime
	IsPrivate   bool
	MediaURL    string
}

type pgPinStorage struct {
	db       *sql.DB
	pinDir   string
	imageURL string
}

func NewPGPinStorage(db *sql.DB, pinDir string, imageURL string) (*pgPinStorage, error) {
	storage := &pgPinStorage{
		db:       db,
		pinDir:   pinDir,
		imageURL: imageURL,
	}

	return storage, nil
}

func (p *pgPinStorage) GetPins(page int, pageSize int) ([]pin.PinData, error) {
	rows, err := p.db.Query(`SELECT id, title, description, author_id, is_private, media_url 
	FROM flow
	WHERE is_private = false
	LIMIT $1
	OFFSET $2
	`, pageSize, (page-1)*pageSize)
	if err != nil {
		return []pin.PinData{}, err
	}

	defer rows.Close()

	var pins []pin.PinData

	for rows.Next() {
		var flowDBRow flowDBSchema
		err := rows.Scan(&flowDBRow.Id, &flowDBRow.Title, &flowDBRow.Description, &flowDBRow.AuthorId, &flowDBRow.IsPrivate, &flowDBRow.MediaURL)
		if err != nil {
			return []pin.PinData{}, err
		}

		pin := pin.PinData{
			FlowID:      flowDBRow.Id,
			Description: flowDBRow.Description.String,
			Header:      flowDBRow.Title.String,
			MediaURL:    flowDBRow.MediaURL,
			AuthorID:    flowDBRow.AuthorId,
		}
		pins = append(pins, pin)
	}

	return pins, nil
}
