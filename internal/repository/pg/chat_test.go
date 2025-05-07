package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/stretchr/testify/assert"
)

func TestGetNewMessages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewChatRepository(db)

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		username := "user1"
		offset := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT id, content, timestamp, is_read, sender, recipient, chat_id 
			FROM message WHERE timestamp > $1 AND recipient = $2 AND sent = false`,
		)).WithArgs(offset, username).
			WillReturnRows(sqlmock.NewRows([]string{"id", "content", "timestamp", "is_read", "sender", "recipient", "chat_id"}).
				AddRow(1, "Hello", time.Date(2023, 1, 2, 10, 0, 0, 0, time.UTC), false, "user2", "user1", 101).
				AddRow(2, "Hi", time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC), false, "user2", "user1", 101))

		messages, err := repo.GetNewMessages(ctx, username, offset)
		assert.NoError(t, err)
		assert.Len(t, messages, 2)

		assert.Equal(t, uint(1), messages[0].MessageID)
		assert.Equal(t, "Hello", messages[0].Content)
		assert.Equal(t, "user2", messages[0].Sender)
		assert.Equal(t, "user1", messages[0].Recipient)
		assert.Equal(t, uint64(101), messages[0].ChatID)
		assert.False(t, messages[0].IsRead)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("EmptyResult", func(t *testing.T) {
		ctx := context.Background()
		username := "user1"
		offset := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT id, content, timestamp, is_read, sender, recipient, chat_id 
			FROM message WHERE timestamp > $1 AND recipient = $2 AND sent = false`,
		)).WithArgs(offset, username).
			WillReturnRows(sqlmock.NewRows([]string{"id", "content", "timestamp", "is_read", "sender", "recipient", "chat_id"}))

		messages, err := repo.GetNewMessages(ctx, username, offset)
		assert.NoError(t, err)
		assert.Empty(t, messages)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		ctx := context.Background()
		username := "user1"
		offset := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT id, content, timestamp, is_read, sender, recipient, chat_id 
			FROM message WHERE timestamp > $1 AND recipient = $2 AND sent = false`,
		)).WithArgs(offset, username).
			WillReturnError(errors.New("database error"))

		messages, err := repo.GetNewMessages(ctx, username, offset)
		assert.Error(t, err)
		assert.Empty(t, messages)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAddMessage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewChatRepository(db)

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		message := domain.Message{
			Content:   "Hello",
			Sender:    "user1",
			Recipient: "user2",
			ChatID:    101,
			Sent:      true,
		}

		mock.ExpectExec(regexp.QuoteMeta(
			`INSERT INTO message (content, sender, recipient, chat_id, sent) 
			SELECT $1, $2, $3, $4, $5 
			WHERE EXISTS ( SELECT 1 FROM chat WHERE id = $4 
			AND (($2 = user1 AND $3 = user2) OR ($2 = user2 AND $3 = user1)) )`,
		)).WithArgs(message.Content, message.Sender, message.Recipient, message.ChatID, message.Sent).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.AddMessage(ctx, message)
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		ctx := context.Background()
		message := domain.Message{
			Content:   "Hello",
			Sender:    "user1",
			Recipient: "user2",
			ChatID:    101,
			Sent:      true,
		}

		mock.ExpectExec(regexp.QuoteMeta(
			`INSERT INTO message (content, sender, recipient, chat_id, sent) 
			SELECT $1, $2, $3, $4, $5 WHERE EXISTS
			( SELECT 1 FROM chat WHERE id = $4 AND (($2 = user1 AND $3 = user2) OR ($2 = user2 AND $3 = user1)) )`,
		)).WithArgs(message.Content, message.Sender, message.Recipient, message.ChatID, message.Sent).
			WillReturnError(errors.New("database error"))

		err := repo.AddMessage(ctx, message)
		assert.Error(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestMarkRead(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewChatRepository(db)

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		messageID := 1
		chatID := 101

		mock.ExpectExec(regexp.QuoteMeta(
			`UPDATE message SET is_read = true 
			WHERE chat_id = $1 AND id <= $2`,
		)).WithArgs(chatID, messageID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.MarkRead(ctx, messageID, chatID)
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		ctx := context.Background()
		messageID := 1
		chatID := 101

		mock.ExpectExec(regexp.QuoteMeta(
			`UPDATE message SET is_read = true 
			WHERE chat_id = $1 AND id <= $2`,
		)).WithArgs(chatID, messageID).
			WillReturnError(errors.New("database error"))

		err := repo.MarkRead(ctx, messageID, chatID)
		assert.Error(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetChats(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewChatRepository(db)

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		username := "user1"

		mock.ExpectQuery(regexp.QuoteMeta(`
		WITH unread_counts AS
		( SELECT chat_id, COUNT(*) FILTER (WHERE is_read = FALSE AND recipient = $1) AS unread_count FROM message GROUP BY chat_id ), 
		last_message AS ( SELECT DISTINCT ON (m.chat_id) m.chat_id, m.id AS message_id, m.content, m.sender, m.recipient, m.timestamp, m.is_read FROM message m ORDER BY m.chat_id, m.timestamp DESC ) 
		SELECT c.id AS chat_id, CASE WHEN c.user1 = $1 THEN c.user2 
		ELSE c.user1 
		END AS other_user_username, u.public_name 
		AS other_user_name, u.avatar AS other_user_avatar, u.is_external_avatar, lm.message_id, lm.content 
		AS message_content, lm.sender AS message_sender, lm.recipient, lm.timestamp 
		AS message_timestamp, lm.is_read 
		AS message_is_read, uc.unread_count 
		FROM chat c JOIN flow_user u ON u.username = CASE 
		WHEN c.user1 = $1 THEN c.user2 ELSE c.user1 END LEFT JOIN last_message lm ON c.id = lm.chat_id 
		LEFT JOIN unread_counts uc ON c.id = uc.chat_id WHERE c.user1 = $1 OR c.user2 = $1 ORDER BY lm.timestamp DESC NULLS LAST;`)).
			WithArgs(username).
			WillReturnRows(sqlmock.NewRows([]string{"chat_id", "other_user_username", "other_user_name", "other_user_avatar", "is_external_avatar", "message_id", "message_content", "message_sender", "message_recipient", "message_timestamp", "message_is_read", "unread_count"}).
				AddRow(101, "user2", "User Two", "avatar.jpg", true, 1, "Hello", "user2", "test", time.Date(2023, 1, 2, 10, 0, 0, 0, time.UTC), false, 1))

		chats, err := repo.GetChats(ctx, username)
		assert.NoError(t, err)
		assert.Len(t, chats, 1)

		assert.Equal(t, uint(101), chats[0].ChatID)
		assert.Equal(t, "user2", chats[0].Username)
		assert.Equal(t, "User Two", chats[0].PublicName)
		assert.Equal(t, "avatar.jpg", chats[0].Avatar)
		assert.True(t, chats[0].IsExternalAvatar)
		assert.Equal(t, uint(1), chats[0].MessageCount)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("EmptyResult", func(t *testing.T) {
		ctx := context.Background()
		username := "user1"
		mock.ExpectQuery(regexp.QuoteMeta(`
		WITH unread_counts AS
		( SELECT chat_id, COUNT(*) FILTER (WHERE is_read = FALSE AND recipient = $1) AS unread_count FROM message GROUP BY chat_id ), 
		last_message AS ( SELECT DISTINCT ON (m.chat_id) m.chat_id, m.id AS message_id, m.content, m.sender, m.recipient, m.timestamp, m.is_read FROM message m ORDER BY m.chat_id, m.timestamp DESC ) 
		SELECT c.id AS chat_id, CASE WHEN c.user1 = $1 THEN c.user2 
		ELSE c.user1 
		END AS other_user_username, u.public_name 
		AS other_user_name, u.avatar AS other_user_avatar, u.is_external_avatar, lm.message_id, lm.content 
		AS message_content, lm.sender AS message_sender, lm.recipient, lm.timestamp 
		AS message_timestamp, lm.is_read 
		AS message_is_read, uc.unread_count 
		FROM chat c JOIN flow_user u ON u.username = CASE 
		WHEN c.user1 = $1 THEN c.user2 ELSE c.user1 END LEFT JOIN last_message lm ON c.id = lm.chat_id 
		LEFT JOIN unread_counts uc ON c.id = uc.chat_id WHERE c.user1 = $1 OR c.user2 = $1 ORDER BY lm.timestamp DESC NULLS LAST;`)).
			WithArgs(username).
			WillReturnRows(sqlmock.NewRows([]string{"chat_id", "other_user_username", "other_user_name", "other_user_avatar", "is_external_avatar", "message_id", "message_content", "message_sender", "message_timestamp", "message_is_read", "unread_count"}))

		chats, err := repo.GetChats(ctx, username)
		assert.NoError(t, err)
		assert.Empty(t, chats)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCreateChat(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    repo := NewChatRepository(db)

    t.Run("Success", func(t *testing.T) {
        ctx := context.Background()
        username := "user1"
        targetUsername := "user2"

        mock.ExpectQuery(regexp.QuoteMeta(
            `WITH normalized_users AS ( SELECT LEAST($1, $2) AS user1, GREATEST($1, $2) AS user2 ), 
			inserted_chat AS ( INSERT INTO chat (user1, user2) SELECT user1, user2 
			FROM normalized_users ON CONFLICT (user1, user2) DO NOTHING RETURNING id ), existing_chat 
			AS ( SELECT id FROM chat WHERE (user1, user2) = (SELECT user1, user2 FROM normalized_users) ) 
			SELECT COALESCE(ic.id, ec.id) AS chat_id, u.avatar, u.public_name, u.is_external_avatar 
			FROM inserted_chat ic FULL JOIN existing_chat ec ON TRUE JOIN flow_user u ON u.username = $2;`,
        )).WithArgs(targetUsername, username).
            WillReturnRows(sqlmock.NewRows([]string{"chat_id", "avatar", "public_name", "is_external_avatar"}).
                AddRow(101, "avatar.jpg", "Public User 2", true))

        chat, err := repo.CreateChat(ctx, username, targetUsername)
        assert.NoError(t, err)
        assert.Equal(t, uint(101), chat.ChatID)
        assert.Equal(t, "user2", chat.Username)
        assert.Equal(t, "Public User 2", chat.PublicName)
        assert.Equal(t, "avatar.jpg", chat.Avatar)
        assert.True(t, chat.IsExternalAvatar)

        assert.NoError(t, mock.ExpectationsWereMet())
    })

    t.Run("DatabaseError", func(t *testing.T) {
        ctx := context.Background()
        username := "user1"
        targetUsername := "user2"

        mock.ExpectQuery(regexp.QuoteMeta(
            `WITH normalized_users AS ( SELECT LEAST($1, $2) AS user1, GREATEST($1, $2) AS user2 ), 
			inserted_chat AS ( INSERT INTO chat (user1, user2) SELECT user1, user2 
			FROM normalized_users ON CONFLICT (user1, user2) DO NOTHING RETURNING id ), existing_chat 
			AS ( SELECT id FROM chat WHERE (user1, user2) = (SELECT user1, user2 FROM normalized_users) ) 
			SELECT COALESCE(ic.id, ec.id) AS chat_id, u.avatar, u.public_name, u.is_external_avatar 
			FROM inserted_chat ic FULL JOIN existing_chat ec ON TRUE JOIN flow_user u ON u.username = $2;`,
			)).WithArgs(targetUsername, username).
            WillReturnError(errors.New("database error"))

        _, err := repo.CreateChat(ctx, username, targetUsername)
        assert.Error(t, err)

        assert.NoError(t, mock.ExpectationsWereMet())
    })
}

func TestGetContacts(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    repo := NewChatRepository(db)

    t.Run("Success", func(t *testing.T) {
        ctx := context.Background()
        username := "user1"

        mock.ExpectQuery(
            `SELECT u\.username, u\.public_name, u\.avatar, u\.is_external_avatar FROM contact c INNER JOIN flow_user u ON u\.username = c\.contact_username WHERE c\.user_username = \$1`,
        ).WithArgs(username).
            WillReturnRows(sqlmock.NewRows([]string{"username", "public_name", "avatar", "is_external_avatar"}).
                AddRow("user2", "Public User 2", "avatar2.jpg", true).
                AddRow("user3", "Public User 3", "avatar3.jpg", false))

        contacts, err := repo.GetContacts(ctx, username)
        assert.NoError(t, err)
        assert.Len(t, contacts, 2)

        assert.Equal(t, "user2", contacts[0].Username)
        assert.Equal(t, "Public User 2", contacts[0].PublicName)
        assert.Equal(t, "avatar2.jpg", contacts[0].Avatar)
        assert.True(t, contacts[0].IsExternalAvatar)

        assert.Equal(t, "user3", contacts[1].Username)
        assert.Equal(t, "Public User 3", contacts[1].PublicName)
        assert.Equal(t, "avatar3.jpg", contacts[1].Avatar)
        assert.False(t, contacts[1].IsExternalAvatar)

        assert.NoError(t, mock.ExpectationsWereMet())
    })

    t.Run("EmptyResult", func(t *testing.T) {
        ctx := context.Background()
        username := "user1"

        mock.ExpectQuery(
            `SELECT u\.username, u\.public_name, u\.avatar, u\.is_external_avatar FROM contact c INNER JOIN flow_user u ON u\.username = c\.contact_username WHERE c\.user_username = \$1`,
        ).WithArgs(username).
            WillReturnRows(sqlmock.NewRows([]string{"username", "public_name", "avatar", "is_external_avatar"}))

        contacts, err := repo.GetContacts(ctx, username)
        assert.NoError(t, err)
        assert.Empty(t, contacts)

        assert.NoError(t, mock.ExpectationsWereMet())
    })

    t.Run("DatabaseError", func(t *testing.T) {
        ctx := context.Background()
        username := "user1"

        mock.ExpectQuery(
            `SELECT u\.username, u\.public_name, u\.avatar, u\.is_external_avatar FROM contact c INNER JOIN flow_user u ON u\.username = c\.contact_username WHERE c\.user_username = \$1`,
        ).WithArgs(username).
            WillReturnError(errors.New("database error"))

        contacts, err := repo.GetContacts(ctx, username)
        assert.Error(t, err)
        assert.Empty(t, contacts)

        assert.NoError(t, mock.ExpectationsWereMet())
    })
}

func TestGetChat_Success(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    repo := NewChatRepository(db)

    ctx := context.Background()
    id := uint64(101)
    username := "user1"

    mock.ExpectQuery(
        `WITH chat_messages AS \(.+\) SELECT c\.id AS chat_id, CASE WHEN c\.user1 = \$2 THEN c\.user1 ELSE c\.user2 END AS first_user_username, CASE WHEN c\.user1 = \$2 THEN c\.user2 ELSE c\.user1 END AS other_user_username, u\.public_name AS other_user_name, u\.avatar AS other_user_avatar, cm\.message_id, cm\.content AS message_content, cm\.sender AS message_sender, cm\.recipient, cm\.timestamp AS message_timestamp, cm\.is_read AS message_is_read FROM chat c JOIN flow_user u ON u\.username = CASE WHEN c\.user1 = \$2 THEN c\.user2 ELSE c\.user1 END LEFT JOIN chat_messages cm ON c\.id = cm\.chat_id WHERE c\.id = \$1;`,
    ).WithArgs(id, username).
        WillReturnRows(sqlmock.NewRows([]string{
            "chat_id", "first_user_username", "other_user_username", "other_user_name", "other_user_avatar",
            "message_id", "message_content", "message_sender", "message_recipient", "message_timestamp", "message_is_read",
        }).
            AddRow(101, "user1", "user2", "Public User 2", "avatar.jpg", 1, "Hello", "user2", "user1", time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC), true))

    chat, err := repo.GetChat(ctx, id, username)
    assert.NoError(t, err)
    assert.Equal(t, uint(101), chat.ChatID)
    assert.Equal(t, "user2", chat.Username)
    assert.Equal(t, "Public User 2", chat.PublicName)
    assert.Equal(t, "avatar.jpg", chat.Avatar)
    assert.Len(t, chat.Messages, 1)

    assert.Equal(t, uint(1), chat.Messages[0].MessageID)
    assert.Equal(t, "Hello", chat.Messages[0].Content)
    assert.Equal(t, "user2", chat.Messages[0].Sender)
    assert.Equal(t, "user1", chat.Messages[0].Recipient)
    assert.Equal(t, time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC), chat.Messages[0].Timestamp)
    assert.True(t, chat.Messages[0].IsRead)

    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetChat_ForbiddenAccess(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    repo := NewChatRepository(db)

    ctx := context.Background()
    id := uint64(101)
    username := "user3"

    mock.ExpectQuery(
        `WITH chat_messages AS \(.+\) SELECT c\.id AS chat_id, CASE WHEN c\.user1 = \$2 THEN c\.user1 ELSE c\.user2 END AS first_user_username, CASE WHEN c\.user1 = \$2 THEN c\.user2 ELSE c\.user1 END AS other_user_username, u\.public_name AS other_user_name, u\.avatar AS other_user_avatar, cm\.message_id, cm\.content AS message_content, cm\.sender AS message_sender, cm\.recipient, cm\.timestamp AS message_timestamp, cm\.is_read AS message_is_read FROM chat c JOIN flow_user u ON u\.username = CASE WHEN c\.user1 = \$2 THEN c\.user2 ELSE c\.user1 END LEFT JOIN chat_messages cm ON c\.id = cm\.chat_id WHERE c\.id = \$1;`,
    ).WithArgs(id, username).
        WillReturnRows(sqlmock.NewRows([]string{
            "chat_id", "first_user_username", "other_user_username", "other_user_name", "other_user_avatar",
            "message_id", "message_content", "message_sender", "message_recipient", "message_timestamp", "message_is_read",
        }).
            AddRow(101, "user1", "user2", "Public User 2", "avatar.jpg", 1, "Hello", "user2", "user1", time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC), true))

    _, err = repo.GetChat(ctx, id, username)
    assert.ErrorIs(t, err, domain.ErrForbidden)

    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetChat_NotFound(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    repo := NewChatRepository(db)

    ctx := context.Background()
    id := uint64(101)
    username := "user1"

    mock.ExpectQuery(
        `WITH chat_messages AS \(.+\) SELECT c\.id AS chat_id, CASE WHEN c\.user1 = \$2 THEN c\.user1 ELSE c\.user2 END AS first_user_username, CASE WHEN c\.user1 = \$2 THEN c\.user2 ELSE c\.user1 END AS other_user_username, u\.public_name AS other_user_name, u\.avatar AS other_user_avatar, cm\.message_id, cm\.content AS message_content, cm\.sender AS message_sender, cm\.recipient, cm\.timestamp AS message_timestamp, cm\.is_read AS message_is_read FROM chat c JOIN flow_user u ON u\.username = CASE WHEN c\.user1 = \$2 THEN c\.user2 ELSE c\.user1 END LEFT JOIN chat_messages cm ON c\.id = cm\.chat_id WHERE c\.id = \$1;`,
    ).WithArgs(id, username).
        WillReturnRows(sqlmock.NewRows([]string{
            "chat_id", "first_user_username", "other_user_username", "other_user_name", "other_user_avatar",
            "message_id", "message_content", "message_sender", "message_recipient", "message_timestamp", "message_is_read",
        }))

    _, err = repo.GetChat(ctx, id, username)
    assert.ErrorIs(t, err, domain.ErrNotFound)

    assert.NoError(t, mock.ExpectationsWereMet())
}
