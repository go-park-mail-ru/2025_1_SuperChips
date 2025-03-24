package mocks

import (
	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
)

type PinMock struct {
	Db   *sql.DB
	Mock sqlmock.Sqlmock
}

func NewPinMock() (*PinMock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}

	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS flow \( flow_id SERIAL PRIMARY KEY, title TEXT, description TEXT, author_id INTEGER NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW\(\), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW\(\), is_private BOOLEAN NOT NULL DEFAULT FALSE, media_url TEXT NOT NULL, FOREIGN KEY \(author_id\) REFERENCES flow_user\(user_id\) \);`).WillReturnResult(sqlmock.NewResult(0, 0))

	return &PinMock{
		Db:   db,
		Mock: mock,
	}, nil
}

