package main

import (
	"fmt"
	"log"
	"net/http"
	"newsParser/handlers"
	"newsParser/parser"
	"newsParser/storage"
	"os"
)

func main() {
	connStr := fmt.Sprintf(
		"host=localhost port=5432 user=postgres password=%s dbname=news sslmode=disable",
		os.Getenv("DB_PASSWORD"),
	)

	db, err := storage.NewDB(connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err = storage.CreateTable(db); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Сервер запущен на порту 8080")
	h := handlers.NewHandler(db)
	parser.Scrape("https://ria.ru", db)
	http.HandleFunc("/result", h.HandleGet)
	http.ListenAndServe(":8080", nil)
}
