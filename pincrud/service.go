package pincrud

import (
	"context"
	"mime/multipart"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

const UnauthorizedID = 0

type PinCRUDService struct {
	rep     PinCRUDRepository
	imgStrg FileRepository
}

func NewPinCRUDService(r PinCRUDRepository, imgStrg FileRepository) *PinCRUDService {
	return &PinCRUDService{
		rep:     r,
		imgStrg: imgStrg,
	}
}

func (s *PinCRUDService) GetPublicPin(ctx context.Context, pinID uint64) (domain.PinData, error) {
	data, _, err := s.rep.GetPin(ctx, pinID, UnauthorizedID)
	if err != nil {
		return domain.PinData{}, err
	}
	if data.IsPrivate {
		return domain.PinData{}, ErrForbidden
	}

	return data, nil
}

func (s *PinCRUDService) GetAnyPin(ctx context.Context, pinID uint64, userID uint64) (domain.PinData, error) {
	data, authorID, err := s.rep.GetPin(ctx, pinID, userID)
	if err != nil {
		return domain.PinData{}, err
	}
	if authorID != userID && data.IsPrivate {
		return domain.PinData{}, ErrForbidden
	}
	return data, nil
}

func (s *PinCRUDService) DeletePin(ctx context.Context, pinID uint64, userID uint64) error {
	mediaURL, authorID, err := s.rep.GetPinCleanMediaURL(ctx, pinID)
	if err != nil {
		return err
	}
	if authorID != userID {
		return ErrForbidden
	}
	err = s.rep.DeletePin(ctx, pinID, userID)
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
	_, authorID, err := s.rep.GetPin(ctx, *patch.FlowID, userID)
	if err != nil {
		return err
	}
	if authorID != userID {
		return ErrForbidden
	}
	err = s.rep.UpdatePin(ctx, patch, userID)
	if err != nil {
		return err
	}
	return nil
}

func (s *PinCRUDService) CreatePin(ctx context.Context, data domain.PinDataCreate, file multipart.File, header *multipart.FileHeader, userID uint64) (uint64, error) {
	imgName, err := s.imgStrg.Save(file, header)
	if err != nil {
		return 0, err
	}
	pinID, err := s.rep.CreatePin(ctx, data, imgName, userID)
	if err != nil {
		return 0, err
	}
	return pinID, nil
}
