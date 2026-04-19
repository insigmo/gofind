package finder

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/insigmo/gofind/internal/models"
)

var (
	resultPattern      = regexp.MustCompile(`"SearchSnippet"(.*\n){35}`)
	uriPattern         = regexp.MustCompile(`<a href="/(.*)\?`)
	descriptionPattern = regexp.MustCompile(`>\n(.*)\n.*</p>`)
	versionPattern     = regexp.MustCompile(`<strong>(.*)</strong> `)
)

const searchStart = 22000

type Parser interface {
	Find(query string) (models.Result, error)
	Print(results models.Result)
}

type parser struct {
	httpClient *http.Client
}

func New(httpClient *http.Client) Parser {
	return &parser{
		httpClient: httpClient,
	}
}

func (p *parser) Find(query string) (models.Result, error) {
	body, err := p.fetchSearchResults(query)
	if err != nil {
		return models.Result{}, fmt.Errorf("error: failed to fetch results: %v", err)
	}
	defer func(body io.ReadCloser) {
		err = body.Close()
		if err != nil {
			fmt.Printf("error happened when body is closing %v", err)
			os.Exit(1)
		}
	}(body)

	results, err := p.Parse(body)
	if err != nil {
		return models.Result{}, fmt.Errorf("error: failed to parse results: %v", err)
	}

	return results, nil
}

func (p *parser) fetchSearchResults(query string) (io.ReadCloser, error) {
	const op = "finder.fetchSearchResults"

	searchURL := fmt.Sprintf("https://pkg.go.dev/search?q=%s", url.QueryEscape(query))
	req, err := http.NewRequest(http.MethodGet, searchURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("op: %s, err: %w", op, err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch results: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		err = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("error when body close: %w", err)
		}
		return nil, fmt.Errorf("err: %w, unexpected status code: %d", err, resp.StatusCode)
	}

	return resp.Body, nil
}

func (p *parser) Parse(r io.Reader) (models.Result, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return models.Result{}, fmt.Errorf("read html: %w", err)
	}

	text := resultPattern.Find([]byte(b[searchStart:]))
	descriptionData := descriptionPattern.FindAllSubmatch(text, 1)
	var description string
	if len(descriptionData) == 0 {
		description = ""
	} else {
		description = strings.TrimSpace(string(descriptionData[0][1]))
	}
	res := models.Result{
		ImportPath: string(uriPattern.FindAllSubmatch(text, 1)[0][1]),
		Synopsis:   description,
		Version:    string(versionPattern.FindAllSubmatch(text, 1)[0][1]),
	}
	return res, nil
}

func (p *parser) Print(result models.Result) {
	fmt.Println(result.ImportPath)
	fmt.Printf("  Last Version: %s\n", result.Version)

	if result.Synopsis != "" {
		fmt.Printf("  Synopsis: %s\n", result.Synopsis)
	}

	if strings.Contains(result.Version, "...") {
		fmt.Printf("  Download: go get %s\n", result.ImportPath)
	} else {
		fmt.Printf("  Download: go get %s@%s\n", result.ImportPath, result.Version)
	}
}
