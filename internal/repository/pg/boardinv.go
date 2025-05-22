package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	boardinvService "github.com/go-park-mail-ru/2025_1_SuperChips/boardinv"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/google/uuid"
)

type pgBoardInvStorage struct {
	db *sql.DB
}

func NewBoardInvStorage(db *sql.DB) *pgBoardInvStorage {
	return &pgBoardInvStorage{db: db}
}

func (p *pgBoardInvStorage) IsBoardAuthor(ctx context.Context, boardID int, userID int) (bool, error) {
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

func (p *pgBoardInvStorage) GetUserIDsFromUsernames(ctx context.Context, inviteeNames []string) ([]boardinvService.NameToID, error) {
	var inviteeDatum []boardinvService.NameToID
	var values []any

	for _, name := range inviteeNames {
		values = append(values, name)
	}

	placeHoldersArr := make([]string, 0)
	for i := 1; i <= len(inviteeNames); i++ {
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

		inviteeData := boardinvService.NameToID{
			Username: username,
			ID:       nil,
		}
		if ID.Valid {
			tempID := int(ID.Int64)
			inviteeData.ID = &tempID
		}

		inviteeDatum = append(inviteeDatum, inviteeData)
	}

	return inviteeDatum, nil
}

func (p *pgBoardInvStorage) CreateInvitation(ctx context.Context, boardID int, userID int, invitation domain.Invitaion, inviteeIDs []int) (string, error) {
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

func (p *pgBoardInvStorage) DeleteInvitation(ctx context.Context, boardID int, link string) error {
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
		return boardinvService.ErrLinkNotFound
	}

	return nil
}

func (p *pgBoardInvStorage) GetInvitationLinks(ctx context.Context, boardID int) ([]domain.LinkParams, error) {
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

	links := []domain.LinkParams{}

	rowData := struct {
		Link       string
		IsPersonal bool
		Expiration sql.NullTime
		UsageLimit sql.NullInt64
		UsageCount sql.NullInt64
		Username   sql.NullString
	}{}

	for rows.Next() {
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
				newLink.UsageCount = &rowData.UsageCount.Int64
			}

			links = append(links, newLink)
			continue
		}

		// Иначе ссылка уже встречалась -> ссылка НЕперсональная -> добавляется новое имя.

		// Проверка согласованности данных.
		if rowData.IsPersonal {
			return nil, boardinvService.ErrInconsistentDataInDB
		}

		lastLink := links[len(links)-1]
		if rowData.Username.Valid {
			*lastLink.Names = append(*lastLink.Names, rowData.Username.String)
		}
	}

	return links, nil
}
