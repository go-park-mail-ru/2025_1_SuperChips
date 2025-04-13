package pincrud

import (
	"mime/multipart"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

func NewPinCRUDService(r PinCRUDRepository, imgStrg FileRepository) *PinCRUDService {
	return &PinCRUDService{
		rep:     r,
		imgStrg: imgStrg,
	}
}

func (s *PinCRUDService) GetPublicPin(pinID uint64) (domain.PinData, error) {
	data, err := s.rep.GetPin(pinID)
	if err != nil {
		return domain.PinData{}, err
	}
	if data.IsPrivate {
		return domain.PinData{}, ErrForbidden
	}

	return data, nil
}

func (s *PinCRUDService) GetAnyPin(pinID uint64, userID uint64) (domain.PinData, error) {
	data, err := s.rep.GetPin(pinID)
	if err != nil {
		return domain.PinData{}, err
	}
	if data.AuthorID != userID {
		return domain.PinData{}, ErrForbidden
	}
	return data, nil
}

func (s *PinCRUDService) DeletePin(pinID uint64, userID uint64) error {
	data, err := s.rep.GetPin(pinID)
	if data.AuthorID != userID {
		return ErrForbidden
	}
	err = s.rep.DeletePin(pinID, userID)
	if err != nil {
		return err
	}
	err = s.imgStrg.Delete(data.MediaURL)
	if err != nil {
		return err
	}
	return nil
}

func (s *PinCRUDService) UpdatePin(patch domain.PinDataUpdate, userID uint64) error {
	data, err := s.rep.GetPin(*patch.FlowID)
	if data.AuthorID != userID {
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
