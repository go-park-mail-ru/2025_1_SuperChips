package repository

import (
	"context"
	"database/sql"
	"strings"

	pin "github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type flowDBSchema struct {
	ID             uint64
	Title          sql.NullString
	Description    sql.NullString
	AuthorId       uint64
	AuthorUsername string
	CreatedAt      sql.NullTime
	UpdatedAt      sql.NullTime
	IsPrivate      bool
	IsNSFW         bool
	IsStar         bool
	LikeCount      int
	MediaURL       string
	Width          sql.NullInt64
	Height         sql.NullInt64
}

type pgPinStorage struct {
	db         *sql.DB
	imgDir     string
	baseURL    string
	imgStrgURL string
}

func NewPGPinStorage(db *sql.DB, imgDir, baseURL string) (*pgPinStorage, error) {
	storage := &pgPinStorage{
		db:         db,
		imgDir:     imgDir,
		baseURL:    baseURL,
		imgStrgURL: baseURL + strings.ReplaceAll(imgDir, ".", ""),
	}

	return storage, nil
}

func (p *pgPinStorage) assembleMediaURL(fileName string) string {
	return p.imgStrgURL + "/" + fileName
}

// Выдаёт пины на странице page, если на странице pageSize пинов.
// NSFW НЕ выдаёт.
// Лайки НЕ проставляет.
// Звёздные пины идут сначала.
func (p *pgPinStorage) GetPins(page int, pageSize int) ([]pin.PinData, error) {
	rows, err := p.db.Query(`
		SELECT 
			f.id, 
			f.title, 
			f.description, 
			f.author_id, 
			f.is_private, 
			f.media_url,
			f.width,
			f.height,
			f.is_nsfw,
			f.is_star,
			fu.username
		FROM flow f
		JOIN flow_user fu ON f.author_id = fu.id
		WHERE f.is_private = false AND f.is_nsfw = false
		ORDER BY f.is_star DESC, f.created_at DESC
		LIMIT $1
		OFFSET $2
	`, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var pins []pin.PinData

	for rows.Next() {
		var flowDBRow flowDBSchema
		err := rows.Scan(
			&flowDBRow.ID,
			&flowDBRow.Title,
			&flowDBRow.Description,
			&flowDBRow.AuthorId,
			&flowDBRow.IsPrivate,
			&flowDBRow.MediaURL,
			&flowDBRow.Width,
			&flowDBRow.Height,
			&flowDBRow.IsNSFW,
			&flowDBRow.IsStar,
			&flowDBRow.AuthorUsername,
		)
		if err != nil {
			return nil, err
		}

		pin := pin.PinData{
			FlowID:         flowDBRow.ID,
			Description:    flowDBRow.Description.String,
			Header:         flowDBRow.Title.String,
			AuthorID:       flowDBRow.AuthorId,
			MediaURL:       p.assembleMediaURL(flowDBRow.MediaURL),
			Width:          int(flowDBRow.Width.Int64),
			Height:         int(flowDBRow.Height.Int64),
			IsNSFW:         flowDBRow.IsNSFW,
			IsStar:         flowDBRow.IsStar,
			AuthorUsername: flowDBRow.AuthorUsername,
		}
		pins = append(pins, pin)
	}

	return pins, nil
}

// Получение всех звёздных пинов пользователя с заданным userID.
// NSFW выдаётся.
// Лайки НЕ проставляются.
func (p *pgPinStorage) GetStarPins(ctx context.Context, userID int) ([]pin.PinData, error) {
	rows, err := p.db.Query(`
		SELECT 
			f.id, 
			f.title, 
			f.description, 
			f.author_id, 
			f.is_private, 
			f.media_url,
			f.width,
			f.height,
			f.is_nsfw,
			f.is_star,
			fu.username
		FROM flow f
		JOIN flow_user fu ON f.author_id = fu.id
		WHERE author_id = $1 AND is_star = true
		ORDER BY f.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var pins []pin.PinData

	for rows.Next() {
		var flowDBRow flowDBSchema
		err := rows.Scan(
			&flowDBRow.ID,
			&flowDBRow.Title,
			&flowDBRow.Description,
			&flowDBRow.AuthorId,
			&flowDBRow.IsPrivate,
			&flowDBRow.MediaURL,
			&flowDBRow.Width,
			&flowDBRow.Height,
			&flowDBRow.IsNSFW,
			&flowDBRow.IsStar,
			&flowDBRow.AuthorUsername,
		)
		if err != nil {
			return nil, err
		}

		pin := pin.PinData{
			FlowID:         flowDBRow.ID,
			Description:    flowDBRow.Description.String,
			Header:         flowDBRow.Title.String,
			AuthorID:       flowDBRow.AuthorId,
			MediaURL:       p.assembleMediaURL(flowDBRow.MediaURL),
			Width:          int(flowDBRow.Width.Int64),
			Height:         int(flowDBRow.Height.Int64),
			IsNSFW:         flowDBRow.IsNSFW,
			IsStar:         flowDBRow.IsStar,
			AuthorUsername: flowDBRow.AuthorUsername,
		}
		pins = append(pins, pin)
	}

	return pins, nil
}

func (p *pgPinStorage) GetPinsCount(ctx context.Context, userID int) (int, error) {
	row := p.db.QueryRowContext(ctx, `
		SELECT COUNT(*) AS flow_count
		FROM flow
		WHERE author_id = $1
	`, userID)

	var flowCount int

	err := row.Scan(&flowCount)
	if err != nil {
		return 0, err
	}

	return flowCount, nil
}
