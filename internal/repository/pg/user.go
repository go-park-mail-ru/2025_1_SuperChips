package repository

import (
	"database/sql"
	"errors"

	user "github.com/go-park-mail-ru/2025_1_SuperChips/domain"
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

func (p *pgUserStorage) AddUser(userInfo user.User) (uint64, error) {
	var id uint64
	err := p.db.QueryRow(`
        INSERT INTO flow_user (username, avatar, public_name, email, password, birthday)
        VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (email, username) DO NOTHING
		RETURNING id
    `, userInfo.Username, userInfo.Avatar, userInfo.PublicName, userInfo.Email, userInfo.Password, userInfo.Birthday).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, user.ErrConflict
	} else if err != nil {
		return 0, err
	}

	return id, nil
}

func (p *pgUserStorage) GetHash(email, password string) (uint64, string, error) {
	var hashedPassword string
	var id uint64

	err := p.db.QueryRow(`
        SELECT id, password FROM flow_user WHERE email = $1
    `, email).Scan(&id, &hashedPassword)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return 0, "", user.ErrInvalidCredentials
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil
}

func (p *pgUserStorage) GetUserPublicInfo(email string) (user.PublicUser, error) {
	var userDB userDB

	err := p.db.QueryRow(`
        SELECT username, email, avatar, birthday, about, public_name,
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

func (p *pgUserStorage) GetUserId(email string) (uint64, error) {
	var id uint64

	err := p.db.QueryRow(`
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
