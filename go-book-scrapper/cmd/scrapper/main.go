func main() {
	category := utils.GetInput("Category:")
	pages := utils.GetPages()

	allBooks := []models.Book{}

	for page := 1; page <= pages; page++ {
		url := scrapper.BuildURL(category, page)

		resp, err := scrapper.FetchHTML(url)
		if err != nil {
			log.Println(err)
			continue
		}

		books, err := worker.ScrapeConcurrently(category, pages)
		if err != nil {
			log.Fatal(err)
		}

	}

	utils.SaveToJson("books.json", allBooks)
}