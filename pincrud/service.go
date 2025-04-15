package pincrud

import (
	"mime/multipart"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

const UnauthorizedID = 0

func NewPinCRUDService(r PinCRUDRepository, imgStrg FileRepository) *PinCRUDService {
	return &PinCRUDService{
		rep:     r,
		imgStrg: imgStrg,
	}
}

func (s *PinCRUDService) GetPublicPin(pinID uint64) (domain.PinData, error) {
	data, _, err := s.rep.GetPin(pinID, UnauthorizedID)
	if err != nil {
		return domain.PinData{}, err
	}
	if data.IsPrivate {
		return domain.PinData{}, ErrForbidden
	}

	return data, nil
}

func (s *PinCRUDService) GetAnyPin(pinID uint64, userID uint64) (domain.PinData, error) {
	data, authorID, err := s.rep.GetPin(pinID, userID)
	if err != nil {
		return domain.PinData{}, err
	}
	if authorID != userID && data.IsPrivate {
		return domain.PinData{}, ErrForbidden
	}
	return data, nil
}

func (s *PinCRUDService) DeletePin(pinID uint64, userID uint64) error {
	mediaURL, authorID, err := s.rep.GetPinCleanMediaURL(pinID)
	if err != nil {
		return err
	}
	if authorID != userID {
		return ErrForbidden
	}
	err = s.rep.DeletePin(pinID, userID)
	if err != nil {
		return err
	}
	err = s.imgStrg.Delete(mediaURL)
	if err != nil {
		return err
	}
	return nil
}

func (s *PinCRUDService) UpdatePin(patch domain.PinDataUpdate, userID uint64) error {
	_, authorID, err := s.rep.GetPin(*patch.FlowID, userID)
	if err != nil {
		return err
	}
	if authorID != userID {
		return ErrForbidden
	}
	err = s.rep.UpdatePin(patch, userID)
	if err != nil {
		return err
	}
	return nil
}

func (s *PinCRUDService) CreatePin(data domain.PinDataCreate, file multipart.File, header *multipart.FileHeader, userID uint64) (uint64, error) {
	imgName, err := s.imgStrg.Save(file, header)
	if err != nil {
		return 0, err
	}
	pinID, err := s.rep.CreatePin(data, imgName, userID)
	if err != nil {
		return 0, err
	}
	return pinID, nil
}
