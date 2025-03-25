package repository

import (
	"database/sql"

	user "github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/security"
)

type userDB struct {
	user_id     uint64
	username    string
	avatar      sql.NullString
	public_name string
	email       string
	create_at   string
	updated_at  string
	password    string
	birthday    sql.NullTime
	about       sql.NullString
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

func (p *pgUserStorage) AddUser(userInfo user.User) error {
	hashedPassword, err := security.HashPassword(userInfo.Password)
	if err != nil {
		return err
	}

	row := p.db.QueryRow(`SELECT user_id FROM flow_user WHERE email = $1 OR username = $2`, userInfo.Email, userInfo.Username)
	var id uint64
	err = row.Scan(&id)
	if err == sql.ErrNoRows {
	} else if err != nil {
		return err
	} else {
		return user.ErrConflict
	}

	_, err = p.db.Exec(`
        INSERT INTO flow_user (username, avatar, public_name, email, password)
        VALUES ($1, $2, $3, $4, $5)
    `, userInfo.Username, userInfo.Avatar, userInfo.PublicName, userInfo.Email, hashedPassword)
	if err != nil {
		return err
	}

	return nil
}

func (p *pgUserStorage) LoginUser(email, password string) (string, error) {
	var hashedPassword string

	err := p.db.QueryRow(`
        SELECT password FROM flow_user WHERE email = $1
    `, email).Scan(&hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", user.ErrInvalidCredentials
		}

		return "", err
	}

	return hashedPassword, nil
}

func (p *pgUserStorage) GetUserPublicInfo(email string) (user.PublicUser, error) {
	var userDB userDB

	err := p.db.QueryRow(`
        SELECT username, email, avatar, birthday FROM flow_user WHERE email = $1
    `, email).Scan(&userDB.username, &userDB.email, &userDB.avatar, &userDB.birthday)
	if err != nil {
		if err == sql.ErrNoRows {
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

	err := p.db.QueryRow(`
        SELECT user_id FROM flow_user WHERE email = $1
    `, email).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, user.ErrUserNotFound
		}
		return 0, err
	}

	return id, nil
}
