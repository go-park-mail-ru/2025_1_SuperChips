package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	_ "github.com/jmoiron/sqlx"
)

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

func (p *pgUserStorage) GetHash(ctx context.Context, email, password string) (uint64, string, error) {
	var hashedPassword string
	var id uint64

	err := p.db.QueryRowContext(ctx, `
        SELECT id, password FROM flow_user WHERE email = $1
    `, email).Scan(&id, &hashedPassword)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return 0, "", domain.ErrInvalidCredentials
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil
}

func (p *pgUserStorage) GetUserPublicInfo(ctx context.Context, email string) (domain.PublicUser, error) {
	var userDB userDB

	err := p.db.QueryRowContext(ctx, `
        SELECT username, email, avatar, birthday, about, public_name
		FROM flow_user WHERE email = $1
    `, email).Scan(&userDB.Username, &userDB.Email, &userDB.Avatar, &userDB.Birthday)
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

func (p *pgUserStorage) FindExternalServiceUser(ctx context.Context, email string, externalID string) (int, string, error) {
	var id int
	var gotEmail string

	err := p.db.QueryRowContext(ctx, `
	SELECT id, email
	FROM flow_user
	WHERE external_id = $1
	AND email = $2`, externalID, email).Scan(&id, &gotEmail)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", domain.ErrNotFound
	}
	if err != nil {
		return 0, "", err
	}

	return id, gotEmail, nil
}

func (p *pgUserStorage) AddExternalUser(ctx context.Context, email, username, password, avatarURL string, externalID string) (uint64, error) {
	var id uint64

	err := p.db.QueryRowContext(ctx, `
	WITH conflict_check AS (
		SELECT id
		FROM flow_user
		WHERE email = $3 OR username = $1
	)
	INSERT INTO flow_user (username, public_name, email, password, external_id, avatar)
	SELECT $1, $2, $3, $4, $5
	WHERE NOT EXISTS (SELECT 1 FROM conflict_check)
	RETURNING id;
    `, username, username, email, password, externalID, avatarURL).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, domain.ErrConflict
	} else if err != nil {
		return 0, err
	}

	return id, nil
}
