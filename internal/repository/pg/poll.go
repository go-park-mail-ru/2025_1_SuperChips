package repository

import (
	"database/sql"
)

type pollDBSchema struct {
	ID        uint64
	Name      sql.NullString
	CreatedAt sql.NullTime
}

type questionDBSchema struct {
	ID        uint64
	PollID    uint64
	OrderNum  int64
	Content   sql.NullString
	Type      sql.NullString
	AuthorID  uint64
	CreatedAt sql.NullTime
}

type pgPollStorage struct {
	db *sql.DB
}

func NewPGPollStorage(db *sql.DB) (*pgPollStorage, error) {
	storage := &pgPollStorage{
		db: db,
	}

	return storage, nil
}
