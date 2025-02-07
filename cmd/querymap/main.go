package main

import (
	"fmt"
	"github.com/KoNekoD/go-querymap/pkg/querymap"
	"net/url"
)

func main() {
	fmt.Println("Please enter the URL:")

	var rawUrl string

	scanLn, err := fmt.Scanln(&rawUrl)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if scanLn == 0 {
		fmt.Println("No input")
		return
	}

	fmt.Println("URL:", rawUrl)

	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		fmt.Println("Error parse:", err)
		return
	}

	qm := querymap.FromURL(parsedUrl)

	fmt.Println(qm)
}
