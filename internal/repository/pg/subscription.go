package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type SubscriptionStorage struct {
	db *sql.DB
} 

func NewSubscriptionStorage(db *sql.DB) *SubscriptionStorage {
	return &SubscriptionStorage{
		db: db,
	}
}

func (repo *SubscriptionStorage) GetUserFollowers(ctx context.Context, id, page, size int) ([]domain.PublicUser, error) {
	offset := (page - 1) * size
	rows, err := repo.db.QueryContext(ctx, `
	SELECT u.username, u.avatar, u.birthday, u.about, u.public_name, u.subscriber_count, u.is_external_avatar
	FROM subscription
	LEFT JOIN flow_user u ON subscription.target_id = u.id
	WHERE subscription.target_id = $1
	ORDER BY CASE WHEN subscription.created_at IS NULL THEN 1 ELSE 0 END, subscription.created_at DESC
	OFFSET $2
	LIMIT $3
	`, id, offset, size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.PublicUser

	for rows.Next() {
		var user userDB
		err := rows.Scan(
			&user.Username,
			&user.Avatar,
			&user.Birthday,
			&user.About,
			&user.PublicName,
			&user.SubscriberCount,
			&user.IsExternalAvatar,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, domain.PublicUser{
			Username: user.Username,
			Avatar: user.Avatar.String,
			Birthday: user.Birthday.Time,
			About: user.About.String,
			PublicName: user.PublicName,
			SubscriberCount: int(user.SubscriberCount),
			IsExternalAvatar: user.IsExternalAvatar.Bool,
		})
	}

	return users, nil
}

func (repo *SubscriptionStorage) GetUserFollowing(ctx context.Context, id, page, size int) ([]domain.PublicUser, error) {
	offset := (page - 1) * size
	rows, err := repo.db.QueryContext(ctx, `
	SELECT
		u.username,
		u.avatar,
		u.birthday,
		u.about,
		u.public_name,
		u.subscriber_count,
		u.is_external_avatar
	FROM subscription
	LEFT JOIN flow_user u ON subscription.target_id = u.id
	WHERE subscription.user_id = $1
	ORDER BY CASE WHEN subscription.created_at IS NULL THEN 1 ELSE 0 END, subscription.created_at DESC
	OFFSET $2
	LIMIT $3
	`, id, offset, size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.PublicUser

	for rows.Next() {
		var user userDB
		err := rows.Scan(
			&user.Username,
			&user.Avatar,
			&user.Birthday,
			&user.About,
			&user.PublicName,
			&user.SubscriberCount,
			&user.IsExternalAvatar,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, domain.PublicUser{
			Username: user.Username,
			Avatar: user.Avatar.String,
			Birthday: user.Birthday.Time,
			About: user.About.String,
			PublicName: user.PublicName,
			SubscriberCount: int(user.SubscriberCount),
			IsExternalAvatar: user.IsExternalAvatar.Bool,
		})
	}

	return users, nil
}

func (repo *SubscriptionStorage) CreateSubscription(ctx context.Context, targetUsername string, currentID int) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var targetID int

	err = tx.QueryRowContext(ctx, `
	SELECT id FROM flow_user
	WHERE username = $1
	`, targetUsername).Scan(&targetID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return err
	}

	if currentID == targetID {
		return domain.ErrValidation
	}

	res, err := tx.ExecContext(ctx, `
	INSERT INTO subscription
		(user_id, target_id)
	VALUES
		($1, $2)
	ON CONFLICT (user_id, target_id) DO NOTHING
	`, currentID, targetID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return domain.ErrConflict
	}

	res, err = tx.ExecContext(ctx, `
	UPDATE flow_user
	SET subscriber_count = subscriber_count + 1
	WHERE id = $1
	`, currentID)
	if err != nil {
		return err
	}

	count, err = res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return domain.ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (repo *SubscriptionStorage) DeleteSubscription(ctx context.Context, targetUsername string, currentID int) error	{
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var targetID int

	err = tx.QueryRowContext(ctx, `
	SELECT id FROM flow_user
	WHERE username = $1
	`, targetUsername).Scan(&targetID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, `
	DELETE FROM subscription
	WHERE user_id = $1 AND target_id = $2
	`, currentID, targetID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return domain.ErrNotFound
	}

	res, err = tx.ExecContext(ctx, `
	UPDATE flow_user
	SET subscriber_count = subscriber_count - 1
	WHERE id = $1
	`, currentID)
	if err != nil {
		return err
	}

	count, err = res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return domain.ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

