package repository

import (
	"context"
	"database/sql"
)

type pgLikeStorage struct {
	db *sql.DB
}

func NewPgLikeStorage(db *sql.DB) *pgLikeStorage {
	return &pgLikeStorage{
		db: db,
	}
}

func (pg *pgLikeStorage) LikeFlow(ctx context.Context, pinID, userID int) (string, error) {
    var action string

    err := pg.db.QueryRowContext(ctx, `
        WITH deleted AS (
            DELETE FROM flow_like
            WHERE user_id = $1 AND flow_id = $2
            RETURNING 'delete' AS action
        ),
        inserted AS (
            INSERT INTO flow_like (user_id, flow_id)
            SELECT $1, $2
            WHERE NOT EXISTS (SELECT 1 FROM deleted)
            RETURNING 'insert' AS action
        )
        SELECT COALESCE((SELECT action FROM inserted), (SELECT action FROM deleted)) AS action;
    `, userID, pinID).Scan(&action)

    if err != nil {
        return "", err
    }

    return action, nil
}

