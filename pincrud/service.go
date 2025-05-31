package pincrud

import (
	"context"
	"image"
	"log"
	"mime/multipart"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	imageUtil "github.com/go-park-mail-ru/2025_1_SuperChips/utils/image"
)

const UnauthorizedID = 0

type PinRepository interface {
	GetPin(ctx context.Context, pinID, userID uint64) (domain.PinData, uint64, error)
	DeletePin(ctx context.Context, pinID uint64, userID uint64) error
	UpdatePin(ctx context.Context, patch domain.PinDataUpdate, userID uint64) error
	CreatePin(ctx context.Context, data domain.PinDataCreate, imgName string, userID uint64) (uint64, error)
	GetPinCleanMediaURL(ctx context.Context, pinID uint64) (string, uint64, error)
}

type BoardRepository interface {
	AddToSavedBoard(ctx context.Context, userID, flowID int) error
}

type FileRepository interface {
	Save(file multipart.File, header *multipart.FileHeader) (string, error)
	Delete(imgName string) error
}

type PinCRUDService struct {
	pinRepo   PinRepository
	boardRepo BoardRepository
	imgStrg   FileRepository
}

func NewPinCRUDService(p PinRepository, b BoardRepository, imgStrg FileRepository) *PinCRUDService {
	return &PinCRUDService{
		pinRepo:   p,
		boardRepo: b,
		imgStrg:   imgStrg,
	}
}

func (s *PinCRUDService) GetPublicPin(ctx context.Context, pinID uint64) (domain.PinData, error) {
	data, _, err := s.pinRepo.GetPin(ctx, pinID, UnauthorizedID)
	if err != nil {
		return domain.PinData{}, err
	}
	
	return data, nil
}

func (s *PinCRUDService) GetAnyPin(ctx context.Context, pinID uint64, userID uint64) (domain.PinData, error) {
	data, _, err := s.pinRepo.GetPin(ctx, pinID, userID)
	if err != nil {
		return domain.PinData{}, err
	}

	return data, nil
}

func (s *PinCRUDService) DeletePin(ctx context.Context, pinID uint64, userID uint64) error {
	mediaURL, authorID, err := s.pinRepo.GetPinCleanMediaURL(ctx, pinID)
	if err != nil {
		return err
	}
	if authorID != userID {
		return ErrForbidden
	}
	err = s.pinRepo.DeletePin(ctx, pinID, userID)
	if err != nil {
		return err
	}
	err = s.imgStrg.Delete(mediaURL)
	if err != nil {
		return err
	}
	return nil
}

func (s *PinCRUDService) UpdatePin(ctx context.Context, patch domain.PinDataUpdate, userID uint64) error {
	_, authorID, err := s.pinRepo.GetPin(ctx, *patch.FlowID, userID)
	if err != nil {
		return err
	}
	if authorID != userID {
		return ErrForbidden
	}
	err = s.pinRepo.UpdatePin(ctx, patch, userID)
	if err != nil {
		return err
	}
	return nil
}

func (s *PinCRUDService) CreatePin(ctx context.Context, data domain.PinDataCreate, file multipart.File, header *multipart.FileHeader, extension string, userID uint64) (uint64, string, error) {
	imgName, err := s.imgStrg.Save(file, header)
	if err != nil {
		return 0, "", err
	}

	if _, err := file.Seek(0, 0); err != nil {
        return 0, "", err
    }

	img, _, err := image.Decode(file)
	if err != nil {
		return 0, "", err
	}

	width, height, err := imageUtil.GetImageDimensions(img)
	if err != nil {
		return 0, "", err
	}

	data.Width = width
	data.Height = height

	colors := imageUtil.GetImageMainColors(img)

	log.Printf("usecase colors len: %v", len(colors))

	data.Colors = colors

	pinID, err := s.pinRepo.CreatePin(ctx, data, imgName, userID)
	if err != nil {
		return 0, "", err
	}

	if err := s.boardRepo.AddToSavedBoard(ctx, int(userID), int(pinID)); err != nil {
		return 0, "", err
	}

	return pinID, imgName, nil
}
