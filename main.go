package main

import (
	"fmt"
	"net/http"
	"word-search-in-files/pkg/searcher"
)

func main() {
	http.Handle("/files/search/", http.HandlerFunc(searcher.SearchFiles))

	fmt.Println("Start server on :8080")
	http.ListenAndServe(":8080", nil)
}
