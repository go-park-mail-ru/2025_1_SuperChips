package repository

import (
	"database/sql"

	pin "github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

const (
	CREATE_PIN_TABLE = `
		CREATE TABLE IF NOT EXISTS flow (
		flow_id INTEGER AUTOINCREMENT PRIMARY KEY,
		title TEXT,
		description TEXT,
		author_id INTEGER NOT NULL,
		create_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		is_private BOOLEAN NOT NULL,
		media_url TEXT,
		FOREIGN KEY (author_id) REFERENCES user(user_id)
		`
)

// GetPins(page int, pageSize int) []domain.PinData

type pgPinStorage struct {
	db *sql.DB
}

func NewPGPinStorage(db *sql.DB) (*pgPinStorage, error) {
	storage := &pgPinStorage{
		db: db,
	}

	storage.initialize()

	return storage, nil
}

func (p *pgPinStorage) initialize() error {
	_, err := p.db.Exec(CREATE_PIN_TABLE)
	if err != nil {
		return err
	}

	return nil
}

func (p *pgPinStorage) GetPins(page int, pageSize int) ([]pin.PinData, error) {
	rows, err := p.db.Query("SELECT * FROM flow LIMIT $1 OFFSET $2", pageSize, (page-1)*pageSize)
	if err != nil {
		return []pin.PinData{}, err
	}

	defer rows.Close()

	var pins []pin.PinData

	for rows.Next() {
		var pin pin.PinData
		err := rows.Scan(&pin.FlowID, &pin.Header, &pin.Description, &pin.AuthorID, &pin.Created_at, &pin.Updated_at, &pin.IsPrivate, &pin.MediaURL)
		if err != nil {
			return pins, err
		}

		pins = append(pins, pin)
	}

	return pins, nil
}
