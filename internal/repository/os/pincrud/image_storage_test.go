package repository

import (
	"bytes"
	"errors"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"

	pincrudService "github.com/go-park-mail-ru/2025_1_SuperChips/pincrud"
	"github.com/stretchr/testify/assert"
)

type testFile struct {
	*bytes.Reader
}

func (t testFile) Close() error {
	return nil
}

func fakeFileHeader(filename string) *multipart.FileHeader {
	return &multipart.FileHeader{Filename: filename}
}

func TestSave_Success(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewOSImageStorage(tmpDir)
	assert.NoError(t, err)

	content := []byte("dummy image data")

	file := testFile{bytes.NewReader(content)}

	header := fakeFileHeader("test.jpg")

	imgName, err := storage.Save(file, header)
	assert.NoError(t, err)
	assert.NotEmpty(t, imgName)

	imgPath := filepath.Join(tmpDir, imgName)
	_, err = os.Stat(imgPath)
	assert.NoError(t, err)

	savedContent, err := os.ReadFile(imgPath)
	assert.NoError(t, err)
	assert.Equal(t, content, savedContent)
}

func TestSave_InvalidExtension(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewOSImageStorage(tmpDir)
	assert.NoError(t, err)

	content := []byte("dummy data")
	file := testFile{bytes.NewReader(content)}

	header := fakeFileHeader("test.txt")
	_, err = storage.Save(file, header)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, pincrudService.ErrInvalidImageExt))
}

func TestSave_EmptyExtension(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewOSImageStorage(tmpDir)
	assert.NoError(t, err)

	content := []byte("dummy data")
	file := testFile{bytes.NewReader(content)}

	header := fakeFileHeader("test")
	_, err = storage.Save(file, header)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, pincrudService.ErrInvalidImageExt))
}

func TestDelete_Success(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewOSImageStorage(tmpDir)
	assert.NoError(t, err)

	fileName := "delete_me.jpg"
	filePath := filepath.Join(tmpDir, fileName)
	err = os.WriteFile(filePath, []byte("dummy content"), 0644)
	assert.NoError(t, err)

	err = storage.Delete(fileName)
	assert.NoError(t, err)

	_, err = os.Stat(filePath)
	assert.True(t, os.IsNotExist(err))
}

func TestDelete_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewOSImageStorage(tmpDir)
	assert.NoError(t, err)

	err = storage.Delete("nonexistent.jpg")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, pincrudService.ErrUntracked))
}
