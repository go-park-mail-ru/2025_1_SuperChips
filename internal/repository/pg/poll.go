package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type pollDBSchema struct {
	ID        uint64
	Name      sql.NullString
	CreatedAt sql.NullTime
	Delay     int
	Screen    sql.NullString
}

type questionDBSchema struct {
	ID        uint64
	PollID    uint64
	OrderNum  int64
	Content   sql.NullString
	Type      sql.NullString
	AuthorID  uint64
	CreatedAt sql.NullTime
}

type pgPollStorage struct {
	db *sql.DB
}

func NewPGPollStorage(db *sql.DB) *pgPollStorage {
	storage := &pgPollStorage{
		db: db,
	}

	return storage
}

func (p *pgPollStorage) GetAllPolls(ctx context.Context) ([]domain.Poll, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT 
			p.id,
			p.name,
			p.screen,
			p.delay,
			q.id,
			q.order_num,
			q.content,
			q.type
		FROM poll p
		JOIN question q
			ON p.id = q.poll_id
		ORDER BY q.id ASC, q.order_num ASC;
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var polls []domain.Poll

	isFirst := true
	for rows.Next() {
		var pollDBRow pollDBSchema
		var questionDBRow questionDBSchema
		err = rows.Scan(
			&pollDBRow.ID,
			&pollDBRow.Name,
			&pollDBRow.Screen,
			&pollDBRow.Delay,
			&questionDBRow.ID,
			&questionDBRow.OrderNum,
			&questionDBRow.Content,
			&questionDBRow.Type)
		if err != nil {
			return nil, err
		}

		question := domain.Question{
			ID:    questionDBRow.ID,
			Text:  questionDBRow.Content.String,
			Order: questionDBRow.OrderNum,
			Type:  questionDBRow.Type.String,
		}

		if isFirst {
			polls = append(polls, domain.Poll{
				ID:        pollDBRow.ID,
				Header:    pollDBRow.Name.String,
				Questions: []domain.Question{question},
				Delay:     pollDBRow.Delay,
				Screen:    strings.Split(pollDBRow.Screen.String, ","),
			})
			isFirst = false
			continue
		}

		lastPoll := &polls[len(polls)-1]
		if lastPoll.ID != pollDBRow.ID {
			polls = append(polls, domain.Poll{
				ID:        pollDBRow.ID,
				Header:    pollDBRow.Name.String,
				Questions: []domain.Question{question},
				Delay:     pollDBRow.Delay,
				Screen:    strings.Split(pollDBRow.Screen.String, ","),
			})
		} else {
			lastPoll.Questions = append(lastPoll.Questions, question)
		}
	}

	return polls, nil
}

func (p *pgPollStorage) AddAnswer(ctx context.Context, pollID uint64, userID uint64, answers []domain.Answer) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, answer := range answers {
		row := tx.QueryRowContext(ctx, `
			INSERT INTO answer (poll_id, question_id, content, type, author_id)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (author_id, question_id)
				DO UPDATE 
					SET content = EXCLUDED.content, type = EXCLUDED.type
			RETURNING id;
		`, pollID, answer.QuestionID, answer.Content, answer.Type, userID)

		var answerID uint64
		err := row.Scan(&answerID)
		if err != nil {
			return errors.New("") // ErrUntracked
		}
	}

	return tx.Commit()
}

func (p *pgPollStorage) GetAllStarStat(ctx context.Context) ([]domain.QuestionStarAvg, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT 
			a.poll_id,
			p.name, 
			a.question_id, 
			q.content,
			AVG(a.content::int) as average_rating
		FROM answer a
		JOIN question q
			ON a.question_id = q.id
		JOIN poll p
			ON q.poll_id = p.id
		WHERE a.type = 'stars'
		GROUP BY 
			a.poll_id,
			p.name, 
			a.question_id, 
			q.content
		ORDER BY a.poll_id ASC, a.question_id ASC;
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var stats []domain.QuestionStarAvg

	type DBSchema struct {
		PollID     int
		Name       string
		QuestionID int
		Content    string
		Avg        float64
	}

	for rows.Next() {
		var DBRow DBSchema
		err = rows.Scan(
			&DBRow.PollID,
			&DBRow.Name,
			&DBRow.QuestionID,
			&DBRow.Content,
			&DBRow.Avg)
		if err != nil {
			return nil, err
		}

		stat := domain.QuestionStarAvg{
			PollID:       DBRow.PollID,
			PollHeader:   DBRow.Name,
			QuestionID:   DBRow.QuestionID,
			QuestionText: DBRow.Content,
			Average:      DBRow.Avg,
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (p *pgPollStorage) GetAllAnswers(ctx context.Context) ([]domain.QuestionAnswer, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT 
			a.poll_id,
			p.name, 
			a.question_id, 
			q.content as question,
			a.content as answer
		FROM answer a
		JOIN question q
			ON a.question_id = q.id
		JOIN poll p
			ON q.poll_id = p.id
		WHERE a.type = 'text'
		ORDER BY a.poll_id ASC, a.question_id ASC;
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var answers []domain.QuestionAnswer

	type DBSchema struct {
		PollID          int
		Name            string
		QuestionID      int
		QuestionContent string
		AnswerContent   string
	}

	for rows.Next() {
		var DBRow DBSchema
		err = rows.Scan(
			&DBRow.PollID,
			&DBRow.Name,
			&DBRow.QuestionID,
			&DBRow.QuestionContent,
			&DBRow.AnswerContent)
		if err != nil {
			return nil, err
		}

		answer := domain.QuestionAnswer{
			PollID:       DBRow.PollID,
			PollHeader:   DBRow.Name,
			QuestionID:   DBRow.QuestionID,
			QuestionText: DBRow.QuestionContent,
			Content:      DBRow.AnswerContent,
		}
		answers = append(answers, answer)
	}

	return answers, nil
}
