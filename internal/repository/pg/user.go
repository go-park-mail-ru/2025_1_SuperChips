package repository

import (
	"database/sql"

	user "github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	security "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/security"
)

type pgUserStorage struct {
	db *sql.DB
}

// AddUser(user domain.User) error
// LoginUser(email, password string) error
// GetUserPublicInfo(email string) (domain.PublicUser, error)
// GetUserId(email string) uint64	

const (
	CREATE_USER_TABLE = `
		CREATE TABLE IF NOT EXISTS user (
		user_id INTEGER PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		avatar TEXT,
		public_name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		create_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		password TEXT NOT NULL,
		birthday DATE,
		about TEXT,
		jwt_version INTEGER NOT NULL DEFAULT 1
		`
)

func NewPGUserStorage(db *sql.DB) (*pgUserStorage, error) {
	storage := &pgUserStorage{
		db: db,
	}

	storage.initialize()

	return storage, nil
}

func (p *pgUserStorage) initialize() error {
	_, err := p.db.Exec(CREATE_USER_TABLE)
	if err != nil {
		return err
	}

	return nil
}

func (p *pgUserStorage) AddUser(user user.User) error {
	if err := user.ValidateUser(); err != nil {
		return err
	}

	hashedPassword, err := security.HashPassword(user.Password)
	if err != nil {
		return err
	}

	_, err = p.db.Exec(`
		INSERT INTO users (username, avatar, public_name, email, password)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, user.Username, user.Avatar, user.Username, user.Email, hashedPassword)
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

	err := p.db.QueryRow(`
		SELECT password FROM users WHERE email = $1
	`, email).Scan(&hashedPassword)
	if err != nil {
		return user.ErrInvalidCredentials
	}

	ok := security.ComparePassword(hashedPassword, password)
	if !ok {
		return user.ErrInvalidCredentials
	}

	return nil
}

func (p *pgUserStorage) GetUserPublicInfo(email string) (user.PublicUser, error) {
	var publicUser user.PublicUser

	err := p.db.QueryRow(`
		SELECT username, email, avatar, birthday FROM users WHERE email = $1
	`, email).Scan(&publicUser.Username, &publicUser.Email, &publicUser.Avatar, &publicUser.Birthday)
	if err != nil {
		return user.PublicUser{}, user.ErrUserNotFound
	}

	return publicUser, nil
}

func (p *pgUserStorage) GetUserId(email string) uint64 {
	var id uint64

	err := p.db.QueryRow(`
		SELECT user_id FROM users WHERE email = $1
	`, email).Scan(&id)
	if err != nil {
		return 0
	}

	return id
}
