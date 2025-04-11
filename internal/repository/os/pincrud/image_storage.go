package repository

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	pincrudService "github.com/go-park-mail-ru/2025_1_SuperChips/pincrud"
	"github.com/go-park-mail-ru/2025_1_SuperChips/utils/image"
	"github.com/google/uuid"
)

type osImageStorage struct {
	imgDir string
}

func NewOSImageStorage(imgDir string) (*osImageStorage, error) {
	storage := &osImageStorage{
		imgDir: imgDir,
	}

	return storage, nil
}

func (strg *osImageStorage) Save(file multipart.File, header *multipart.FileHeader) (string, error) {
	if !image.IsImageFile(header.Filename) || filepath.Ext(header.Filename) == "" {
		return "", pincrudService.ErrInvalidImageExt
	}

	imgUUID := uuid.New()
	imgName := imgUUID.String() + filepath.Ext(header.Filename)
	imgPath := filepath.Join(strg.imgDir, imgName)

	dst, err := os.Create(imgPath)
	if err != nil {
		return "", pincrudService.ErrUntracked
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", pincrudService.ErrUntracked
	}

	return imgName, nil
}

func (strg *osImageStorage) Delete(imgName string) error {
	imgPath := filepath.Join(strg.imgDir, imgName)

	_, err := os.Stat(imgPath)
	if os.IsNotExist(err) {
		return pincrudService.ErrUntracked
	}

	err = os.Remove(imgPath)
	if err != nil {
		return pincrudService.ErrUntracked
	}

	return nil
}
