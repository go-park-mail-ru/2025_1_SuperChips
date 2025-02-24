package main

import (
	"net/http"
	"fmt"
	"os"
	"github.com/go-park-mail-ru/2025_1_SuperChips/auth/internal/handler"
	"github.com/go-park-mail-ru/2025_1_SuperChips/auth/internal/database"
)

func main() {
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	database.InitializeDB(connectionString)
	defer database.Close()

	http.HandleFunc("/api/v1/auth/login", handler.HandleLogin)
	http.ListenAndServe(":8080", nil)
}