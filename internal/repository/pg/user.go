package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	_ "github.com/jmoiron/sqlx"
)

type userDB struct {
	ID               uint64         `db:"id"`
	Username         string         `db:"username"`
	Avatar           sql.NullString `db:"avatar"`
	PublicName       string         `db:"public_name"`
	Email            string         `db:"email"`
	CreatedAt        string         `db:"created_at"`
	UpdatedAt        string         `db:"updated_at"`
	Password         string         `db:"password"`
	Birthday         sql.NullTime   `db:"birthday"`
	About            sql.NullString `db:"about"`
	IsExternalAvatar sql.NullBool
	SubscriberCount  sql.NullInt64
}

type pgUserStorage struct {
	db *sql.DB
}

func NewPGUserStorage(db *sql.DB) (*pgUserStorage, error) {
	storage := &pgUserStorage{
		db: db,
	}

	return storage, nil
}

func (p *pgUserStorage) AddUser(ctx context.Context, userInfo domain.User) (uint64, error) {
	var id uint64
	err := p.db.QueryRowContext(ctx, `
	WITH conflict_check AS (
		SELECT id
		FROM flow_user
		WHERE username = $1 OR email = $4
	)
	INSERT INTO flow_user (username, avatar, public_name, email, password)
	SELECT $1, $2, $3, $4, $5
	WHERE NOT EXISTS (SELECT 1 FROM conflict_check)
	RETURNING id;
    `, userInfo.Username, userInfo.Avatar, userInfo.Username, userInfo.Email, userInfo.Password).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, domain.ErrConflict
	} else if err != nil {
		return 0, err
	}

	return id, nil
}

func (p *pgUserStorage) GetHash(ctx context.Context, email, password string) (uint64, string, string, error) {
	var hashedPassword string
	var id uint64
	var username string

	err := p.db.QueryRowContext(ctx, `
        SELECT id, password, username
		FROM flow_user
		WHERE email = $1
    `, email).Scan(&id, &hashedPassword, &username)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", "", domain.ErrInvalidCredentials
	}
	if err != nil {
		return 0, "", "", err
	}

	return id, hashedPassword, username, nil
}

func (p *pgUserStorage) GetUserPublicInfo(ctx context.Context, email string) (domain.PublicUser, error) {
	var userDB userDB

	err := p.db.QueryRowContext(ctx, `
        SELECT username, email, avatar, birthday, about, public_name, subscriber_count
		FROM flow_user WHERE email = $1
    `, email).Scan(&userDB.Username, &userDB.Email, &userDB.Avatar, &userDB.Birthday, &userDB.About, &userDB.PublicName, &userDB.SubscriberCount)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return domain.PublicUser{}, domain.ErrInvalidCredentials
	} else if err != nil {
		return domain.PublicUser{}, err
	}

	publicUser := domain.PublicUser{
		Username: userDB.Username,
		Email:    userDB.Email,
		Avatar:   userDB.Avatar.String,
		Birthday: userDB.Birthday.Time,
		SubscriberCount: int(userDB.SubscriberCount.Int64),
	}

	return publicUser, nil
}

func (p *pgUserStorage) GetUserId(ctx context.Context, email string) (uint64, error) {
	var id uint64

	err := p.db.QueryRowContext(ctx, `
        SELECT id FROM flow_user WHERE email = $1
    `, email).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, domain.ErrUserNotFound
		}
		return 0, err
	}

	return id, nil
}

func (p *pgUserStorage) CheckImgPermission(ctx context.Context, imageName string, userID int) (bool, error) {
    query := `
    SELECT EXISTS (
        SELECT 1 FROM flow f
        WHERE f.media_url = $1
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
    EXISTS (SELECT 1 FROM flow WHERE media_url = $1) AS image_exists
    `

    var hasAccess, imageExists bool
    err := p.db.QueryRowContext(ctx, query, imageName, userID).Scan(&hasAccess, &imageExists)
    if err != nil {
        return false, fmt.Errorf("failed to check image permission: %w", err)
    }

    if !imageExists {
        return false, nil
    }

    return hasAccess, nil
}

func (p *pgUserStorage) FindExternalServiceUser(ctx context.Context, email string, externalID string) (int, string, string, error) {
	var id int
	var gotEmail string
	var username string

	err := p.db.QueryRowContext(ctx, `
	SELECT id, email, username
	FROM flow_user
	WHERE external_id = $1
	AND email = $2`, externalID, email).Scan(&id, &gotEmail, &username)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", "", domain.ErrNotFound
	}
	if err != nil {
		return 0, "", "", err
	}

	return id, gotEmail, username, nil
}

func (p *pgUserStorage) AddExternalUser(ctx context.Context, email, username, password, avatarURL string, externalID string) (uint64, error) {
	var id uint64

	err := p.db.QueryRowContext(ctx, `
	WITH conflict_check AS (
		SELECT id
		FROM flow_user
		WHERE email = $3 OR username = $1
	)
	INSERT INTO flow_user (username, public_name, email, password, external_id, avatar, is_external_avatar)
	SELECT $1, $2, $3, $4, $5, $6, $7
	WHERE NOT EXISTS (SELECT 1 FROM conflict_check)
	RETURNING id;
    `, username, username, email, password, externalID, avatarURL, true).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, domain.ErrConflict
	} else if err != nil {
		return 0, err
	}

	return id, nil
}
