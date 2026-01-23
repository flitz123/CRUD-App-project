func parseBooks(resp *http.Response, category string) ([]Book, error) {
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	books := []Book{}

	doc.Find(".product_pod").Each(func(i int, s *goquery.Selection) {
		title, _ := s.Find("h3 a").Attr("title")
		priceText := s.Find(".price_color").Text()
		rating, _ := s.Find("p").Attr("class")
		stockText := s.Find(".instock").Text()

		book := Book{
			Title:    title,
			Price:    parsePrice(priceText),
			Rating:   rating,
			InStock:  strings.Contains(stockText, "In Stock"),
			Category: category,
		}

		books = append(books, book)
	})

	return books, nil
}

func parsePrice(price string) float64 {
	price = strings.Replace(price, "£", "", 1)
	value, _ := strconv.ParseFloat(price, 64)
	return value
}