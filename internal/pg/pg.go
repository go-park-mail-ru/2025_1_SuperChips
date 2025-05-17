package pg

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	_ "github.com/jackc/pgx"
)

func ConnectDB(cfg configs.PostgresConfig, ctx context.Context) (*sql.DB, error) {
	var pool *sql.DB
	var err error

	maxRetries := 10
	retryDelay := time.Second

	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", 
		cfg.PgHost, 
		cfg.PgPort, 
		cfg.PgUser, 
		cfg.PgPassword, 
		cfg.PgDB)

	for attempt := 0; attempt < maxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("context canceled: %w", err)
		}

		pool, err = sql.Open("postgres", connString)
		if err != nil {
			log.Printf("Attempt %d: Failed to open database connection: %v", attempt+1, err)
		}

		pool.SetMaxOpenConns(cfg.MaxOpenConns)
		pool.SetMaxIdleConns(cfg.MaxIdleConns)
		pool.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
		
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

