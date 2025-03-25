package repository

import (
	"database/sql"

	pin "github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/utils/pins"
)

type flowPinDB struct {
	Id     uint64
	Title       sql.NullString
	Description sql.NullString
	Author_id   uint64
	Created_at  sql.NullTime
	Updated_at  sql.NullTime
	Is_private  bool
	Media_url   string
}

type pgPinStorage struct {
	db      *sql.DB
	baseDir string
}

func NewPGPinStorageWithImages(db *sql.DB, baseDir string) (*pgPinStorage, error) {
	storage := &pgPinStorage{
		db:      db,
		baseDir: baseDir,
	}

	if err := pins.AddAllPins(storage.db, storage.baseDir); err != nil {
		return nil, err
	}

	return storage, nil
}

func NewPGPinStorage(db *sql.DB) (*pgPinStorage, error) {
	storage := &pgPinStorage{
		db: db,
	}


	return storage, nil
}

func (p *pgPinStorage) GetPins(page int, pageSize int) ([]pin.PinData, error) {
	rows, err := p.db.Query("SELECT id, title, description, author_id, is_private, media_url FROM flow LIMIT $1 OFFSET $2", pageSize, (page-1)*pageSize)
	if err != nil {
		return []pin.PinData{}, err
	}

	defer rows.Close()

	var pins []pin.PinData

	for rows.Next() {
		var flowPinDB flowPinDB
		err := rows.Scan(&flowPinDB.Id, &flowPinDB.Title, &flowPinDB.Description, &flowPinDB.Author_id, &flowPinDB.Is_private, &flowPinDB.Media_url)
		if err != nil {
			return []pin.PinData{}, err
		}

		if !flowPinDB.Is_private {
			pin := pin.PinData{
				FlowID:      flowPinDB.Id,
				Description: flowPinDB.Description.String,
				Header:      flowPinDB.Title.String,
				MediaURL:    flowPinDB.Media_url,
				AuthorID:    flowPinDB.Author_id,
			}
			pins = append(pins, pin)
		}
	}

	return pins, nil
}


