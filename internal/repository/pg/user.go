package repository

import "database/sql"

type pgUserStorage struct {
	db *sql.DB
}

func NewPGUserStorage(db *sql.DB) (*pgUserStorage, error) {
	storage := &pgUserStorage{
		db: db,
	}

	storage.initialize()

	return storage, nil
}

func (p *pgUserStorage) initialize() error {
	return nil
}