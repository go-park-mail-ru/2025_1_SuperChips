package repository

import (
	"context"

	starService "github.com/go-park-mail-ru/2025_1_SuperChips/star"
)

func (p *pgPinStorage) SetStarProperty(ctx context.Context, userID int, pinID int) error {
	result, err := p.db.ExecContext(ctx, `
		UPDATE flow
		SET is_star = true
		WHERE author_id = $1 AND id = $2
	`, userID, pinID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return starService.ErrPinNotFound
	}
	if rowsAffected >= 2 {
		return starService.ErrInconsistentDataInDB
	}
	
	return nil
}

func (p *pgPinStorage) UnSetStarProperty(ctx context.Context, userID int, pinID int) error {
	result, err := p.db.ExecContext(ctx, `
		UPDATE flow
		SET is_star = false
		WHERE author_id = $1 AND id = $2
	`, userID, pinID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return starService.ErrPinNotFound
	}
	if rowsAffected >= 2 {
		return starService.ErrInconsistentDataInDB
	}
	
	return nil
}

func (p *pgPinStorage) ReassignStarProperty(ctx context.Context, userID int, oldPinID int, newPinID int) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		UPDATE flow
		SET is_star = false
		WHERE author_id = $1 AND id = $2
	`, userID, oldPinID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return starService.ErrPinNotFound
	}
	if rowsAffected >= 2 {
		return starService.ErrInconsistentDataInDB
	}

	result, err = tx.ExecContext(ctx, `
		UPDATE flow
		SET is_star = true
		WHERE author_id = $1 AND id = $2
	`, userID, newPinID)
	if err != nil {
		return err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return starService.ErrPinNotFound
	}
	if rowsAffected >= 2 {
		return starService.ErrInconsistentDataInDB
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}