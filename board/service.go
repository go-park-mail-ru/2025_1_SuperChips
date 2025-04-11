package board

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/validator"
	"github.com/go-park-mail-ru/2025_1_SuperChips/utils/wrapper"
)

type BoardRepository interface {
	CreateBoard(ctx context.Context, board *domain.Board, username string, userID int) error    // создание доски
	DeleteBoard(ctx context.Context, boardID, userID int) error                                 // удаление доски
	AddToBoard(ctx context.Context, boardID, userID, flowID int) error                          // добавление пина в доску
	DeleteFromBoard(ctx context.Context, boardID, userID, flowID int) error                     // удаление пина из доски
	UpdateBoard(ctx context.Context, boardID, userID int, newName string, isPrivate bool) error // обновление данных доски
	GetBoard(ctx context.Context, boardID int) (domain.Board, error)                            // получить доску
	GetUserPublicBoards(ctx context.Context, username string) ([]domain.Board, error)           // получить публичные доски пользователя
	GetUserAllBoards(ctx context.Context, userID int) ([]domain.Board, error)                   // получтиь все доски пользователя
	GetBoardFlow(ctx context.Context, boardID, userID, page, pageSize int) ([]domain.PinData, error)      // получить пины доски (с пагинацией)
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

func (b *BoardService) CreateBoard(ctx context.Context, board domain.Board, username string, userID int) (int, error) {
	v := validator.New()

	if v.Check(board.Name == "", "name", "cannot be empty") {
		return 0, wrapper.WrapError(domain.ErrValidation, v.GetError("name"))
	}

	if v.Check(len(board.Name) < 64, "name", "cannot be longer 64") {
		return 0, wrapper.WrapError(domain.ErrValidation, v.GetError("name"))
	}

	if err := b.repo.CreateBoard(ctx, &board, username, userID); err != nil {
		return 0, err
	}

	return board.ID, nil
}

func (b *BoardService) DeleteBoard(ctx context.Context, boardID, userID int) error {
	v := validator.New()

	if v.Check(userID <= 0 || boardID <= 0, "id", "cannot be less or equal to zero") {
		return domain.ErrValidation
	}

	if err := b.repo.DeleteBoard(ctx, boardID, userID); err != nil {
		return err
	}

	return nil
}

func (b *BoardService) AddToBoard(ctx context.Context, boardID, userID, flowID int) error {
	v := validator.New()

	if v.Check(flowID <= 0 || boardID <= 0 || userID <= 0, "id", "cannot be less or equal to zero") {
		return domain.ErrValidation
	}

	if err := b.repo.AddToBoard(ctx, boardID, userID, flowID); err != nil {
		return err
	}

	return nil
}

func (b *BoardService) UpdateBoard(ctx context.Context, boardID, userID int, newName string, isPrivate bool) error {
	v := validator.New()

	if v.Check(boardID <= 0 || userID <= 0, "id", "cannot be less or equal to zero") {
		return domain.ErrValidation
	}

	if v.Check(newName == "", "name", "cannot be empty") {
		return domain.ErrNoBoardName
	}

	if err := b.repo.UpdateBoard(ctx, boardID, userID, newName, isPrivate); err != nil {
		return err
	}

	return nil
}

func (b *BoardService) DeleteFromBoard(ctx context.Context, boardID, userID, flowID int) error {
	v := validator.New()

	if v.Check(flowID <= 0 || boardID <= 0 || userID <= 0, "id", "cannot be less or equal to zero") {
		return domain.ErrValidation
	}

	if err := b.repo.DeleteFromBoard(ctx, boardID, userID, flowID); err != nil {
		return err
	}

	return nil
}

func (b *BoardService) GetBoard(ctx context.Context, boardID, userID int, authorized bool) (domain.Board, error) {
	v := validator.New()

	if v.Check(boardID <= 0 || userID < 0, "id", "board id cannot be less or equal to zero or user id cannot be less than zero") {
		return domain.Board{}, domain.ErrValidation
	}

	board, err := b.repo.GetBoard(ctx, boardID)
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

func (b *BoardService) GetUserPublicBoards(ctx context.Context, username string) ([]domain.Board, error) {
	return b.repo.GetUserPublicBoards(ctx, username)
}

func (b *BoardService) GetUserAllBoards(ctx context.Context, userID int) ([]domain.Board, error) {
	return b.repo.GetUserAllBoards(ctx, userID)
}

func (b *BoardService) GetBoardFlow(ctx context.Context, boardID, userID, page, pageSize int, authorized bool) ([]domain.PinData, error) {
	v := validator.New()	

	if v.Check(boardID <= 0 || userID < 0 || page <= 0, "id and page", "cannot be less than zero") {
		return nil, domain.ErrValidation
	}

	if v.Check(pageSize < 1 || pageSize > 30, "page size", "cannot be less than one and more than 30") {
		return nil, domain.ErrValidation
	}

	flows, err := b.repo.GetBoardFlow(ctx, boardID, userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	return flows, nil
}
