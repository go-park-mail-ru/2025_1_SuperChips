package rest

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB(connString string) (*sql.DB, error) {
	pool, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(); err != nil {
		return nil, err
	}

	return pool, nil
}
