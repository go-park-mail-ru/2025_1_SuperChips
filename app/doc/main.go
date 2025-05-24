package main

import (
	"log"
	"net/http"

	_ "github.com/go-park-mail-ru/2025_1_SuperChips/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)


func main() {
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	log.Println("Server started at http://localhost:8040")
	log.Println("Swagger UI available at http://localhost:8040/swagger/index.html")
	log.Fatal(http.ListenAndServe(":8040", nil))
}