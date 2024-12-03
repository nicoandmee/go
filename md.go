package main

import (
	"fmt"
	"log"
	"net/http"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

func fetchBody(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
}

func main() {
	// TODO: Fetch from URL provided as positional argument


	input := `<strong>Bold Text</strong>`

	markdown, err := htmltomarkdown.ConvertString(input)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(markdown)
	// Output: **Bold Text**
}
