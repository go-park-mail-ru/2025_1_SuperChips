package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSearchPins(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    repo := NewSearchRepository(db)

    t.Run("Success", func(t *testing.T) {
        ctx := context.Background()
        query := "test"
        page := 1
        pageSize := 10
        offset := (page - 1) * pageSize

        mock.ExpectQuery(regexp.QuoteMeta(
            `SELECT f.id, f.title, f.description, f.author_id, f.is_private, f.media_url, f.width, f.height, f.is_nsfw, fu.username 
            FROM flow f JOIN flow_user fu ON f.author_id = fu.id 
            WHERE f.is_private = false AND 
            (to_tsvector(f.title || ' ' || f.description) @@ plainto_tsquery($1) OR f.title ILIKE '%' || $1 || '%' OR f.description ILIKE '%' || $1 || '%') 
            ORDER BY f.like_count DESC LIMIT $2 OFFSET $3`,
        )).WithArgs(query, pageSize, offset).
            WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "author_id", "is_private", "media_url", "width", "height", "is_nsfw", "username"}).
                AddRow(1, "Pin 1", "Description 1", 101, false, "http://example.com/image1.jpg", 800, 600, false, "user1").
                AddRow(2, "Pin 2", "Description 2", 102, false, "http://example.com/image2.jpg", 1024, 768, true, "user2"))

        pins, err := repo.SearchPins(ctx, query, page, pageSize)

        assert.NoError(t, err)
        assert.Len(t, pins, 2)
        assert.Equal(t, uint64(1), pins[0].FlowID)
        assert.Equal(t, "Pin 1", pins[0].Header)
        assert.Equal(t, "Description 1", pins[0].Description)
        assert.Equal(t, uint64(101), pins[0].AuthorID)
        assert.False(t, pins[0].IsPrivate)
        assert.Equal(t, "http://example.com/image1.jpg", pins[0].MediaURL)
        assert.Equal(t, 800, pins[0].Width)
        assert.Equal(t, 600, pins[0].Height)
        assert.False(t, pins[0].IsNSFW)
        assert.Equal(t, "user1", pins[0].AuthorUsername)
        assert.NoError(t, mock.ExpectationsWereMet())
    })

    t.Run("EmptyResult", func(t *testing.T) {
        ctx := context.Background()
        query := "nonexistent"
        page := 1
        pageSize := 10
        offset := (page - 1) * pageSize

        mock.ExpectQuery(regexp.QuoteMeta(
            `SELECT f.id, f.title, f.description, f.author_id, f.is_private, f.media_url, f.width, f.height, f.is_nsfw, fu.username 
            FROM flow f JOIN flow_user fu ON f.author_id = fu.id 
            WHERE f.is_private = false AND 
            (to_tsvector(f.title || ' ' || f.description) @@ plainto_tsquery($1) OR f.title ILIKE '%' || $1 || '%' OR f.description ILIKE '%' || $1 || '%') 
            ORDER BY f.like_count DESC LIMIT $2 OFFSET $3`,
        )).WithArgs(query, pageSize, offset).
            WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "author_id", "is_private", "media_url", "width", "height", "is_nsfw", "username"}))

        pins, err := repo.SearchPins(ctx, query, page, pageSize)

        assert.NoError(t, err)
        assert.Empty(t, pins)
        assert.NoError(t, mock.ExpectationsWereMet())
    })

    t.Run("DatabaseError", func(t *testing.T) {
        ctx := context.Background()
        query := "test"
        page := 1
        pageSize := 10
        offset := (page - 1) * pageSize

        mock.ExpectQuery(regexp.QuoteMeta(
            `SELECT f.id, f.title, f.description, f.author_id, f.is_private, f.media_url, f.width, f.height, f.is_nsfw, fu.username 
            FROM flow f JOIN flow_user fu ON f.author_id = fu.id 
            WHERE f.is_private = false AND 
            (to_tsvector(f.title || ' ' || f.description) @@ plainto_tsquery($1) OR f.title ILIKE '%' || $1 || '%' OR f.description ILIKE '%' || $1 || '%') 
            ORDER BY f.like_count DESC LIMIT $2 OFFSET $3`,
        )).WithArgs(query, pageSize, offset).
            WillReturnError(errors.New("database error"))

        pins, err := repo.SearchPins(ctx, query, page, pageSize)

        assert.Error(t, err)
        assert.Empty(t, pins)
        assert.NoError(t, mock.ExpectationsWereMet())
    })
}

func TestSearchUsers(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    repo := NewSearchRepository(db)

    t.Run("Success", func(t *testing.T) {
        ctx := context.Background()
        query := "user"
        page := 1
        pageSize := 10
        offset := (page - 1) * pageSize

        mock.ExpectQuery(regexp.QuoteMeta(
            `SELECT username, email, avatar, birthday, about, public_name, is_external_avatar, subscriber_count 
            FROM flow_user 
            WHERE to_tsvector(username) @@ plainto_tsquery($1) OR username ILIKE '%' || $1 || '%' LIMIT $2 OFFSET $3`,
        )).WithArgs(query, pageSize, offset).
            WillReturnRows(sqlmock.NewRows([]string{"username", "email", "avatar", "birthday", "about", "public_name", "is_external_avatar", "subscriber_count"}).
                AddRow("user1", "user1@example.com", "http://example.com/avatar1.jpg", time.Now(), "About user 1", "Public User 1", true, 100).
                AddRow("user2", "user2@example.com", "http://example.com/avatar2.jpg", time.Now(), "About user 2", "Public User 2", false, 50))

        users, err := repo.SearchUsers(ctx, query, page, pageSize)
        assert.NoError(t, err)
        assert.Len(t, users, 2)

        assert.Equal(t, "user1", users[0].Username)
        assert.Equal(t, "user1@example.com", users[0].Email)
        assert.Equal(t, "http://example.com/avatar1.jpg", users[0].Avatar)
        assert.Equal(t, "About user 1", users[0].About)
        assert.Equal(t, "Public User 1", users[0].PublicName)
        assert.True(t, users[0].IsExternalAvatar)
        assert.Equal(t, 100, users[0].SubscriberCount)

        assert.NoError(t, mock.ExpectationsWereMet())
    })

    t.Run("EmptyResult", func(t *testing.T) {
        ctx := context.Background()
        query := "nonexistent"
        page := 1
        pageSize := 10
        offset := (page - 1) * pageSize

        mock.ExpectQuery(regexp.QuoteMeta(
            `SELECT username, email, avatar, birthday, about, public_name, is_external_avatar, subscriber_count 
            FROM flow_user 
            WHERE to_tsvector(username) @@ plainto_tsquery($1) OR username ILIKE '%' || $1 || '%' LIMIT $2 OFFSET $3`,
        )).WithArgs(query, pageSize, offset).
            WillReturnRows(sqlmock.NewRows([]string{"username", "email", "avatar", "birthday", "about", "public_name", "is_external_avatar", "subscriber_count"}))

        users, err := repo.SearchUsers(ctx, query, page, pageSize)
        assert.NoError(t, err)
        assert.Empty(t, users)

        assert.NoError(t, mock.ExpectationsWereMet())
    })

    t.Run("DatabaseError", func(t *testing.T) {
        ctx := context.Background()
        query := "user"
        page := 1
        pageSize := 10
        offset := (page - 1) * pageSize

        mock.ExpectQuery(regexp.QuoteMeta(
            `SELECT username, email, avatar, birthday, about, public_name, is_external_avatar, subscriber_count 
            FROM flow_user 
            WHERE to_tsvector(username) @@ plainto_tsquery($1) 
            OR username ILIKE '%' || $1 || '%' LIMIT $2 OFFSET $3`,
        )).WithArgs(query, pageSize, offset).
            WillReturnError(errors.New("database error"))

        users, err := repo.SearchUsers(ctx, query, page, pageSize)
        assert.Error(t, err)
        assert.Empty(t, users)

        assert.NoError(t, mock.ExpectationsWereMet())
    })
}

func TestSearchBoards(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    repo := NewSearchRepository(db)

    t.Run("Success", func(t *testing.T) {
        ctx := context.Background()
        query := "board"
        page := 1
        pageSize := 10
        previewNum := 3
        previewStart := 0
        offset := (page - 1) * pageSize

        mock.ExpectQuery(regexp.QuoteMeta(
            `SELECT board.id, board.author_id, board.board_name, board.created_at, board.is_private, board.flow_count, flow_user.username 
			FROM board 
			INNER JOIN flow_user ON board.author_id = flow_user.id 
			WHERE board.is_private = false 
			AND ( board.board_name ILIKE '%' || $1 || '%' OR to_tsvector(board.board_name) @@ plainto_tsquery($1) ) LIMIT $2 OFFSET $3`,
		)).WithArgs(query, pageSize, offset).
            WillReturnRows(sqlmock.NewRows([]string{"id", "author_id", "board_name", "created_at", "is_private", "flow_count", "username"}).
                AddRow(1, 101, "Board 1", time.Now(), false, 5, "user1").
                AddRow(2, 102, "Board 2", time.Now(), false, 3, "user2"))

        mock.ExpectQuery(regexp.QuoteMeta(
            `SELECT f.id, f.title, f.description, f.author_id, f.created_at, f.updated_at, f.is_private, f.media_url, f.like_count 
			FROM flow f 
			JOIN board_post bp 
			ON f.id = bp.flow_id 
			WHERE bp.board_id = $1 
			AND (f.is_private = false OR f.author_id = $2) 
			ORDER BY bp.saved_at DESC LIMIT $3 OFFSET $4`,
        )).WithArgs(1, 0, previewNum, previewStart).
            WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "author_id", "created_at", "updated_at", "is_private", "media_url", "like_count"}).
                AddRow(101, "Flow 1", "Description 1", 101, time.Now(), time.Now(), false, "http://example.com/flow1.jpg", 10)).
            WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "author_id", "created_at", "updated_at", "is_private", "media_url", "like_count"}).
                AddRow(102, "Flow 2", "Description 2", 102, time.Now(), time.Now(), false, "http://example.com/flow2.jpg", 5))

        boards, err := repo.SearchBoards(ctx, query, page, pageSize, previewNum, previewStart)
        assert.NoError(t, err)
        assert.Len(t, boards, 2)

        assert.Equal(t, int(1), boards[0].ID)
        assert.Equal(t, int(101), boards[0].AuthorID)
        assert.Equal(t, "Board 1", boards[0].Name)
        assert.False(t, boards[0].IsPrivate)
        assert.Equal(t, 5, boards[0].FlowCount)
        assert.Equal(t, "user1", boards[0].AuthorUsername)
        assert.Len(t, boards[0].Preview, 1)

        assert.NoError(t, mock.ExpectationsWereMet())
    })

    t.Run("EmptyResult", func(t *testing.T) {
        ctx := context.Background()
        query := "nonexistent"
        page := 1
        pageSize := 10
        previewNum := 3
        previewStart := 0
        offset := (page - 1) * pageSize

        mock.ExpectQuery(regexp.QuoteMeta(
            `SELECT board.id, board.author_id, board.board_name, board.created_at, board.is_private, board.flow_count, flow_user.username 
			FROM board INNER JOIN flow_user 
			ON board.author_id = flow_user.id 
			WHERE board.is_private = false AND ( board.board_name ILIKE '%' || $1 || '%' OR to_tsvector(board.board_name) @@ plainto_tsquery($1) ) LIMIT $2 OFFSET $3`,
        )).WithArgs(query, pageSize, offset).
            WillReturnRows(sqlmock.NewRows([]string{"id", "author_id", "board_name", "created_at", "is_private", "flow_count", "username"}))

        boards, err := repo.SearchBoards(ctx, query, page, pageSize, previewNum, previewStart)
        assert.NoError(t, err)
        assert.Empty(t, boards)

        assert.NoError(t, mock.ExpectationsWereMet())
    })
}
