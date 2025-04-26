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

func (p *pgPollStorage) getPoll(ctx context.Context, pollID uint64) (domain.Poll, error) {
	row := p.db.QueryRowContext(ctx, `
        SELECT
			id,
			name,
			delay,
			screen
		FROM poll
		WHERE id = $1
		LIMIT 1;
    `, pollID)

	var pollDBRow pollDBSchema
	err := row.Scan(
		&pollDBRow.ID,
		&pollDBRow.Name,
		&pollDBRow.Delay,
		&pollDBRow.Screen)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Poll{}, errors.New("") // ErrPinNotFound
	}
	if err != nil {
		return domain.Poll{}, errors.New("") // ErrUntracked
	}

	questions, err := p.getQuestions(ctx, pollID)
	if err != nil {
		return domain.Poll{}, err
	}

	poll := domain.Poll{
		ID:        pollDBRow.ID,
		Header:    pollDBRow.Name.String,
		Questions: questions,
		Delay:     pollDBRow.Delay,
		Screen:    strings.Split(pollDBRow.Screen.String, ","),
	}

	return poll, nil
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

func (p *pgPollStorage) getQuestions(ctx context.Context, pollID uint64) ([]domain.Question, error) {
	rows, err := p.db.QueryContext(ctx, `
        SELECT
			id,
			order_num,
			content,
			type
		FROM question
		WHERE poll_id = $1
		ORDER BY order_num ASC;
    `, pollID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var questions []domain.Question

	for rows.Next() {
		var questionDBRow questionDBSchema
		err = rows.Scan(
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
		questions = append(questions, question)
	}

	return questions, nil
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
