package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type ChatRepository struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{
		db: db,
	}
}

func (repo *ChatRepository) GetNewMessages(ctx context.Context, username string, offset time.Time) ([]domain.Message, error) {
	rows, err := repo.db.QueryContext(ctx, `
	SELECT id, content, timestamp, is_read, sender, recipient, chat_id
	FROM message
	WHERE timestamp > $1
	AND recipient = $2
	AND sent = false;
	`, offset, username)
	if err != nil {
		return nil, err
	}

	var messages []domain.Message

	for rows.Next() {
		var message domain.Message
		if err := rows.Scan(
			&message.MessageID,
			&message.Content,
			&message.Timestamp,
			&message.IsRead,
			&message.Sender,
			&message.Recipient,
			&message.ChatID,
		); err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (repo *ChatRepository) AddMessage(ctx context.Context, message domain.Message) error {
	// the EXISTS statement
	// ensures that
	// message can only be added to a chat
	// where both participants are the
	// ones mentioned in the message struct
	// for safety purposes
	_, err := repo.db.ExecContext(ctx, `
	INSERT INTO message (content, sender, recipient, chat_id, sent)
	SELECT $1, $2, $3, $4, $5
	WHERE EXISTS (
		SELECT 1 FROM chat 
		WHERE id = $4 AND 
		(($2 = user1 AND $3 = user2) OR ($2 = user2 AND $3 = user1))
	);
	`, message.Content, message.Sender, message.Recipient, message.ChatID, message.Sent)
	if err != nil {
		return err
	}

	return nil
}

func (repo *ChatRepository) MarkRead(ctx context.Context, messageID, chatID int) error {
	_, err := repo.db.ExecContext(ctx, `
	UPDATE message
	SET is_read = true
	WHERE chat_id = $1 AND id <= $2
	`, chatID, messageID)
	if err != nil {
		return err
	}

	return nil
}

func (repo *ChatRepository) GetChats(ctx context.Context, username string) ([]domain.Chat, error) {
	query := `
    WITH unread_counts AS (
        SELECT 
            chat_id,
            COUNT(*) FILTER (WHERE is_read = FALSE AND recipient = $1) AS unread_count
        FROM message
        GROUP BY chat_id
    ),
    last_message AS (
        SELECT DISTINCT ON (m.chat_id)
            m.chat_id,
            m.id AS message_id,
            m.content,
            m.sender,
			m.recipient,
            m.timestamp,
            m.is_read
        FROM message m
        ORDER BY m.chat_id, m.timestamp DESC
    )
    SELECT 
        c.id AS chat_id,
        CASE 
            WHEN c.user1 = $1 THEN c.user2 
            ELSE c.user1 
        END AS other_user_username,
        u.public_name AS other_user_name,
        u.avatar AS other_user_avatar,
		u.is_external_avatar,
        lm.message_id,
        lm.content AS message_content,
        lm.sender AS message_sender,
		lm.recipient,
        lm.timestamp AS message_timestamp,
        lm.is_read AS message_is_read,
        uc.unread_count
    FROM chat c
    JOIN flow_user u ON u.username = CASE 
                                      WHEN c.user1 = $1 THEN c.user2 
                                      ELSE c.user1 
                                   END
    LEFT JOIN last_message lm ON c.id = lm.chat_id
    LEFT JOIN unread_counts uc ON c.id = uc.chat_id
    WHERE c.user1 = $1 OR c.user2 = $1
    ORDER BY lm.timestamp DESC NULLS LAST;
    `

	rows, err := repo.db.QueryContext(ctx, query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chatsMap := make(map[int]*domain.Chat)

	for rows.Next() {
		var (
			chatID              int
			otherUserUsername   string
			otherUserPublicName string
			otherUserAvatar     string
			isExternalAvatar    sql.NullBool
			messageID           sql.NullInt64
			messageContent      sql.NullString
			messageSender       sql.NullString
			messageRecipient    sql.NullString
			messageTimestamp    sql.NullTime
			messageIsRead       sql.NullBool
			unreadCount         sql.NullInt64
		)

		err := rows.Scan(
			&chatID,
			&otherUserUsername,
			&otherUserPublicName,
			&otherUserAvatar,
			&isExternalAvatar,
			&messageID,
			&messageContent,
			&messageSender,
			&messageRecipient,
			&messageTimestamp,
			&messageIsRead,
			&unreadCount,
		)
		if err != nil {
			return nil, err
		}

		if _, exists := chatsMap[chatID]; !exists {
			chatsMap[chatID] = &domain.Chat{
				ChatID:           uint(chatID),
				Username:         otherUserUsername,
				PublicName:       otherUserPublicName,
				Avatar:           otherUserAvatar,
				IsExternalAvatar: isExternalAvatar.Bool,
				MessageCount:     uint(unreadCount.Int64),
				Messages: []domain.Message{},
			}
		}

		if messageID.Valid {
			chatsMap[chatID].Messages = append(chatsMap[chatID].Messages, domain.Message{
				MessageID: uint(messageID.Int64),
				Content:   messageContent.String,
				Sender:    messageSender.String,
				Recipient: messageRecipient.String,
				Timestamp: messageTimestamp.Time,
				IsRead:    messageIsRead.Bool,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	chats := make([]domain.Chat, 0, len(chatsMap))
	for _, chat := range chatsMap {
		chats = append(chats, *chat)
	}

	return chats, nil
}

func (repo *ChatRepository) CreateChat(ctx context.Context, username, targetUsername string) (domain.Chat, error) {
    var chat domain.Chat
    var isExternalAvatar sql.NullBool

	err := repo.db.QueryRowContext(ctx, `
    WITH normalized_users AS (
        SELECT 
            LEAST($1, $2) AS user1,
            GREATEST($1, $2) AS user2
    ),
    inserted_chat AS (
        INSERT INTO chat (user1, user2)
        SELECT user1, user2 FROM normalized_users
        ON CONFLICT (user1, user2) DO NOTHING
        RETURNING id
    ),
    existing_chat AS (
        SELECT id
        FROM chat
        WHERE (user1, user2) = (SELECT user1, user2 FROM normalized_users)
    )
    SELECT 
        COALESCE(ic.id, ec.id) AS chat_id,
        u.avatar,
        u.public_name,
        u.is_external_avatar
    FROM inserted_chat ic
    FULL JOIN existing_chat ec ON TRUE
    JOIN flow_user u ON u.username = $1;
	`, targetUsername, username).
	Scan(&chat.ChatID, &chat.Avatar, &chat.PublicName, &isExternalAvatar)
    if err != nil {
        return domain.Chat{}, err
    }

    chat.Username = targetUsername
    chat.IsExternalAvatar = isExternalAvatar.Bool

    return chat, nil
}

func (repo *ChatRepository) GetContacts(ctx context.Context, username string) ([]domain.Contact, error) {
	var isExternalAvatar sql.NullBool

	rows, err := repo.db.QueryContext(ctx, `
	SELECT u.username, u.public_name, u.avatar, u.is_external_avatar
	FROM contact c
	INNER JOIN flow_user u ON u.username = c.contact_username
	WHERE c.user_username = $1
	`, username)
	if err != nil {
		return nil, err
	}

	var contacts []domain.Contact

	for rows.Next() {
		var contact domain.Contact
		if err := rows.Scan(
			&contact.Username,
			&contact.PublicName,
			&contact.Avatar,
			&isExternalAvatar,
		); err != nil {
			return nil, err
		}

		contact.IsExternalAvatar = isExternalAvatar.Bool

		contacts = append(contacts, contact)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return contacts, nil
}

func (repo *ChatRepository) CreateContact(ctx context.Context, username, targetUsername string) (domain.Chat, error) {
	
	err := repo.AddToContacts(ctx, username, targetUsername)
	if err != nil {
		return domain.Chat{}, err
	}

	return repo.CreateChat(ctx, targetUsername, username)
}

func (repo *ChatRepository) AddToContacts(ctx context.Context, username, targetUsername string) error {
	var id sql.NullInt64
	
	err := repo.db.QueryRowContext(ctx, `
	INSERT INTO contact
	(user_username, contact_username)
	VALUES
	($1, $2)
	ON CONFLICT (user_username, contact_username) DO NOTHING
	RETURNING id
	`, username, targetUsername).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrConflict
	}
	if err != nil {
		return err
	}

	return nil
}

func (repo *ChatRepository) GetChat(ctx context.Context, id uint64, username string) (domain.Chat, error) {
	query := `
	WITH chat_messages AS (
		SELECT
			m.id AS message_id,
			m.chat_id,
			m.content,
			m.sender,
			m.recipient,
			m.timestamp,
			m.is_read
		FROM message m
		WHERE m.chat_id = $1
		ORDER BY m.timestamp DESC
	)
	SELECT
		c.id AS chat_id,
		CASE
			WHEN c.user1 = $2 THEN c.user1
			ELSE c.user2
		END AS first_user_username,
		CASE
			WHEN c.user1 = $2 THEN c.user2
			ELSE c.user1
		END AS other_user_username,
		u.public_name AS other_user_name,
		u.avatar AS other_user_avatar,
		cm.message_id,
		cm.content AS message_content,
		cm.sender AS message_sender,
		cm.recipient,
		cm.timestamp AS message_timestamp,
		cm.is_read AS message_is_read
	FROM chat c
	JOIN flow_user u ON u.username = CASE
		WHEN c.user1 = $2 THEN c.user2
		ELSE c.user1
	END
	LEFT JOIN chat_messages cm ON c.id = cm.chat_id
	WHERE c.id = $1;
    `

	rows, err := repo.db.QueryContext(ctx, query, id, username)
	if err != nil {
		return domain.Chat{}, err
	}
	defer rows.Close()

	var chat *domain.Chat

	for rows.Next() {
		var (
			firstUserUsername   string
			otherUserUsername   string
			otherUserPublicName string
			otherUserAvatar     string
			messageID           sql.NullInt64
			messageContent      sql.NullString
			messageSender       sql.NullString
			messageRecipient    sql.NullString
			messageTimestamp    sql.NullTime
			messageIsRead       sql.NullBool
		)

		err := rows.Scan(
			&id,
			&firstUserUsername,
			&otherUserUsername,
			&otherUserPublicName,
			&otherUserAvatar,
			&messageID,
			&messageContent,
			&messageSender,
			&messageRecipient,
			&messageTimestamp,
			&messageIsRead,
		)
		if err != nil {
			return domain.Chat{}, err
		}

		if chat == nil {
			log.Printf("username: %s, first: %s, second: %s", username, firstUserUsername, otherUserUsername)
			if firstUserUsername != username && otherUserUsername != username {
				return domain.Chat{}, domain.ErrForbidden
			}

			chat = &domain.Chat{
				ChatID:     uint(id),
				Username:   otherUserUsername,
				PublicName: otherUserPublicName,
				Avatar:     otherUserAvatar,
				Messages:   []domain.Message{},
			}
		}

		if messageID.Valid {
			chat.Messages = append(chat.Messages, domain.Message{
				MessageID: uint(messageID.Int64),
				Content:   messageContent.String,
				Sender:    messageSender.String,
				Timestamp: messageTimestamp.Time,
				IsRead:    messageIsRead.Bool,
				Recipient: messageRecipient.String,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return domain.Chat{}, err
	}

	if chat == nil {
		return domain.Chat{}, domain.ErrNotFound
	}

	return *chat, nil
}

