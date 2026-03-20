package main

import (
	"fmt"
	"os"

	"github.com/manus/go-find/internal/fetcher"
	"github.com/manus/go-find/internal/parser"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		_, err := fmt.Fprintf(os.Stderr, "Usage: find <query>\n")
		if err != nil {
			panic(err)
		}
		os.Exit(1)
	}

	query := args[0]

	body, err := fetcher.FetchSearchResults(query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to fetch results: %v\n", err)
		os.Exit(1)
	}
	defer body.Close()

	results, err := parser.ParseFirstThree(body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse results: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Printf("No packages found for \"%s\".\n", query)
		return
	}

	for i, res := range results {
		fmt.Printf("%d. %s\n", i+1, res.ImportPath)
		fmt.Printf("   Last Version: %s\n", res.Version)
		fmt.Printf("   Synopsis: %s\n", res.Synopsis)
	}
}
