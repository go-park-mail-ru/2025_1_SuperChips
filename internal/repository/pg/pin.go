package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	baseDir string
}

func NewPGPinStorage(db *sql.DB, baseDir string) (*pgPinStorage, error) {
	storage := &pgPinStorage{
		db: db,
		baseDir: baseDir,
	}

	if err := storage.initialize(); err != nil {
		return nil, err
	}
	
	if err := storage.addAllPins(); err != nil {
		return nil, err
	}

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

// временная функция, добавляющая все файлы из директории в базу данных
func (p *pgPinStorage) addAllPins() error {
	files, err := os.ReadDir(p.baseDir)
	if err != nil {
		return err
	}

	id := 1

	for _, file := range files {
		if !file.IsDir() && isImageFile(file.Name()) {
			_, err := p.db.Exec("INSERT INTO flow (title, media_url, author_id) VALUES ($1, $2, $3)", fmt.Sprintf("Header %d", id), fmt.Sprintf("https://yourflow.ru/static/img/%s", file.Name()), id)
			if err != nil {
				return err
			}
			id++
		}
	}

	return nil
}

func isImageFile(filename string) bool {
    ext := strings.ToLower(filepath.Ext(filename))

    pattern := "*.jpg;*.jpeg;*.png;*.gif;*.bmp;*.tiff;*.webp"

    for _, p := range strings.Split(pattern, ";") {
        match, err := filepath.Match(p, ext)
        if err != nil {
            continue
        }
        if match {
            return true
        }
    }

    return false
}
