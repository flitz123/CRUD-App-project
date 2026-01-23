package main

import{
	"fmt"
	"bufio"
	"os"
	"strings"
}

func getUserInput(prompt string) string{
	reader := bufio.NewReader(os.StdIn)
    fmt.Print(prompt)
	input, _ := reader.NewString('\n')
    return strings.TrimSpace(input)
}

category := getUserInput("Enter book category: ") 
pages := getUserInput("Enter number of pages to scrape: ") 