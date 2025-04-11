package board

import (
	"errors"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

type BoardRepository interface {
	CreateBoard(board *domain.Board, username string) (int, error)
	DeleteBoard(boardID, userID int) error
	AddToBoard(boardID, userID, flowID int) error      // == update board
	DeleteFromBoard(boardID, userID, flowID int) error // == update board
	UpdateBoard(boardID, userID int, newName string, isPrivate bool) error
	GetBoard(boardID int) (domain.Board, error)
	GetUserPublicBoards(username string) ([]domain.Board, error)   // == get board
	GetUserAllBoards(userID int) ([]domain.Board, error)
	GetBoardFlow(boardID, userID, page int) ([]domain.PinData, error)
}

type BoardService struct {
	repo BoardRepository
}

var (
	ErrForbidden = errors.New("forbidden")
)

func NewBoardService(repo BoardRepository) *BoardService {
	return &BoardService{
		repo: repo,
	}
}

func (b *BoardService) CreateBoard(board domain.Board, username string) (int, error) {
	if err := board.ValidateBoard(); err != nil {
		return 0, err
	}

	userID, err := b.repo.CreateBoard(&board, username)
	if err != nil {
		return 0, err
	}

	if board.AuthorID != userID {
		return 0, ErrForbidden
	}

	return board.Id, nil
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

func (b *BoardService) UpdateBoard(boardID, userID int, newName string, isPrivate bool) error {
	if boardID <= 0 || userID <= 0 {
		return domain.ErrValidation
	}

	if newName == "" {
		return domain.ErrNoBoardName
	}

	if err := b.repo.UpdateBoard(boardID, userID, newName, isPrivate); err != nil {
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

func (b *BoardService) GetBoard(boardID, userID int, authorized bool) (domain.Board, error) {
	if boardID <= 0 || userID < 0 {
		return domain.Board{}, domain.ErrValidation
	}

	board, err := b.repo.GetBoard(boardID)
	if err != nil {
		return domain.Board{}, err
	}

	if board.IsPrivate {
		if !authorized {
			return domain.Board{}, ErrForbidden
		} else {
			if board.AuthorID != userID {
				return domain.Board{}, ErrForbidden
			}
		}
	}

	return board, nil
}
 
func (b *BoardService) GetUserPublicBoards(username string) ([]domain.Board, error) {
	return b.repo.GetUserPublicBoards(username)
}

func (b *BoardService) GetUserAllBoards(userID int) ([]domain.Board, error) {
	return b.repo.GetUserAllBoards(userID)
}

func (b *BoardService) GetBoardFlow(boardID, userID, page int, authorized bool) ([]domain.PinData, error) {
	if boardID <= 0 || userID < 0 || page <= 0 {
		return nil, domain.ErrValidation
	}

	flows, err := b.repo.GetBoardFlow(boardID, userID, page)
	if err != nil {
		return nil, err
	}

	return flows, nil
}

