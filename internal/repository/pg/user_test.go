package repository_test

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	pg "github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/pg"
)

func TestUserRepository_AddUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	repo, err := pg.NewPGUserStorage(db)
	if err != nil {
		t.Fatal(err)
	}

	user := domain.User{
		Username:   "testuser",
		Email:      "test@example.com",
		Password:   "password123",
		PublicName: "Test User",
		Birthday:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	expectedId := uint64(12)

	mock.ExpectQuery(`INSERT INTO flow_user \(username, avatar, public_name, email, password, birthday\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\) ON CONFLICT \(email, username\) DO NOTHING RETURNING id`).
		WithArgs(user.Username, "", user.PublicName, user.Email, user.Password, user.Birthday).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedId))

	_, err = repo.AddUser(user)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestUserRepository_LoginUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	repo, err := pg.NewPGUserStorage(db)
	if err != nil {
		t.Fatal(err)
	}

	email := "test@example.com"
	password := "password123"
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(`SELECT id, password FROM flow_user WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "password"}).AddRow(2, password))

	_, pswd, err := repo.GetHash(email, password)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}

	if password != pswd {
		t.Errorf("passwords dont match")
	}
}

func TestUserRepository_GetUserPublicInfo(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	repo, err := pg.NewPGUserStorage(db)
	if err != nil {
		t.Fatal(err)
	}

	email := "test@example.com"
	mockUser := domain.PublicUser{
		Username: "testuser",
		Email:    email,
		Avatar:   "avatar.png",
		Birthday: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	mock.ExpectQuery(`SELECT username, email, avatar, birthday, about, public_name, FROM flow_user WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"username", "email", "avatar", "birthday"}).
			AddRow(mockUser.Username, mockUser.Email, mockUser.Avatar, mockUser.Birthday))

	user, err := repo.GetUserPublicInfo(email)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if user.Username != mockUser.Username || user.Email != mockUser.Email || user.Avatar != mockUser.Avatar || user.Birthday != mockUser.Birthday {
		t.Errorf("unexpected user data: got %+v, want %+v", user, mockUser)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestUserRepository_GetUserId(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	repo, err := pg.NewPGUserStorage(db)
	if err != nil {
		t.Fatal(err)
	}

	email := "test@example.com"
	var expectedID uint64 = 1

	mock.ExpectQuery(`SELECT id FROM flow_user WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(expectedID))

	id, err := repo.GetUserId(email)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if id != expectedID {
		t.Errorf("unexpected user ID: got %d, want %d", id, expectedID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestUserRepository_AddUser_Conflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	repo, err := pg.NewPGUserStorage(db)
	if err != nil {
		t.Fatal(err)
	}

	user := domain.User{
		Username:   "testuser",
		Email:      "test@example.com",
		Password:   "password123",
		PublicName: "Test User",
		Birthday:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	mock.NewRows([]string{"id", "username", "avatar", "public_name", "email", "password"}).
	AddRow(1, user.Username, "", user.PublicName, user.Email, user.Password)

	mock.ExpectQuery(`INSERT INTO flow_user \(username, avatar, public_name, email, password, birthday\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\) ON CONFLICT \(email, username\) DO NOTHING RETURNING id`).
		WithArgs(user.Username, "", user.PublicName, user.Email, user.Password, user.Birthday).
		WillReturnError(domain.ErrConflict)

	_, err = repo.AddUser(user)
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("expected conflict error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}
