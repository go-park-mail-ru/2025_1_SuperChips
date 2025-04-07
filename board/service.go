package board

import (
	"errors"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type BoardRepository interface {
	CreateBoard(board domain.Board) error
	DeleteBoard(boardID, userID int) error
	AddToBoard(boardID, userID, flowID int) error      // == update board
	DeleteFromBoard(boardID, userID, flowID int) error // == update board
	UpdateBoard(board domain.Board, userID int, newName *string, isPrivate *bool) error
	GetBoard(name string, authorID int) (domain.Board, error) // == get board
	GetBoardByID(boardID int) (domain.Board, error)
	GetUserPublicBoards(userID int) ([]domain.Board, error)   // == get board
	GetUserAllBoards(userID int) ([]domain.Board, error)
}

type BoardService struct {
	repo BoardRepository
}

var (
	ErrForbidden = errors.New("forbidden")
)

func (b *BoardService) CreateBoard(board domain.Board) error {
	if err := board.ValidateBoard(); err != nil {
		return err
	}

	if err := b.repo.CreateBoard(board); err != nil {
		return err
	}

	return nil
}

func (b *BoardService) DeleteBoard(boardID, userID int) error {
	if boardID <= 0 || userID <= 0 {
        return domain.ErrValidation
    }


	if err := b.repo.DeleteBoard(boardID, userID); err != nil {
		return err
	}

	return nil
}

func (b *BoardService) AddToBoard(boardID, userID, flowID int) error {
	if flowID <= 0 || boardID <= 0 || userID <= 0 {
        return domain.ErrValidation
    }

	if err := b.repo.AddToBoard(boardID, userID, flowID); err != nil {
		return err
	}

	return nil
}

func (b *BoardService) UpdateBoard(board domain.Board, userID int, newName *string, isPrivate *bool) error {
	if err := board.ValidateBoard(); err != nil {
		return err
	}

	newBoard := domain.Board{
		Name: *newName,
		AuthorID: board.AuthorID,
		IsPrivate: *isPrivate,
	}

	if err := newBoard.ValidateBoard(); err != nil {
		return err
	}

	if err := b.repo.UpdateBoard(board, userID, newName, isPrivate); err != nil {
		return err
	}

	return nil
}

func (b *BoardService) DeleteFromBoard(boardID, userID, flowID int) error {
	if flowID <= 0 || boardID <= 0 || userID <= 0 {
        return domain.ErrValidation
    }

	if err := b.repo.DeleteFromBoard(boardID, userID, flowID); err != nil {
		return err
	}

	return nil
}

func (b *BoardService) GetBoard(name string, authorID, userID int) (domain.Board, error) {
	board := domain.Board{
		Name:     name,
		AuthorID: authorID,
	}

	if err := board.ValidateBoard(); err != nil {
		return domain.Board{}, err
	}

	board, err := b.repo.GetBoard(name, authorID)
	if err != nil {
		return domain.Board{}, err
	}

	return board, nil
}

func (b *BoardService) GetBoardByID(boardID, userID int) (domain.Board, error) {
	board, err := b.repo.GetBoardByID(boardID)
	if err != nil {
		return domain.Board{}, err
	}

	if board.AuthorID != userID {
		return domain.Board{}, ErrForbidden
	}

	return board, nil
}
 
func (b *BoardService) GetUserPublicBoards(userID int) ([]domain.Board, error) {
	return b.repo.GetUserPublicBoards(userID)
}

func (b *BoardService) GetUserAllBoards(userID int) ([]domain.Board, error) {
	return b.repo.GetUserAllBoards(userID)
}
