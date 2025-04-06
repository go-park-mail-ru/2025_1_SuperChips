package board

import (
	"errors"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type BoardRepository interface {
	CreateBoard(board domain.Board) error
	DeleteBoard(board domain.Board, userID int) error
	AddToBoard(board domain.Board, userID, flowID int) error      // == update board
	DeleteFromBoard(board domain.Board, userID, flowID int) error // == update board
	UpdateBoard(board domain.Board, userID int, newName string, isPrivate bool) error
	GetBoard(name string, authorID, userID int) (domain.Board, error) // == get board
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

func (b *BoardService) DeleteBoard(board domain.Board, userID int) error {
	if board.AuthorID != userID {
		return ErrForbidden
	}

	if err := board.ValidateBoard(); err != nil {
		return err
	}

	if err := b.DeleteBoard(board, userID); err != nil {
		return err
	}

	return nil
}

func (b *BoardService) AddToBoard(board domain.Board, userID, flowID int) error {
	if board.AuthorID != userID {
		return ErrForbidden
	}

	if err := board.ValidateBoard(); err != nil {
		return err
	}

	if err := b.repo.AddToBoard(board, userID, flowID); err != nil {
		return err
	}

	return nil
}

func (b *BoardService) UpdateBoard(board domain.Board, userID int, newName string, isPrivate bool) error {
	if board.AuthorID != userID {
		return ErrForbidden
	}

	if err := board.ValidateBoard(); err != nil {
		return err
	}

	newBoard := domain.Board{
		Name: newName,
		AuthorID: board.AuthorID,
		IsPrivate: isPrivate,
	}

	if err := newBoard.ValidateBoard(); err != nil {
		return err
	}

	if err := b.repo.UpdateBoard(board, userID, newName, isPrivate); err != nil {
		return err
	}

	return nil
}

func (b *BoardService) DeleteFromBoard(board domain.Board, userID, flowID int) error {
	if board.AuthorID != userID {
		return ErrForbidden
	}

	if err := board.ValidateBoard(); err != nil {
		return err
	}

	if err := b.repo.DeleteFromBoard(board, userID, flowID); err != nil {
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

	board, err := b.repo.GetBoard(name, authorID, userID)
	if err != nil {
		return domain.Board{}, err
	}

	return board, nil
}

func (b *BoardService) GetUserPublicBoards(userID int) ([]domain.Board, error) {
	return b.repo.GetUserPublicBoards(userID)
}

func (b *BoardService) GetUserAllBoards(userID int) ([]domain.Board, error) {
	return b.repo.GetUserAllBoards(userID)
}
