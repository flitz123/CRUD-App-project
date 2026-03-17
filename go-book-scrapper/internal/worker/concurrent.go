package worker

import (
	"go-book-scrapper/internal/models"
	"go-book-scrapper/internal/scrapper"
	"sync"
)

func ScrapeConcurrently(category string, pageLimit int) ([]models.Book, error) {
	var wg sync.WaitGroup
	bookChan := make(chan []models.Book)
	allBooks := []models.Book{}

	for i := 1; i <= pageLimit; i++ {
		wg.Add(1)
		go func(page int) {
			defer wg.Done()
			url := scrapper.BuildURL(category, page)
			resp, err := scrapper.FetchHTML(url)
			if err != nil {
				return
			}
			books, _ := scrapper.ParseBooks(resp, category)
			bookChan <- books
		}(i)
	}

	go func() {
		wg.Wait()
		close(bookChan)
	}()

	for books := range bookChan {
		allBooks = append(allBooks, books...)
	}

	return allBooks, nil
}