package board

import "github.com/go-park-mail-ru/2025_1_SuperChips/domain"

type BoardRepository interface {
	CreateBoard(name string, authorID int) error
	DeleteBoard(name string, authorID int) error
	AddToBoard(name string, authorID, flowID int) error       // == update board
	DeleteFromBoard(name string, authorID, flowID int) error  // == update board
	GetBoard(name string, authorID int) (domain.Board, error) // == get board
	GetUserPublicBoards(userID int) ([]domain.Board, error)   // == get board
	GetUserAllBoards(userID int) ([]domain.Board, error)
}

type BoardService struct {
	repo BoardRepository
}

func (b *BoardService) CreateBoard(name string, authorID int) error {
	board := domain.Board{
		Name:     name,
		AuthorID: authorID,
	}

	if err := board.ValidateBoard(); err != nil {
		return err
	}

	if err := b.repo.CreateBoard(name, authorID); err != nil {
		return err
	}

	return nil
}

func (b *BoardService) DeleteBoard(name string, authorID int) error {
	board := domain.Board{
		Name:     name,
		AuthorID: authorID,
	}

	if err := board.ValidateBoard(); err != nil {
		return err
	}

	if err := b.DeleteBoard(name, authorID); err != nil {
		return err
	}

	return nil
}

func (b *BoardService) AddToBoard(name string, authorID, flowID int) error {
	board := domain.Board{
		Name:     name,
		AuthorID: authorID,
	}

	if err := board.ValidateBoard(); err != nil {
		return err
	}

	if err := b.repo.AddToBoard(name, authorID, flowID); err != nil {
		return err
	}

	return nil
}

func (b *BoardService) DeleteFromBoard(name string, authorID, flowID int) error {
	board := domain.Board{
		Name:     name,
		AuthorID: authorID,
	}

	if err := board.ValidateBoard(); err != nil {
		return err
	}

	if err := b.repo.DeleteFromBoard(name, authorID, flowID); err != nil {
		return err
	}

	return nil
}

func (b *BoardService) GetBoard(name string, authorID int) (domain.Board, error) {
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

func (b *BoardService) GetUserPublicBoards(userID int) ([]domain.Board, error) {
	return b.repo.GetUserPublicBoards(userID)
}

func (b *BoardService) GetUserAllBoards(userID int) ([]domain.Board, error) {
	return b.repo.GetUserAllBoards(userID)
}
