package utils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func GetInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func GetPages() int {
	pagesStr := GetInput("Enter number of pages to scrape: ")
	pages, err := strconv.Atoi(pagesStr)
	if err != nil {
		return 1
	}
	return pages
}