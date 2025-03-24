package repository

import (
	"context"

	user "github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	security "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/security"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userDB struct {
	user_id     uint64
	username    string
	avatar      pgtype.Text
	public_name string
	email       string
	create_at   string
	updated_at  string
	password    string
	birthday    pgtype.Timestamptz
	about       pgtype.Text
}

type pgUserStorage struct {
	db *pgxpool.Pool
}

// AddUser(user domain.User) error
// LoginUser(email, password string) error
// GetUserPublicInfo(email string) (domain.PublicUser, error)
// GetUserId(email string) uint64

const (
	CREATE_USER_TABLE = `
        CREATE TABLE IF NOT EXISTS flow_user (
            user_id SERIAL PRIMARY KEY,
            username TEXT NOT NULL UNIQUE,
            avatar TEXT,
            public_name TEXT NOT NULL,
            email TEXT NOT NULL UNIQUE,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            password TEXT NOT NULL,
            birthday DATE,
            about TEXT,
            jwt_version INTEGER NOT NULL DEFAULT 1
        );
    `
)

func NewPGUserStorage(db *pgxpool.Pool) (*pgUserStorage, error) {
	storage := &pgUserStorage{
		db: db,
	}

	if err := storage.initialize(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (p *pgUserStorage) initialize() error {
	_, err := p.db.Exec(context.Background(), CREATE_USER_TABLE)
	if err != nil {
		return err
	}

	return nil
}

func (p *pgUserStorage) AddUser(userInfo user.User) error {
	if err := userInfo.ValidateUser(); err != nil {
		return err
	}

	hashedPassword, err := security.HashPassword(userInfo.Password)
	if err != nil {
		return err
	}

	row := p.db.QueryRow(context.Background(), `SELECT user_id FROM flow_user WHERE email = $1 OR username = $2`, userInfo.Email, userInfo.Username)
	var id uint64
	err = row.Scan(&id)
	if err == pgx.ErrNoRows {
	} else if err != nil {
		return err
	} else {
		return user.ErrConflict
	}

	_, err = p.db.Exec(context.Background(), `
        INSERT INTO flow_user (username, avatar, public_name, email, password)
        VALUES ($1, $2, $3, $4, $5)
    `, userInfo.Username, userInfo.Avatar, userInfo.PublicName, userInfo.Email, hashedPassword)
	if err != nil {
		return err
	}

	return nil
}

func (p *pgUserStorage) LoginUser(email, password string) error {
	var hashedPassword string

	if err := user.ValidateEmailAndPassword(email, password); err != nil {
		return err
	}

	err := p.db.QueryRow(context.Background(), `
        SELECT password FROM flow_user WHERE email = $1
    `, email).Scan(&hashedPassword)
	if err != nil {
		if err == pgx.ErrNoRows {
			return user.ErrInvalidCredentials
		}

		return err
	}

	ok := security.ComparePassword(password, hashedPassword)
	if !ok {
		return user.ErrInvalidCredentials
	}

	return nil
}

func (p *pgUserStorage) GetUserPublicInfo(email string) (user.PublicUser, error) {
	var userDB userDB

	err := p.db.QueryRow(context.Background(), `
        SELECT username, email, avatar, birthday FROM flow_user WHERE email = $1
    `, email).Scan(&userDB.username, &userDB.email, &userDB.avatar, &userDB.birthday)
	if err != nil {
		if err == pgx.ErrNoRows {
			return user.PublicUser{}, user.ErrUserNotFound
		}
		return user.PublicUser{}, err
	}

	publicUser := user.PublicUser{
		Username: userDB.username,
		Email:    userDB.email,
		Avatar:   userDB.avatar.String,
		Birthday: userDB.birthday.Time,
	}

	return publicUser, nil
}

func (p *pgUserStorage) GetUserId(email string) (uint64, error) {
	var id uint64

	err := p.db.QueryRow(context.Background(), `
        SELECT user_id FROM flow_user WHERE email = $1
    `, email).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, user.ErrUserNotFound
		}
		return 0, err
	}

	return id, nil
}
