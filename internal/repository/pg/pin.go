package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pin "github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type flowDB struct {
	Flow_id     uint64
	Title       sql.NullString
	Description sql.NullString
	Author_id   uint64
	Created_at  string
	Updated_at  string
	Is_private  bool
	Media_url   string
}

const (
	CREATE_PIN_TABLE = `
		CREATE TABLE IF NOT EXISTS flow (
		flow_id SERIAL PRIMARY KEY,
		title TEXT,
		description TEXT,
		author_id INTEGER NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		is_private BOOLEAN NOT NULL DEFAULT FALSE,
		media_url TEXT NOT NULL,
		FOREIGN KEY (author_id) REFERENCES flow_user(user_id)
		);
		`
)

// GetPins(page int, pageSize int) []domain.PinData

type pgPinStorage struct {
	db      *sql.DB
	baseDir string
}

func NewPGPinStorage(db *sql.DB, baseDir string) (*pgPinStorage, error) {
	storage := &pgPinStorage{
		db:      db,
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
	rows, err := p.db.Query("SELECT flow_id, title, description, author_id, is_private, media_url FROM flow LIMIT $1 OFFSET $2", pageSize, (page-1)*pageSize)
	if err != nil {
		return []pin.PinData{}, err
	}

	defer rows.Close()

	var pins []pin.PinData

	for rows.Next() {
		var flowDB flowDB
		err := rows.Scan(&flowDB.Flow_id, &flowDB.Title, &flowDB.Description, &flowDB.Author_id, &flowDB.Is_private, &flowDB.Media_url)
		if err != nil {
			return []pin.PinData{}, err
		}

		if !flowDB.Is_private {
			pin := pin.PinData{
				Description: flowDB.Description.String,
				Header:   flowDB.Title.String,
				MediaURL: flowDB.Media_url,
				AuthorID: flowDB.Author_id,
			}
			pins = append(pins, pin)
		}
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

	// add one user who will have all pins
	_, err = p.db.Exec("INSERT INTO flow_user (username, avatar, public_name, email, password) VALUES ($1, $2, $3, $4, $5)", "admin", "", "admin", "admin@yourflow", "admin")
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && isImageFile(file.Name()) {
			_, err := p.db.Exec("INSERT INTO flow (title, media_url, author_id) VALUES ($1, $2, $3)", fmt.Sprintf("Header %d", id), fmt.Sprintf("https://yourflow.ru/static/img/%s", file.Name()), 1)
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
