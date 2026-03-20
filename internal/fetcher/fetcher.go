package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// FetchSearchResults performs a GET request to pkg.go.dev/search with the given query.
func FetchSearchResults(query string) (io.ReadCloser, error) {
	searchURL := fmt.Sprintf("https://pkg.go.dev/search?q=%s", url.QueryEscape(query))

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent as requested in the task
	req.Header.Set("User-Agent", "find-pkg/0.1 (manus-agent/go-find)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch results: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
