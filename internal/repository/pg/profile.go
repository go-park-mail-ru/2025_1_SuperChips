package repository

import (
	"database/sql"
	"errors"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)


type pgProfileStorage struct {
	db *sql.DB
}

func NewPGProfileStorage(db *sql.DB) (*pgProfileStorage, error) {
	storage := &pgProfileStorage{
		db: db,
	}

	return storage, nil
}

func (p *pgProfileStorage) GetUserPublicInfoByEmail(email string) (domain.User, error) {
	var userDB userDB

	err := p.db.QueryRow(`
		SELECT id, username, email, avatar, birthday, about, public_name
		FROM flow_user WHERE email = $1
	`, email).Scan(&userDB.Id, &userDB.Username, &userDB.Email, &userDB.Avatar, &userDB.Birthday, &userDB.About, &userDB.PublicName)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	} else if err != nil {
		return domain.User{}, err
	}

	user := domain.User{
		Id:         userDB.Id,
		Username:   userDB.Username,
		Email:      userDB.Email,
		Avatar:     userDB.Avatar.String,
		Birthday:   userDB.Birthday.Time,
		PublicName: userDB.PublicName,
		About:      userDB.About.String,
	}

	return user, nil
}

func (p *pgProfileStorage) GetUserPublicInfoByUsername(username string) (domain.User, error) {
	var userDB userDB

	err := p.db.QueryRow(`
		SELECT id, username, email, avatar, birthday, about, public_name
		FROM flow_user WHERE username = $1
	`, username).Scan(&userDB.Id, &userDB.Username, &userDB.Email, &userDB.Avatar, &userDB.Birthday, &userDB.About, &userDB.PublicName)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	} else if err != nil {
		return domain.User{}, err
	}

	user := domain.User{
		Id:         userDB.Id,
		Username:   userDB.Username,
		Email:      userDB.Email,
		Avatar:     userDB.Avatar.String,
		Birthday:   userDB.Birthday.Time,
		PublicName: userDB.PublicName,
		About:      userDB.About.String,
	}

	return user, nil
}

func (p *pgProfileStorage) SaveUserAvatar(email string, avatar string) error {
	_, err := p.db.Exec(`
		UPDATE flow_user SET avatar = $1 WHERE email = $2
	`, avatar, email)
	if err != nil {
		return err
	}

	return nil
}

func (p *pgProfileStorage) UpdateUserData(user domain.User, oldEmail string) error {
	_, err := p.db.Exec(`
		UPDATE flow_user SET username = $1, birthday = $2, about = $3, public_name = $4, email = $5
		WHERE email = $6
	`, user.Username, user.Birthday, user.About, user.PublicName, user.Email, oldEmail)
	if err != nil {
		return err
	}

	return nil
}
