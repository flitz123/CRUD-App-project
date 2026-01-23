func saveToJSON(filename string, data[]Book) error{
	fil, err := os.Create(filename)
	if err !=nil{
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(data)
}

err := saveToJSON("books.json", allBooks)
if err != nil{
	log.Fatal(err)
}