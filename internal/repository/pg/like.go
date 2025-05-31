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

func (p *pgLikeStorage) CheckPinAccess(ctx context.Context, pinID, userID uint64) error {
	query := `
	SELECT EXISTS (
		SELECT 1 FROM flow f
		WHERE f.id = $1
		AND (
			f.is_private = false
			OR f.author_id = $2
			OR EXISTS (
				SELECT 1 FROM board_post bp
				JOIN board b ON bp.board_id = b.id
				WHERE bp.flow_id = f.id
				AND (b.author_id = $2 OR EXISTS (
					SELECT 1 FROM board_coauthor bc
					WHERE bc.board_id = b.id AND bc.coauthor_id = $2
				))
			)
		)
	) AS has_access,
	EXISTS (SELECT 1 FROM flow WHERE id = $1) AS pin_exists
	`

	var hasAccess, pinExists bool
	err := p.db.QueryRowContext(ctx, query, pinID, userID).Scan(&hasAccess, &pinExists)
	if err != nil {
		return err
	}

	if !pinExists {
		return domain.ErrNotFound
	}
	if !hasAccess {
		return domain.ErrForbidden
	}

	return nil
}

func (pg *pgLikeStorage) LikeFlow(ctx context.Context, pinID, userID int) (string, string, error) {
    var action string
	var author string

	if err := pg.CheckPinAccess(ctx, uint64(pinID), uint64(userID)); err != nil {
		return "", "", err
	}

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

