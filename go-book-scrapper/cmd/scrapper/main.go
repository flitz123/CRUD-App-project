package main

import (
	"log"

	"go-book-scrapper/internal/utils"
	"go-book-scrapper/internal/worker"
)

func main() {
	category := utils.GetInput("Enter book category: ")
	pages := utils.GetPages()

	books, err := worker.ScrapeConcurrently(category, pages)
	if err != nil {
		log.Fatal(err)
	}

	err = utils.SaveToJSON("books.json", books)
	if err != nil {
		log.Fatal(err)
	}
}