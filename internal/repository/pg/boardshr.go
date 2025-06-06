package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	boardshrService "github.com/go-park-mail-ru/2025_1_SuperChips/boardshr"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/google/uuid"
)

type pgBoardShrStorage struct {
	db *sql.DB
}

func NewBoardShrStorage(db *sql.DB) *pgBoardShrStorage {
	return &pgBoardShrStorage{db: db}
}

func (p *pgBoardShrStorage) IsBoardAuthor(ctx context.Context, boardID int, userID int) (bool, error) {
	var isAuthor bool

	row := p.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM board 
			WHERE id = $1 AND author_id = $2
		) AS is_author
	`, boardID, userID)

	err := row.Scan(&isAuthor)
	if err != nil {
		return false, err
	}

	return isAuthor, nil
}

func (p *pgBoardShrStorage) GetUserIDFromUsername(ctx context.Context, name string) (int, error) {
	rows := p.db.QueryRowContext(ctx, `
		SELECT id FROM flow_user WHERE username = $1
	`, name)

	var userID int
	err := rows.Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, boardshrService.ErrNonExistentUsername
	}
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (p *pgBoardShrStorage) GetUserIDsFromUsernames(ctx context.Context, names []string) ([]boardshrService.NameToID, error) {
	var inviteeDatum []boardshrService.NameToID
	var values []any

	for _, name := range names {
		values = append(values, name)
	}

	placeHoldersArr := make([]string, 0)
	for i := 1; i <= len(names); i++ {
		placeHoldersArr = append(placeHoldersArr, "$"+strconv.Itoa(i))
	}
	placeHolders := strings.Join(placeHoldersArr, ", ")

	sqlQuery := fmt.Sprintf(`
		SELECT a.username, b.id
		FROM (SELECT %s AS username) AS a
		LEFT JOIN flow_user AS b
			ON a.username = b.username
	`, placeHolders)

	rows, err := p.db.QueryContext(ctx, sqlQuery, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var username string
	var ID sql.NullInt64
	for rows.Next() {
		err := rows.Scan(&username, &ID)
		if err != nil {
			return nil, err
		}

		userData := boardshrService.NameToID{
			Username: username,
			ID:       nil,
		}
		if ID.Valid {
			tempID := int(ID.Int64)
			userData.ID = &tempID
		}

		inviteeDatum = append(inviteeDatum, userData)
	}

	return inviteeDatum, nil
}

func (p *pgBoardShrStorage) CreateInvitation(ctx context.Context, boardID int, userID int, invitation domain.Invitaion, inviteeIDs []int) (string, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	link := uuid.New().String()

	var invitationID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO board_invitation (board_id, link, is_personal, expiration, usage_limit, usage_count) 
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
	`, boardID, link, len(inviteeIDs) == 0, invitation.TimeLimit, invitation.UsageLimit, 0).Scan(&invitationID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.ErrConflict
	}
	if err != nil {
		return "", err
	}

	if len(inviteeIDs) == 0 {
		if err := tx.Commit(); err != nil {
			return "", err
		}
		return link, nil
	}

	// Следующий код выполняется только для персональных ссылок.

	values := []any{}
	paramCounter := 1
	placeHoldersArr := make([]string, 0)

	for _, inviteeID := range inviteeIDs {
		values = append(values, invitationID, inviteeID)
		placeHoldersArr = append(placeHoldersArr, fmt.Sprintf("($%d, $%d)", paramCounter, paramCounter+1))
		paramCounter += 2
	}

	sqlQuery := fmt.Sprintf(
		"INSERT INTO invitation_user (invitation_id, user_id) VALUES %s",
		strings.Join(placeHoldersArr, ", "))

	result, err := tx.ExecContext(ctx, sqlQuery, values...)
	if err != nil {
		return "", err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return "", err
	}
	if rowsAffected != int64(len(inviteeIDs)) {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return link, nil
}

func (p *pgBoardShrStorage) DeleteInvitation(ctx context.Context, boardID int, link string) error {
	result, err := p.db.ExecContext(ctx, `
		DELETE FROM board_invitation
		WHERE board_id = $1 AND link = $2
	`, boardID, link)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return boardshrService.ErrLinkNotFound
	}

	return nil
}

func (p *pgBoardShrStorage) GetInvitationLinks(ctx context.Context, boardID int) ([]domain.LinkParams, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT 
			bi.link AS link,
			bi.is_personal,
			bi.expiration,
			bi.usage_limit,
			bi.usage_count,
			fu.username
		FROM board_invitation AS bi
		LEFT JOIN invitation_user AS iu
			ON bi.id = iu.invitation_id
		LEFT JOIN flow_user AS fu
			ON iu.user_id = fu.id
		WHERE board_id = $1
		ORDER BY bi.id
	`, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := []domain.LinkParams{}

	for rows.Next() {
		rowData := struct {
			link       string
			isPersonal bool
			expiration sql.NullTime
			usageLimit sql.NullInt64
			usageCount int64
			username   sql.NullString
		}{}

		err := rows.Scan(
			&rowData.link,
			&rowData.isPersonal,
			&rowData.expiration,
			&rowData.usageLimit,
			&rowData.usageCount,
			&rowData.username)
		if err != nil {
			return nil, err
		}

		// Ссылка встречается впервые.
		if len(links) == 0 || links[len(links)-1].Link != rowData.link {
			newLink := domain.LinkParams{
				Link: rowData.link,
			}

			if !rowData.isPersonal && rowData.username.Valid {
				newLink.Names = &[]string{rowData.username.String}
			}

			if rowData.expiration.Valid {
				newLink.TimeLimit = &rowData.expiration.Time
			}

			if rowData.usageLimit.Valid {
				newLink.UsageLimit = &rowData.usageLimit.Int64
			}
			newLink.UsageCount = rowData.usageCount

			links = append(links, newLink)
			continue
		}

		// Иначе ссылка уже встречалась -> ссылка НЕперсональная -> добавляется новое имя.

		// Проверка согласованности данных.
		if rowData.isPersonal {
			return nil, boardshrService.ErrInconsistentDataInDB
		}

		lastLink := links[len(links)-1]
		if rowData.username.Valid {
			*lastLink.Names = append(*lastLink.Names, rowData.username.String)
		}
	}

	return links, nil
}

func (p *pgBoardShrStorage) IsBoardEditor(ctx context.Context, boardID int, userID int) (bool, error) {
	row := p.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM board
				WHERE id = $1 AND author_id = $2
			UNION
			SELECT 1 FROM board_coauthor
				WHERE board_id = $1 AND coauthor_id = $2
		) AS is_editor
	`, boardID, userID)

	var isEditor bool
	err := row.Scan(&isEditor)
	if err != nil {
		return false, err
	}

	return isEditor, nil
}

func (p *pgBoardShrStorage) GetUsernameFromUserID(ctx context.Context, userID int) (string, error) {
	row := p.db.QueryRowContext(ctx, `
		SELECT username FROM flow_user
		WHERE id = $1
		LIMIT 1
	`, userID)

	var name string
	err := row.Scan(&name)
	if err != nil {
		return "", err
	}

	return name, nil
}

func (p *pgBoardShrStorage) GetLinkParams(ctx context.Context, link string) (int, domain.LinkParams, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT 
			bi.board_id AS board_id,
			bi.link AS link,
			bi.is_personal,
			bi.expiration,
			bi.usage_limit,
			bi.usage_count,
			fu.username
		FROM board_invitation AS bi
		LEFT JOIN invitation_user AS iu
			ON bi.id = iu.invitation_id
		LEFT JOIN flow_user AS fu
			ON iu.user_id = fu.id
		WHERE link = $1
	`, link)
	if err != nil {
		return 0, domain.LinkParams{}, err
	}
	defer rows.Close()

	boardID := 0
	linkParams := &domain.LinkParams{}

	isFirstRow := true

	for rows.Next() {
		rowData := struct {
			link       string
			isPersonal bool
			expiration sql.NullTime
			usageLimit sql.NullInt64
			usageCount int64
			username   sql.NullString
		}{}

		err := rows.Scan(
			&boardID,
			&rowData.link,
			&rowData.isPersonal,
			&rowData.expiration,
			&rowData.usageLimit,
			&rowData.usageCount,
			&rowData.username)
		if err != nil {
			return 0, domain.LinkParams{}, err
		}

		if isFirstRow {
			linkParams.Link = rowData.link

			if !rowData.isPersonal && rowData.username.Valid {
				linkParams.Names = &[]string{rowData.username.String}
			}

			if rowData.expiration.Valid {
				linkParams.TimeLimit = &rowData.expiration.Time
			}

			if rowData.usageLimit.Valid {
				linkParams.UsageLimit = &rowData.usageLimit.Int64
			}
			linkParams.UsageCount = rowData.usageCount

			isFirstRow = false
			continue
		}

		// Публиная ссылка должна встречаться единожды, иначе данные несогласованны.
		if rowData.isPersonal {
			return 0, domain.LinkParams{}, boardshrService.ErrInconsistentDataInDB
		}

		if rowData.username.Valid {
			*linkParams.Names = append(*linkParams.Names, rowData.username.String)
		}
	}

	// Проверка, что ссылка содержалась в БД.
	if isFirstRow {
		return 0, domain.LinkParams{}, boardshrService.ErrLinkNotFound
	}

	return boardID, *linkParams, nil
}

func (p *pgBoardShrStorage) AddBoardCoauthorByLink(ctx context.Context, boardID int, userID int, link string) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO board_coauthor (board_id, coauthor_id) VALUES ($1, $2)
	`, boardID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return boardshrService.ErrFailCoauthorInsert
	}

	result, err = tx.ExecContext(ctx, `
		UPDATE board_invitation SET usage_count = usage_count + 1
		WHERE link = $1
	`, link)
	if err != nil {
		return err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return boardshrService.ErrFailCoauthorInsert
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (p *pgBoardShrStorage) DeleteCoauthor(ctx context.Context, boardID int, userID int) error {
    tx, err := p.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    _, err = tx.ExecContext(ctx, `
        DELETE FROM board_post
        WHERE board_id = $1
        AND flow_id IN (
            SELECT f.id FROM flow f
            WHERE f.author_id = $2
            AND f.is_private = true
        )
    `, boardID, userID)
    if err != nil {
        return fmt.Errorf("failed to delete private pins: %w", err)
    }

    result, err := tx.ExecContext(ctx, `
        DELETE FROM board_coauthor
        WHERE board_id = $1 AND coauthor_id = $2
    `, boardID, userID)
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rowsAffected == 0 {
        return boardshrService.ErrFailCoauthorDelete
    }

    _, err = tx.ExecContext(ctx, `
        UPDATE board
        SET flow_count = (
            SELECT COUNT(*) FROM board_post
            WHERE board_id = $1
        )
        WHERE id = $1
    `, boardID)
    if err != nil {
        return fmt.Errorf("failed to update flow count: %w", err)
    }

    if err := tx.Commit(); err != nil {
        return err
    }

    return nil
}

func (p *pgBoardShrStorage) GetCoauthors(ctx context.Context, boardID int) ([]domain.Contact, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT 
			fu.username, 
			fu.public_name, 
			fu.avatar, 
			fu.is_external_avatar
		FROM board_coauthor AS bc
		LEFT JOIN flow_user AS fu
			ON bc.coauthor_id = fu.id
		WHERE bc.board_id = $1
	`, boardID)
	if err != nil {
		return nil, err
	}

	var coauthors []domain.Contact
	var isExternalAvatar sql.NullBool

	for rows.Next() {
		var coauthor domain.Contact
		if err := rows.Scan(
			&coauthor.Username,
			&coauthor.PublicName,
			&coauthor.Avatar,
			&isExternalAvatar,
		); err != nil {
			return nil, err
		}

		coauthor.IsExternalAvatar = isExternalAvatar.Bool

		coauthors = append(coauthors, coauthor)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return coauthors, nil
}

func (p *pgBoardShrStorage) GetAuthor(ctx context.Context, boardID int) (domain.Contact, error) {
	rows := p.db.QueryRowContext(ctx, `
		SELECT 
			fu.username, 
			fu.public_name, 
			fu.avatar, 
			fu.is_external_avatar
		FROM board AS b
		LEFT JOIN flow_user AS fu
			ON b.author_id = fu.id
		WHERE b.id = $1
	`, boardID)

	var isExternalAvatar sql.NullBool
	var author domain.Contact

	err := rows.Scan(
		&author.Username,
		&author.PublicName,
		&author.Avatar,
		&isExternalAvatar,
	)
	if err != nil {
		return domain.Contact{}, err
	}

	author.IsExternalAvatar = isExternalAvatar.Bool

	return author, nil
}

func (p *pgBoardShrStorage) GetCoauthorsIDs(ctx context.Context, boardID int) ([]int, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT coauthor_id
		FROM board_coauthor
		WHERE board_id = $1
	`, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ID int
	IDs := []int{}

	for rows.Next() {
		err := rows.Scan(&ID)
		if err != nil {
			return nil, err
		}

		IDs = append(IDs, ID)
	}

	return IDs, nil
}
