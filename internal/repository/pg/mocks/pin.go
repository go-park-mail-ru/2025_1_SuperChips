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

	return &PinMock{
		Db:   db,
		Mock: mock,
	}, nil
}

