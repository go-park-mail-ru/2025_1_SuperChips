package repository_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	pg "github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/pg"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/pg/mocks"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/security"
)

func TestUserRepository_AddUser(t *testing.T) {
	userMock, err := mocks.NewUserMock()
	if err != nil {
		t.Fatal(err)
	}

	db, mock := userMock.Db, userMock.Mock

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

	mock.ExpectQuery(`SELECT id FROM flow_user WHERE email = \$1 OR username = \$2`).
		WithArgs(user.Email, user.Username).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(`INSERT INTO flow_user`).
		WithArgs(user.Username, user.Avatar, user.PublicName, user.Email, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.AddUser(user)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestUserRepository_LoginUser(t *testing.T) {
	userMock, err := mocks.NewUserMock()
	if err != nil {
		t.Fatal(err)
	}

	db, mock := userMock.Db, userMock.Mock

	defer db.Close()

	repo, err := pg.NewPGUserStorage(db)
	if err != nil {
		t.Fatal(err)
	}

	email := "test@example.com"
	password := "password123"
	hashedPassword, err := security.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(`SELECT password FROM flow_user WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"password"}).AddRow(hashedPassword))

	pswd, err := repo.LoginUser(email, password)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}

	if !security.ComparePassword(password, pswd) {
		t.Errorf("passwords dont match")
	}
}

func TestUserRepository_GetUserPublicInfo(t *testing.T) {
	userMock, err := mocks.NewUserMock()
	if err != nil {
		t.Fatal(err)
	}

	db, mock := userMock.Db, userMock.Mock

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

	mock.ExpectQuery(`SELECT username, email, avatar, birthday FROM flow_user WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"username", "email", "avatar", "birthday"}).
			AddRow(mockUser.Username, mockUser.Email, mockUser.Avatar, nil))

	user, err := repo.GetUserPublicInfo(email)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if user.Username != mockUser.Username || user.Email != mockUser.Email || user.Avatar != mockUser.Avatar {
		t.Errorf("unexpected user data: got %+v, want %+v", user, mockUser)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestUserRepository_GetUserId(t *testing.T) {
	userMock, err := mocks.NewUserMock()
	if err != nil {
		t.Fatal(err)
	}

	db, mock := userMock.Db, userMock.Mock

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
	userMock, err := mocks.NewUserMock()
	if err != nil {
		t.Fatal(err)
	}

	db, mock := userMock.Db, userMock.Mock

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

	mock.ExpectQuery(`SELECT id FROM flow_user WHERE email = \$1 OR username = \$2`).
		WithArgs(user.Email, user.Username).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

	err = repo.AddUser(user)
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("expected conflict error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}
