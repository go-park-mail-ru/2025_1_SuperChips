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

	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS flow_user \( user_id SERIAL PRIMARY KEY, username TEXT NOT NULL UNIQUE, avatar TEXT, public_name TEXT NOT NULL, email TEXT NOT NULL UNIQUE, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW\(\), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW\(\), password TEXT NOT NULL, birthday DATE, about TEXT, jwt_version INTEGER NOT NULL DEFAULT 1 \);`).WillReturnResult(sqlmock.NewResult(0, 0))

	return &UserMock{
		Db:   db,
		Mock: mock,
	}, nil
}
