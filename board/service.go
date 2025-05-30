package board

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	imageUtil "github.com/go-park-mail-ru/2025_1_SuperChips/utils/image"
)

type BoardRepository interface {
	GetUsernameID(ctx context.Context, username string, userID int) (int, error)                                    // получить айди юзернейма
	CreateBoard(ctx context.Context, board *domain.Board, username string, userID int) error                        // создание доски
	DeleteBoard(ctx context.Context, boardID, userID int) error                                                     // удаление доски
	AddToBoard(ctx context.Context, boardID, userID, flowID int) error                                              // добавление пина в доску
	DeleteFromBoard(ctx context.Context, boardID, userID, flowID int) error                                         // удаление пина из доски
	UpdateBoard(ctx context.Context, boardID, userID int, newName string, isPrivate bool) error                     // обновление данных доски
	GetBoard(ctx context.Context, boardID, userID, previewNum, previewStart int) (domain.Board, []string, error)    // получить доску
	GetUserPublicBoards(ctx context.Context, username string, previewNum, previewStart int) ([]domain.Board, error) // получить публичные доски пользователя
	GetUserAllBoards(ctx context.Context, userID, previewNum, previewStart int) ([]domain.Board, error)             // получтиь все доски пользователя
	GetBoardFlow(ctx context.Context, boardID, userID, page, pageSize int) ([]domain.PinData, error)                // получить пины доски (с пагинацией)
}

type PinRepository interface {
	GetFromBoard(ctx context.Context, boardID, userID, flowID int) (domain.PinData, int, error)
}

type BoardSharingRepository interface {
	GetCoauthorsIDs(ctx context.Context, boardID int) ([]int, error)
}

type BoardService struct {
	repo     BoardRepository
	repoPin  PinRepository
	repoShr  BoardSharingRepository
	baseURL  string
	imageDir string
}

var (
	ErrForbidden = errors.New("forbidden")
)

const (
	previewNum   = 3
	previewStart = 0
)

func NewBoardService(repo BoardRepository, repoPin PinRepository, repoShr BoardSharingRepository, baseURL, imageDir string) *BoardService {
	return &BoardService{
		repo:     repo,
		repoPin:  repoPin,
		repoShr:  repoShr,
		baseURL:  baseURL,
		imageDir: imageDir,
	}
}

func (b *BoardService) CreateBoard(ctx context.Context, board domain.Board, username string, userID int) (int, error) {
	id, err := b.repo.GetUsernameID(ctx, username, userID)
	if err != nil {
		return 0, err
	}

	// пользователь пытается добавить в чужие доски
	if id != userID {
		return 0, ErrForbidden
	}

	if err := b.repo.CreateBoard(ctx, &board, username, userID); err != nil {
		return 0, err
	}

	return board.ID, nil
}

func (b *BoardService) DeleteBoard(ctx context.Context, boardID, userID int) error {
	if err := b.repo.DeleteBoard(ctx, boardID, userID); err != nil {
		return err
	}

	return nil
}

func (b *BoardService) AddToBoard(ctx context.Context, boardID, userID, flowID int) error {

	if err := b.repo.AddToBoard(ctx, boardID, userID, flowID); err != nil {
		return err
	}

	return nil
}

func (b *BoardService) UpdateBoard(ctx context.Context, boardID, userID int, newName string, isPrivate bool) error {
	if err := b.repo.UpdateBoard(ctx, boardID, userID, newName, isPrivate); err != nil {
		return err
	}

	return nil
}

func (b *BoardService) GetFromBoard(ctx context.Context, boardID, userID, flowID int, authorized bool) (domain.PinData, error) {
	data, authorID, err := b.repoPin.GetFromBoard(ctx, boardID, userID, flowID)
	if err != nil {
		return domain.PinData{}, err
	}

	// Это обращение можно упразднить и делать проверку прав доступа через board.IsEditable, однако изначальный вариант работает, а времени на исправление возможных багов уже нет.
	coauthorsIDs, err := b.repoShr.GetCoauthorsIDs(ctx, boardID)
	if err != nil {
		return domain.PinData{}, err
	}

	board, _, err := b.repo.GetBoard(ctx, boardID, userID, 0, 0)
	if err != nil {
		return domain.PinData{}, err
	}

	if board.IsPrivate {
		if !authorized {
			return domain.PinData{}, ErrForbidden
		}
		// if !board.IsEditable {
		if data.IsPrivate && !(authorID == userID || slices.Contains(coauthorsIDs, userID)) {
			return domain.PinData{}, ErrForbidden
		}
	}

	return data, nil
}

func (b *BoardService) DeleteFromBoard(ctx context.Context, boardID, userID, flowID int) error {
	if err := b.repo.DeleteFromBoard(ctx, boardID, userID, flowID); err != nil {
		return err
	}

	return nil
}

func (b *BoardService) GetBoard(ctx context.Context, boardID, userID int, authorized bool) (domain.Board, error) {
	board, colors, err := b.repo.GetBoard(ctx, boardID, userID, previewNum, previewStart)
	if err != nil {
		return domain.Board{}, err
	}

	coauthorsIDs, err := b.repoShr.GetCoauthorsIDs(ctx, boardID)
	if err != nil {
		return domain.Board{}, err
	}

	if board.IsPrivate {
		if !authorized {
			return domain.Board{}, ErrForbidden
		}
		if board.AuthorID != userID && !slices.Contains(coauthorsIDs, userID) {
			return domain.Board{}, ErrForbidden
		}
	}

	for i := range board.Preview {
		board.Preview[i].MediaURL = b.generateImageURL(board.Preview[i].MediaURL)
	}

	if len(colors) > 0 && len(colors)%4 == 0 {
		sortedColors, err := imageUtil.SortColorsByLuminance(colors)
		if err != nil {
			return domain.Board{}, fmt.Errorf("failed to sort colors: %w", err)
		}

		sectionSize := len(sortedColors) / 4
		avgColors := make([]string, 4)

		for i := range 4 {
			start := i * sectionSize
			end := start + sectionSize
			section := sortedColors[start:end]

			avgHex, err := imageUtil.AverageHexColors(section)
			if err != nil {
				return domain.Board{}, fmt.Errorf("failed to average colors: %w", err)
			}
			avgColors[i] = avgHex
		}

		board.Gradient = avgColors
	}

	return board, nil
}

// todo: paginate this
func (b *BoardService) GetUserPublicBoards(ctx context.Context, username string) ([]domain.Board, error) {
	boards, err := b.repo.GetUserPublicBoards(ctx, username, previewNum, previewStart)
	if err != nil {
		return []domain.Board{}, err
	}

	for i := range boards {
		for j := range boards[i].Preview {
			boards[i].Preview[j].MediaURL = b.generateImageURL(boards[i].Preview[j].MediaURL)
		}
	}

	return boards, nil
}

// todo: paginate this
func (b *BoardService) GetUserAllBoards(ctx context.Context, userID int) ([]domain.Board, error) {
	boards, err := b.repo.GetUserAllBoards(ctx, userID, previewNum, previewStart)
	if err != nil {
		return []domain.Board{}, err
	}

	for i := range boards {
		for j := range boards[i].Preview {
			boards[i].Preview[j].MediaURL = b.generateImageURL(boards[i].Preview[j].MediaURL)
		}
	}

	return boards, nil
}

func (b *BoardService) GetBoardFlow(ctx context.Context, boardID, userID, page, pageSize int, authorized bool) ([]domain.PinData, error) {
	flows, err := b.repo.GetBoardFlow(ctx, boardID, userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	for i := range flows {
		flows[i].MediaURL = b.generateImageURL(flows[i].MediaURL)
	}

	return flows, nil
}

func (p *BoardService) generateImageURL(filename string) string {
	return p.baseURL + filepath.Join(strings.ReplaceAll(p.imageDir, ".", ""), filename)
}
