package repository

import (
	"context"
	"database/sql"
	"errors"

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

func (pg *pgLikeStorage) LikeFlow(ctx context.Context, pinID, userID int) (string, string, error) {
    var action string
	var author string

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
	SELECT 
		fu.username AS author_username, 
		COALESCE((SELECT action FROM inserted), (SELECT action FROM deleted)) AS action
	FROM flow f
	JOIN flow_user fu ON f.author_id = fu.id
	WHERE f.id = $2`,
		userID, pinID).Scan(&author, &action)
    if errors.Is(err, sql.ErrNoRows) {
        return "", "", domain.ErrForbidden
    }
    if err != nil {
        return "", "", err
    }

    return action, author, nil
}

