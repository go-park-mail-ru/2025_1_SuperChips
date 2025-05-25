package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{
		db: db,
	}
}

func (r *CommentRepository) GetComments(ctx context.Context, flowID, userID, page, size int) ([]domain.Comment, error) {
	var isExternalAvatar sql.NullBool
	offset := (page - 1) * size

	rows, err := r.db.QueryContext(ctx, `
    SELECT 
        c.id, 
        c.author_id, 
        c.flow_id, 
        c.contents, 
        c.like_count, 
        c.created_at, 
        fu.username, 
        fu.avatar, 
        fu.is_external_avatar,
        EXISTS (
            SELECT 1 FROM comment_like cl 
            WHERE cl.comment_id = c.id AND cl.user_id = $2
        ) AS is_liked
    FROM comment c
    JOIN flow_user fu ON fu.id = c.author_id
    LEFT JOIN flow f ON f.id = c.flow_id
    WHERE c.flow_id = $1
    AND (f.is_private = false OR f.author_id = $2)
    ORDER BY c.created_at DESC
    OFFSET $3
    LIMIT $4
	`, flowID, userID, offset, size)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var comments []domain.Comment

	for rows.Next() {
		var comment domain.Comment
		if err := rows.Scan(
			&comment.ID,
			&comment.AuthorID,
			&comment.FlowID,
			&comment.Content,
			&comment.LikeCount,
			&comment.Timestamp,
			&comment.AuthorUsername,
			&comment.AuthorAvatar,
			&isExternalAvatar,
			&comment.IsLiked,
		); err != nil {
			return nil, err
		}
		comment.AuthorIsExternalAvatar = isExternalAvatar.Bool

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return comments, nil
}

func (r *CommentRepository) LikeComment(ctx context.Context, flowID, commentID, userID int) (string, error) {
    var action string

    err := r.db.QueryRowContext(ctx, `
	deleted AS (
		DELETE FROM comment_like
		WHERE user_id = $1 AND comment_id = $3
		RETURNING 'delete' AS action
	),
	inserted AS (
		INSERT INTO comment_like (user_id, comment_id)
		SELECT $1, $3
		WHERE NOT EXISTS (SELECT 1 FROM deleted)
		RETURNING 'insert' AS action
	),
	update_like_count AS (
		UPDATE comment
		SET like_count = like_count + CASE
			WHEN EXISTS (SELECT 1 FROM inserted) THEN 1
			WHEN EXISTS (SELECT 1 FROM deleted) THEN -1
			ELSE 0
		END
		WHERE id = $3
	)
	SELECT COALESCE((SELECT action FROM inserted), (SELECT action FROM deleted)) AS action
    `, userID, flowID, commentID).Scan(&action)
    if errors.Is(err, sql.ErrNoRows) {
        return "", domain.ErrForbidden
    }
    if err != nil {
        return "", err
    }

    return action, nil
}

func (r *CommentRepository) AddComment(ctx context.Context, flowID, userID int, content string) error {
	var id int
	
	err := r.db.QueryRowContext(ctx, `
	INSERT INTO comment (author_id, flow_id, contents)
	SELECT $1, $2, $3
	RETURNING id;
	`, userID, flowID, content).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrForbidden
	}
	if err != nil {
		return err
	}

	return nil
}

func (r *CommentRepository) DeleteComment(ctx context.Context, commentID, userID int) error {
	var id int

	err := r.db.QueryRowContext(ctx, `
	DELETE FROM comment
	WHERE id = $1 AND author_id = $2
	RETURNING id
	`, commentID, userID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrForbidden
	}
	if err != nil {
		return err
	}

	return nil
}
