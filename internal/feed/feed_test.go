package feed

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
)

var cfg configs.Config

func init() {
	config, err := configs.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("Error while loading config for test: %s", err)
	}

	cfg = config
}

// Мокирование данных конфигурации для тестов
func mockConfig() configs.Config {
	return configs.Config{
		ImageBaseDir: "test_images", // Путь к папке с тестовыми изображениями
		IpAddress:    "localhost",
		Port:         "8080",
	}
}

func TestIsImageFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"image.jpg", true},
		{"image.jpeg", true},
		{"image.png", true},
		{"image.gif", true},
		{"image.bmp", true},
		{"image.tiff", true},
		{"image.webp", true},
		{"document.pdf", false},
		{"image.txt", false},
		{"image", false},
	}

	for i, c := range tests {
		t.Run(c.filename, func(t *testing.T) {
			res := isImageFile(c.filename)
			if res != c.expected {
				printDifference(t, i, "isImageFile", res, c.expected)
			}
		})
	}
}

func TestCountPages(t *testing.T) {
	// Добавляем 5 изображений.
	storage := NewPinSliceStorage(cfg)
	storage.Pins = append(storage.Pins, PinData{}, PinData{}, PinData{}, PinData{}, PinData{})

	tests := []struct {
		pageSize int
		expected int
	}{
		{2, 3},  // 5 изображений, 2 на странице -> 3 страницы.
		{5, 1},  // 5 изображений, 5 на странице -> 1 страница.
		{10, 1}, // 5 изображений, 10 на странице -> 1 страница.
	}

	for i, c := range tests {
		t.Run("", func(t *testing.T) {
			result := storage.CountPages(c.pageSize)
			if result != c.expected {
				printDifference(t, i, "CountPages", result, c.expected)
			}
		})
	}
}

func TestGetPinPage(t *testing.T) {
	// Добавляем 5 фиктивных изображений.
	storage := NewPinSliceStorage(cfg)
	storage.Pins = append(storage.Pins, PinData{}, PinData{}, PinData{}, PinData{}, PinData{})

	tests := []struct {
		page     int
		pageSize int
		expected int
	}{
		{1, 2, 2}, // 1-ая страница с размером страницы 2 -> 2 элемента
		{2, 2, 2}, // 2-ая страница с размером страницы 2 -> 2 элемента
		{3, 2, 1}, // 3-ая страница с размером страницы 2 -> 1 элемент
		{1, 5, 5}, // 1-ая страница с размером страницы 5 -> 5 элементов
		{2, 5, 0}, // 2-ая страница с размером страницы 5 -> 0 элементов
	}

	for i, c := range tests {
		t.Run("", func(t *testing.T) {
			result := storage.GetPinPage(c.page, c.pageSize)
			if len(result) != c.expected {
				printDifference(t, i, "GetPinPage", len(result), c.expected)
			}
		})
	}
}

func TestInitialize(t *testing.T) {
	cfg := mockConfig()

	// Создаем папку для тестовых изображений.
	dir := cfg.ImageBaseDir
	os.MkdirAll(dir, os.ModePerm)
	defer os.RemoveAll(cfg.ImageBaseDir)

	// Создаем тестовые файлы.
	files := []string{"image.jpg", "image.png", "image.txt"}
	for _, file := range files {
		_, err := os.Create(filepath.Join(dir, file))
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	storage := NewPinSliceStorage(cfg)

	// Проверяем, что были извлечены корректные файлы (image.jpg и image.png).
	if len(storage.Pins) != 2 {
		printDifference(t, 0, "Pins Count", len(storage.Pins), 2)
	}

	// Проверяем корректность данных для первого изображения.
	if !strings.Contains(storage.Pins[0].Image, "image.jpg") {
		printDifference(t, 1, "Pin Image", storage.Pins[0].Image, "image.jpg")
	}
}

func printDifference(t *testing.T, num int, name string, got any, exp any) {
	t.Errorf("[%d] wrong %v", num, name)
	t.Errorf("--> got     : %+v", got)
	t.Errorf("--> expected: %+v", exp)
}
