var wg sync.WaitGroup
bookChan := make(chan []Book)

for i := 1; i <= pagelimit; i++ {
	wg.Add(1)
	go func(page int) {
		defer wg.Done()
		url := buildURL(category, page)
		resp, err := fetchHTML(url)
		if err != nil {
			return
		}
		books, _ := parseBooks(resp, category)
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