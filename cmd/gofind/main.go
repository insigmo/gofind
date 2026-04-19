// cli for finding packages

package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/insigmo/gofind/internal/services/finder"
)

const timeout = 15 * time.Second

func main() {
	query := getPackageName()
	client := &http.Client{
		Timeout: timeout,
	}
	queryFinder := finder.New(client)

	results, err := queryFinder.Find(query)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	queryFinder.Print(results)
}

func getPackageName() string {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Usage: gofind <query>")
		os.Exit(1)
	}
	return args[0]
}
