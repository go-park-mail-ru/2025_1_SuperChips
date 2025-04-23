package repository

import (
	"context"
	"database/sql"
	"errors"

	user "github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/security"
	_ "github.com/jmoiron/sqlx"
)

type userDB struct {
	Id         uint64         `db:"id"`
	Username   string         `db:"username"`
	Avatar     sql.NullString `db:"avatar"`
	PublicName string         `db:"public_name"`
	Email      string         `db:"email"`
	CreatedAt  string         `db:"created_at"`
	UpdatedAt  string         `db:"updated_at"`
	Password   string         `db:"password"`
	Birthday   sql.NullTime   `db:"birthday"`
	About      sql.NullString `db:"about"`
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

func (p *pgUserStorage) AddUser(ctx context.Context, userInfo user.User) (uint64, error) {
	var id uint64
	err := p.db.QueryRowContext(ctx, `
        INSERT INTO flow_user (username, avatar, public_name, email, password, birthday)
        VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (email, username) DO NOTHING
		RETURNING id
    `, userInfo.Username, userInfo.Avatar, userInfo.Username, userInfo.Email, userInfo.Password, userInfo.Birthday).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, user.ErrConflict
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
		return 0, "", user.ErrInvalidCredentials
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil
}

func (p *pgUserStorage) GetUserPublicInfo(ctx context.Context, email string) (user.PublicUser, error) {
	var userDB userDB

	err := p.db.QueryRowContext(ctx, `
        SELECT username, email, avatar, birthday, about, public_name
		FROM flow_user WHERE email = $1
    `, email).Scan(&userDB.Username, &userDB.Email, &userDB.Avatar, &userDB.Birthday)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return user.PublicUser{}, user.ErrInvalidCredentials
	} else if err != nil {
		return user.PublicUser{}, err
	}

	publicUser := user.PublicUser{
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
			return 0, user.ErrUserNotFound
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
		return 0, "", user.ErrNotFound
	}
	if err != nil {
		return 0, "", err
	}

	return id, gotEmail, nil
}

func (p *pgUserStorage) AddExternalUser(ctx context.Context, email, username string, externalID string) (uint64, error) {
	var id uint64
	dummyPassword, err := security.GenerateRandomHash()
	if err != nil {
		return 0, err
	}

	err = p.db.QueryRowContext(ctx, `
        INSERT INTO flow_user (username, public_name, email, password, external_id)
        VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (email, username) DO NOTHING
		RETURNING id
    `, username, username, email, dummyPassword, externalID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, user.ErrConflict
	} else if err != nil {
		return 0, err
	}

	return id, nil
}
