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

	var fields []string
	var values []any
	paramCounter := 1

	fields = append(fields, "board_id")
	values = append(values, boardID)
	paramCounter++

	fields = append(fields, "link")
	values = append(values, link)
	paramCounter++

	if len(inviteeIDs) == 0 {
		fields = append(fields, "is_personal")
		values = append(values, true)
		paramCounter++
	}

	if invitation.TimeLimit != nil {
		fields = append(fields, "expiration")
		values = append(values, invitation.TimeLimit)
		paramCounter++
	}

	if invitation.UsageLimit != nil {
		fields = append(fields, "usage_limit")
		values = append(values, *invitation.UsageLimit)
		paramCounter++

		fields = append(fields, "usage_count")
		values = append(values, 0)
		paramCounter++
	}

	placeHoldersArr := make([]string, 0)
	for i := 1; i < paramCounter; i++ {
		placeHoldersArr = append(placeHoldersArr, "$"+strconv.Itoa(i))
	}
	placeHolders := strings.Join(placeHoldersArr, ", ")

	sqlQuery := fmt.Sprintf(
		"INSERT INTO board_invitation (%s) VALUES (%s) RETURNING id",
		strings.Join(fields, ", "),
		placeHolders)

	var invitationID int
	err = tx.QueryRowContext(ctx, sqlQuery, values...).Scan(&invitationID)
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

	values = []any{}
	paramCounter = 1
	placeHoldersArr = make([]string, 0)

	for _, inviteeID := range inviteeIDs {
		values = append(values, invitationID, inviteeID)
		placeHoldersArr = append(placeHoldersArr, fmt.Sprintf("($%d, $%d)", paramCounter, paramCounter+1))
		paramCounter += 2
	}

	sqlQuery = fmt.Sprintf(
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
			Link       string
			IsPersonal bool
			Expiration sql.NullTime
			UsageLimit sql.NullInt64
			UsageCount int64
			Username   sql.NullString
		}{}

		err := rows.Scan(
			&rowData.Link,
			&rowData.IsPersonal,
			&rowData.Expiration,
			&rowData.UsageLimit,
			&rowData.UsageCount,
			&rowData.Username)
		if err != nil {
			return nil, err
		}

		// Ссылка встречается впервые.
		if len(links) == 0 || links[len(links)-1].Link != rowData.Link {
			newLink := domain.LinkParams{
				Link: rowData.Link,
			}

			if !rowData.IsPersonal && rowData.Username.Valid {
				newLink.Names = &[]string{rowData.Username.String}
			}

			if rowData.Expiration.Valid {
				newLink.TimeLimit = &rowData.Expiration.Time
			}

			if rowData.UsageLimit.Valid {
				newLink.UsageLimit = &rowData.UsageLimit.Int64
			}
			newLink.UsageCount = rowData.UsageCount

			links = append(links, newLink)
			continue
		}

		// Иначе ссылка уже встречалась -> ссылка НЕперсональная -> добавляется новое имя.

		// Проверка согласованности данных.
		if rowData.IsPersonal {
			return nil, boardshrService.ErrInconsistentDataInDB
		}

		lastLink := links[len(links)-1]
		if rowData.Username.Valid {
			*lastLink.Names = append(*lastLink.Names, rowData.Username.String)
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
			Link       string
			IsPersonal bool
			Expiration sql.NullTime
			UsageLimit sql.NullInt64
			UsageCount int64
			Username   sql.NullString
		}{}

		err := rows.Scan(
			&boardID,
			&rowData.Link,
			&rowData.IsPersonal,
			&rowData.Expiration,
			&rowData.UsageLimit,
			&rowData.UsageCount,
			&rowData.Username)
		if err != nil {
			return 0, domain.LinkParams{}, err
		}

		if isFirstRow {
			linkParams.Link = rowData.Link

			if !rowData.IsPersonal && rowData.Username.Valid {
				linkParams.Names = &[]string{rowData.Username.String}
			}

			if rowData.Expiration.Valid {
				linkParams.TimeLimit = &rowData.Expiration.Time
			}

			if rowData.UsageLimit.Valid {
				linkParams.UsageLimit = &rowData.UsageLimit.Int64
			}
			linkParams.UsageCount = rowData.UsageCount

			isFirstRow = false
			continue
		}

		// Публиная ссылка должна встречаться единожды, иначе данные несогласованны.
		if rowData.IsPersonal {
			return 0, domain.LinkParams{}, boardshrService.ErrInconsistentDataInDB
		}

		if rowData.Username.Valid {
			*linkParams.Names = append(*linkParams.Names, rowData.Username.String)
		}
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
	result, err := p.db.ExecContext(ctx, `
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

	return nil
}

func (p *pgBoardShrStorage) GetCoauthors(ctx context.Context, boardID int) ([]string, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT
			fu.username
		FROM board_coauthor bc
		JOIN flow_user fu
			ON bc.coauthor_id = fu.id
		WHERE bc.board_id = $1
	`, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	name := ""
	names := []string{}

	for rows.Next() {
		err := rows.Scan(&name)
		if err != nil {
			return nil, err
		}

		names = append(names, name)
	}

	return names, nil
}
