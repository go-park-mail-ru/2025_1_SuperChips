package pincrud

import (
	"mime/multipart"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
)

func NewPinCRUDService(r PinCRUDRepository) *PinCRUDService {
	return &PinCRUDService{
		rep: r,
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

func (s *PinCRUDService) DeletePinByID(pinID uint64, userID uint64) error {
	data, err := s.rep.GetPin(pinID)
	if data.AuthorID != userID {
		return ErrForbidden
	}
	err = s.rep.DeletePinByID(pinID, userID)
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

func (s *PinCRUDService) CreatePin(data domain.PinDataCreate, file multipart.File, header *multipart.FileHeader, userID uint64) error {
	err := s.rep.CreatePin(data, file, header, userID)
	if err != nil {
		return err
	}
	return nil
}
