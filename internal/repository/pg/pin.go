package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pin "github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type flowDB struct {
	Flow_id     uint64
	Title       pgtype.Text
	Description pgtype.Text
	Author_id   uint64
	Created_at  pgtype.Timestamptz
	Updated_at  pgtype.Timestamptz
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
	db      *pgxpool.Pool
	baseDir string
}

func NewPGPinStorage(db *pgxpool.Pool, baseDir string) (*pgPinStorage, error) {
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
	_, err := p.db.Exec(context.Background(), CREATE_PIN_TABLE)
	if err != nil {
		return err
	}

	return nil
}

func (p *pgPinStorage) GetPins(page int, pageSize int) ([]pin.PinData, error) {
	rows, err := p.db.Query(context.Background(), "SELECT flow_id, title, description, author_id, is_private, media_url FROM flow LIMIT $1 OFFSET $2", pageSize, (page-1)*pageSize)
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

	_, err = p.db.Exec(context.Background(), "INSERT INTO flow_user (username, avatar, public_name, email, password) VALUES ($1, $2, $3, $4, $5)", "admin", "", "admin", "admin@yourflow", "admin")
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && isImageFile(file.Name()) {
			_, err := p.db.Exec(context.Background(), "INSERT INTO flow (title, media_url, author_id) VALUES ($1, $2, $3)", fmt.Sprintf("Header %d", id), fmt.Sprintf("https://yourflow.ru/static/img/%s", file.Name()), 1)
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
