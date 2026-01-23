func buildURL(category string page int) string{
	if page == 1 {
		return fmt.Sprintf(
			"https://books.toscrape.com/catalogue/category/books/category/books/%s/index.html",
			category,
		)
	}
	return fmt.Sprintf(
		"https://books.toscrape.com/catalogue/category/books/%s/page-%d.html",
		category,
		page,
	)
}