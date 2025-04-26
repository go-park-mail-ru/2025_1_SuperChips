package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2025_1_SuperChips/auth_service"
	"github.com/stretchr/testify/assert"
)

func setupMock(t *testing.T) (sqlmock.Sqlmock, *pgUserStorage) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    storage := &pgUserStorage{
        db: db,
    }
    return mock, storage
}

func TestAddUser_Success(t *testing.T) {
    mock, storage := setupMock(t)
    defer mock.ExpectClose()

    userInfo := models.User{
        Username:  "user1",
        Avatar:    "avatar_url",
        PublicName: "user1",
        Email:     "user@example.com",
        Password:  "pass123",
        Birthday:  time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
    }

    mock.ExpectQuery("INSERT INTO flow_user").
        WithArgs(userInfo.Username, userInfo.Avatar, userInfo.Username, userInfo.Email, userInfo.Password, userInfo.Birthday).
        WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

    id, err := storage.AddUser(context.Background(), userInfo)
    assert.NoError(t, err)
    assert.Equal(t, uint64(1), id)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetHash_Success(t *testing.T) {
    mock, storage := setupMock(t)
    defer mock.ExpectClose()

    email := "user@example.com"
    password := "hashedpassword"
    id := uint64(1)

    rows := sqlmock.NewRows([]string{"id", "password"}).
        AddRow(id, password)

    mock.ExpectQuery("SELECT id, password FROM flow_user WHERE email =").
        WithArgs(email).
        WillReturnRows(rows)

    returnedID, returnedHash, err := storage.GetHash(context.Background(), email, "")
    assert.NoError(t, err)
    assert.Equal(t, id, returnedID)
    assert.Equal(t, password, returnedHash)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetHash_UserNotFound(t *testing.T) {
    mock, storage := setupMock(t)
    defer mock.ExpectClose()

    email := "user@example.com"

    mock.ExpectQuery("SELECT id, password FROM flow_user WHERE email =").
        WithArgs(email).
        WillReturnError(sql.ErrNoRows)

    _, _, err := storage.GetHash(context.Background(), email, "")
    assert.Equal(t, models.ErrInvalidCredentials, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserPublicInfo_Success(t *testing.T) {
    mock, storage := setupMock(t)
    defer mock.ExpectClose()

    email := "user@example.com"
    username := "user1"
    avatar := "avatar.jpg"
    birthday := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

    rows := sqlmock.NewRows([]string{"username", "email", "avatar", "birthday"}).
        AddRow(username, email, avatar, birthday)

    mock.ExpectQuery(`SELECT username, email, avatar, birthday, about, public_name, FROM flow_user WHERE email = \$1`).
        WithArgs(email).
        WillReturnRows(rows)

    publicUser, err := storage.GetUserPublicInfo(context.Background(), email)
    assert.NoError(t, err)
    assert.Equal(t, username, publicUser.Username)
    assert.Equal(t, email, publicUser.Email)
    assert.Equal(t, avatar, publicUser.Avatar)
    assert.Equal(t, birthday, publicUser.Birthday)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserPublicInfo_NullAvatar(t *testing.T) {
    mock, storage := setupMock(t)
    defer mock.ExpectClose()

    email := "user@example.com"
    username := "user1"
    var avatar sql.NullString
    birthday := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

    rows := sqlmock.NewRows([]string{"username", "email", "avatar", "birthday"}).
        AddRow(username, email, avatar, birthday)

    mock.ExpectQuery(`SELECT username, email, avatar, birthday, about, public_name, FROM flow_user WHERE email = \$1`).
        WithArgs(email).
        WillReturnRows(rows)

    publicUser, err := storage.GetUserPublicInfo(context.Background(), email)
    assert.NoError(t, err)
    assert.Equal(t, username, publicUser.Username)
    assert.Equal(t, email, publicUser.Email)
    assert.Equal(t, "", publicUser.Avatar)
    assert.Equal(t, birthday, publicUser.Birthday)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserPublicInfo_UserNotFound(t *testing.T) {
    mock, storage := setupMock(t)
    defer mock.ExpectClose()

    email := "user@example.com"

    mock.ExpectQuery(`SELECT username, email, avatar, birthday, about, public_name, FROM flow_user WHERE email = \$1`).
        WithArgs(email).
        WillReturnError(sql.ErrNoRows)

    _, err := storage.GetUserPublicInfo(context.Background(), email)
    assert.Equal(t, models.ErrInvalidCredentials, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserId_Success(t *testing.T) {
    mock, storage := setupMock(t)
    defer mock.ExpectClose()

    email := "user@example.com"
    id := uint64(1)

    rows := sqlmock.NewRows([]string{"id"}).AddRow(id)

    mock.ExpectQuery("SELECT id FROM flow_user WHERE email =").
        WithArgs(email).
        WillReturnRows(rows)

    returnedID, err := storage.GetUserId(context.Background(), email)
    assert.NoError(t, err)
    assert.Equal(t, id, returnedID)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserId_UserNotFound(t *testing.T) {
    mock, storage := setupMock(t)
    defer mock.ExpectClose()

    email := "user@example.com"

    mock.ExpectQuery("SELECT id FROM flow_user WHERE email =").
        WithArgs(email).
        WillReturnError(sql.ErrNoRows)

    _, err := storage.GetUserId(context.Background(), email)
    assert.Equal(t, models.ErrUserNotFound, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

