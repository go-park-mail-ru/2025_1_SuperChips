package repository

import (
	"context"
	"database/sql"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
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
	WITH access_check AS (
		SELECT 
			CASE 
				WHEN f.is_private = false OR f.author_id = $1 THEN true
				ELSE false
			END AS has_access
		FROM flow f
		WHERE f.id = $2
	),
	deleted AS (
		DELETE FROM flow_like
		WHERE user_id = $1 AND flow_id = $2
		AND (SELECT has_access FROM access_check) = true
		RETURNING 'delete' AS action
	),
	inserted AS (
		INSERT INTO flow_like (user_id, flow_id)
		SELECT $1, $2
		WHERE NOT EXISTS (SELECT 1 FROM deleted)
		AND (SELECT has_access FROM access_check) = true
		RETURNING 'insert' AS action
	),
	update_like_count AS (
		UPDATE flow
		SET like_count = like_count + CASE
			WHEN EXISTS (SELECT 1 FROM inserted) THEN 1
			WHEN EXISTS (SELECT 1 FROM deleted) THEN -1
			ELSE 0
		END
		WHERE id = $2
	)
	SELECT COALESCE((SELECT action FROM inserted), (SELECT action FROM deleted)) AS action
    `, userID, pinID).Scan(&action)
    if err == sql.ErrNoRows {
        return "", domain.ErrForbidden
    }
    if err != nil {
        return "", err
    }

    return action, nil
}

