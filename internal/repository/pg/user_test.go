package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
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

    userInfo := domain.User{
        Username:  "user1",
        Avatar:    "avatar_url",
        PublicName: "user1",
        Email:     "user@example.com",
        Password:  "pass123",
    }

    mock.ExpectQuery(`
    WITH conflict_check AS \(
        SELECT id
        FROM flow_user
        WHERE username = \$1 OR email = \$4
    \)
    INSERT INTO flow_user \(username, avatar, public_name, email, password\)
    SELECT \$1, \$2, \$3, \$4, \$5
    WHERE NOT EXISTS \(SELECT 1 FROM conflict_check\)
    RETURNING id;
        `).
        WithArgs(userInfo.Username, userInfo.Avatar, userInfo.Username, userInfo.Email, userInfo.Password).
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
    username := "coolusername"
    id := uint64(1)

    rows := sqlmock.NewRows([]string{"id", "password", "username"}).
        AddRow(id, password, username)

    mock.ExpectQuery("SELECT id, password, username FROM flow_user WHERE email = \\$1").
        WithArgs(email).
        WillReturnRows(rows)

    returnedID, returnedHash, _, err := storage.GetHash(context.Background(), email, "")
    assert.NoError(t, err)
    assert.Equal(t, id, returnedID)
    assert.Equal(t, password, returnedHash)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetHash_UserNotFound(t *testing.T) {
    mock, storage := setupMock(t)
    defer mock.ExpectClose()

    email := "user@example.com"

    mock.ExpectQuery("SELECT id, password, username FROM flow_user WHERE email = \\$1").
        WithArgs(email).
        WillReturnError(sql.ErrNoRows)

    _, _, _, err := storage.GetHash(context.Background(), email, "")
    assert.Equal(t, domain.ErrInvalidCredentials, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserPublicInfo_Success(t *testing.T) {
    mock, storage := setupMock(t)
    defer mock.ExpectClose()

    email := "user@example.com"
    username := "user1"
    avatar := "avatar.jpg"
    birthday := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

    rows := sqlmock.NewRows([]string{"username", "email", "avatar", "birthday", "about", "public_name", "subscriber_count"}).
        AddRow(username, email, avatar, birthday, "", "", 0)

    mock.ExpectQuery(`SELECT username, email, avatar, birthday, about, public_name, subscriber_count FROM flow_user WHERE email = \$1`).
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

    rows := sqlmock.NewRows([]string{"username", "email", "avatar", "birthday", "about", "public_name", "subscriber_count"}).
        AddRow(username, email, avatar, birthday, "", "", 0)

    mock.ExpectQuery(`SELECT username, email, avatar, birthday, about, public_name, subscriber_count FROM flow_user WHERE email = \$1`).
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

    mock.ExpectQuery(`SELECT username, email, avatar, birthday, about, public_name, subscriber_count FROM flow_user WHERE email = \$1`).
        WithArgs(email).
        WillReturnError(sql.ErrNoRows)

    _, err := storage.GetUserPublicInfo(context.Background(), email)
    assert.Equal(t, domain.ErrInvalidCredentials, err)
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
    assert.Equal(t, domain.ErrUserNotFound, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

