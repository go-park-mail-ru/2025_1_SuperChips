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
	var externalID sql.NullString

	err := p.db.QueryRow(`
		SELECT id, username, email, avatar, birthday, about, public_name, external_id, is_external_avatar, subscriber_count
		FROM flow_user WHERE email = $1
	`, email).Scan(&userDB.ID, &userDB.Username, &userDB.Email, &userDB.Avatar, &userDB.Birthday, &userDB.About, &userDB.PublicName, &externalID, &userDB.IsExternalAvatar, &userDB.SubscriberCount)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	} else if err != nil {
		return domain.User{}, err
	}

	user := domain.User{
		ID:         userDB.ID,
		Username:   userDB.Username,
		Email:      userDB.Email,
		Avatar:     userDB.Avatar.String,
		Birthday:   userDB.Birthday.Time,
		PublicName: userDB.PublicName,
		About:      userDB.About.String,
		IsExternal: externalID.String != "",
		IsExternalAvatar: userDB.IsExternalAvatar.Bool,
		SubscriberCount: int(userDB.SubscriberCount.Int64),
	}

	return user, nil
}

func (p *pgProfileStorage) GetUserPublicInfoByUsername(username string) (domain.User, error) {
	var userDB userDB

	err := p.db.QueryRow(`
		SELECT id, username, email, avatar, birthday, about, public_name, is_external_avatar, subscriber_count
		FROM flow_user WHERE username = $1
	`, username).Scan(&userDB.ID, &userDB.Username, &userDB.Email, &userDB.Avatar, &userDB.Birthday, &userDB.About, &userDB.PublicName, &userDB.IsExternalAvatar, &userDB.SubscriberCount)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	} else if err != nil {
		return domain.User{}, err
	}

	user := domain.User{
		ID:               userDB.ID,
		Username:         userDB.Username,
		Email:            userDB.Email,
		Avatar:           userDB.Avatar.String,
		Birthday:         userDB.Birthday.Time,
		PublicName:       userDB.PublicName,
		About:            userDB.About.String,
		IsExternalAvatar: userDB.IsExternalAvatar.Bool,
		SubscriberCount:  int(userDB.SubscriberCount.Int64),
	}

	return user, nil
}

func (p *pgProfileStorage) SaveUserAvatar(email string, avatar string) error {
	_, err := p.db.Exec(`
		UPDATE flow_user SET avatar = $1,
		is_external_avatar = false
		WHERE email = $2
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

func (p *pgProfileStorage) GetHashedPassword(email string) (string, error) {
	var password string
	err := p.db.QueryRow(`
	SELECT password
	FROM flow_user
	WHERE email = $1`, email).Scan(&password)
	if err != nil {
		return "", err
	}

	return password, nil
}

func (p *pgProfileStorage) SetNewPassword(email string, newPassword string) (int, error) {
	var id int
	err := p.db.QueryRow(`
	UPDATE flow_user
	SET password = $1
	WHERE email = $2
	RETURNING id`, newPassword, email).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}
