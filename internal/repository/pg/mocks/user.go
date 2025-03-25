package mocks

import (
	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
)

type UserMock struct {
	Db   *sql.DB
	Mock sqlmock.Sqlmock
}

func NewUserMock() (*UserMock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}


	return &UserMock{
		Db:   db,
		Mock: mock,
	}, nil
}
