package repository

import (
	"database/sql"
	"path/filepath"
	"strings"

	pin "github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	_ "github.com/jmoiron/sqlx"
)

type flowPinDB struct {
	Id          uint64         `db:"id"`
	Title       sql.NullString `db:"title"`
	Description sql.NullString `db:"description"`
	AuthorId    uint64         `db:"author_id"`
	CreatedAt   sql.NullTime   `db:"created_at"`
	UpdatedAt   sql.NullTime   `db:"updated_at"`
	IsPrivate   bool           `db:"is_private"`
	MediaURL    string         `db:"media_url"`
}

type pgPinStorage struct {
	db       *sql.DB
	imageDir string
	baseURL  string
}

func NewPGPinStorage(db *sql.DB, imageDir string, baseURL string) (*pgPinStorage, error) {
	storage := &pgPinStorage{
		db: db,
		imageDir: imageDir,
		baseURL: baseURL,
	}

	return storage, nil
}

func (p *pgPinStorage) GetPins(page int, pageSize int) ([]pin.PinData, error) {
	rows, err := p.db.Query(`SELECT id, title, description, author_id, is_private, media_url 
	FROM flow
	WHERE is_private = false
	ORDER BY created_at DESC
	LIMIT $1
	OFFSET $2
	`, pageSize, (page-1)*pageSize)
	if err != nil {
		return []pin.PinData{}, err
	}

	defer rows.Close()

	var pins []pin.PinData

	for rows.Next() {
		var flowPinDB flowPinDB
		err := rows.Scan(&flowPinDB.Id, &flowPinDB.Title, &flowPinDB.Description, &flowPinDB.AuthorId, &flowPinDB.IsPrivate, &flowPinDB.MediaURL)
		if err != nil {
			return []pin.PinData{}, err
		}

		pin := pin.PinData{
			FlowID:      flowPinDB.Id,
			Description: flowPinDB.Description.String,
			Header:      flowPinDB.Title.String,
			MediaURL:    p.generateImageURL(flowPinDB.MediaURL),
			AuthorID:    flowPinDB.AuthorId,
		}
		pins = append(pins, pin)
	}

	return pins, nil
}

func (p *pgPinStorage) generateImageURL(filename string) string {
	return p.baseURL + filepath.Join(strings.ReplaceAll(p.imageDir, ".", ""), filename)
}