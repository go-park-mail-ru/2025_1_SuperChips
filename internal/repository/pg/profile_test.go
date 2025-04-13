package repository_test

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	pg "github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/pg"
)

func TestGetUserPublicInfoByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	email := "test@example.com"
	rows := sqlmock.NewRows([]string{"id", "username", "email", "avatar", "birthday", "about", "public_name"}).
		AddRow(1, "cool_user", email, "avatar_url", nil, nil, "Cool User")

	mock.ExpectQuery(`SELECT id, username, email, avatar, birthday, about, public_name
		FROM flow_user WHERE email = ?`).
		WithArgs(email).
		WillReturnRows(rows)

	repo, err := pg.NewPGProfileStorage(db, "", "", "")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	user, err := repo.GetUserPublicInfoByEmail(email)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if user.Email != email {
		t.Errorf("expected email %v, got %v", email, user.Email)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetUserPublicInfoByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	defer db.Close()

	username := "cool_user"
	rows := sqlmock.NewRows([]string{"id", "username", "email", "avatar", "birthday", "about", "public_name"}).
		AddRow(1, username, "cool_user@yandex.ru", "avatar_url", nil, nil, "Cool User")

	mock.ExpectQuery(`SELECT id, username, email, avatar, birthday, about, public_name
		FROM flow_user WHERE username = ?`).
		WithArgs(username).
		WillReturnRows(rows)

	repo, err := pg.NewPGProfileStorage(db, "", "", "")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	user, err := repo.GetUserPublicInfoByUsername(username)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if user.Username != username {
		t.Errorf("expected username %v, got %v", username, user.Username)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestSaveUserAvatar(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	email := "hello@mail.ru"
	avatarURL := "new_avatar_url"

	mock.ExpectExec(`UPDATE flow_user SET avatar = \$1 WHERE email = \$2`).
		WithArgs(avatarURL, email).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo, err := pg.NewPGProfileStorage(db, "", "", "")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	err = repo.SaveUserAvatar(email, avatarURL)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestUpdateUserData(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	defer db.Close()

	oldEmail := "not_so_cool_email@notcool.ru"

	user := domain.User{
		Username:   "cool_guy",
		Birthday:   time.Date(2000, time.April, 1, 24, 24, 24, 0, time.UTC),
		About:      "",
		PublicName: "very cool guy",
		Email:      "cool_email@cool.ru",
	}

	mock.ExpectExec(`UPDATE flow_user
	SET
	username = \$1,
	birthday = \$2,
	about = \$3,
	public_name = \$4,
	email = \$5
	WHERE email = \$6`).
		WithArgs(user.Username, user.Birthday, user.About, user.PublicName, user.Email, oldEmail).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo, err := pg.NewPGProfileStorage(db, "", "", "")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	err = repo.UpdateUserData(user, oldEmail)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetHashedPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.NewRows([]string{"password", "email"}).AddRow("verymegahash", "cool_email@email.ru")

	email := "cool_email@email.ru"

	mock.ExpectQuery(`SELECT password FROM flow_user WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(mock.NewRows([]string{"password"}).AddRow("verymegahash"))

	repo, err := pg.NewPGProfileStorage(db, "", "", "")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	password, err := repo.GetHashedPassword(email)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if password != "verymegahash" {
		t.Errorf("expected password %v, got %v", "verymegahash", password)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestSetNewPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	defer db.Close()

	password := "mega_hash"
	email := "cool@mail.ru"

	mock.NewRows([]string{"id", "password", "email"}).AddRow(1, password, email)

	mock.ExpectQuery(`UPDATE flow_user SET password = \$1 WHERE email = \$2
	RETURNING id`).
	WithArgs(password, email).
	WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))

	repo, err := pg.NewPGProfileStorage(db, "", "", "")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	id, err := repo.SetNewPassword(email, password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != 1 {
		t.Errorf("expected id %v, got %v", 1, id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}