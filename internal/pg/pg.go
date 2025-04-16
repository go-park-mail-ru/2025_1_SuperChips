package pg

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx"

)

func ConnectDB(connString string, ctx context.Context) (*sql.DB, error) {
	var pool *sql.DB
	var err error

	maxRetries := 10
	retryDelay := time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("context canceled: %w", err)
		}

		pool, err = sql.Open("postgres", connString)
		if err != nil {
			log.Printf("Attempt %d: Failed to open database connection: %v", attempt+1, err)
		}

		if err := pool.Ping(); err != nil {
			log.Printf("Attempt %d: Failed to ping database: %v", attempt+1, err)
			pool.Close()
		} else {
			return pool, nil
		}

		time.Sleep(retryDelay)
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts", maxRetries)
}

