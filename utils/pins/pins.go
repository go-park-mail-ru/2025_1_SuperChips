package pins

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/go-park-mail-ru/2025_1_SuperChips/utils/image"
)

// временная функция, добавляющая все файлы из директории в базу данных
func AddAllPins(db *sql.DB, baseDir string) error {
	files, err := os.ReadDir(baseDir)
	if err != nil {
		return err
	}

	id := 1

	_, err = db.Exec("INSERT INTO flow_user (username, avatar, public_name, email, password) VALUES ($1, $2, $3, $4, $5)", "admin", "", "admin", "admin@yourflow", "admin")
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && image.IsImageFile(file.Name()) {
			_, err := db.Exec("INSERT INTO flow (title, media_url, author_id) VALUES ($1, $2, $3)", fmt.Sprintf("Header %d", id), fmt.Sprintf("https://yourflow.ru/static/img/%s", file.Name()), 1)
			if err != nil {
				return err
			}
			id++
		}
	}

	return nil
}