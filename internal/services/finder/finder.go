package finder

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/insigmo/find/internal/models"
)

var (
	resultRe = regexp.MustCompile(
		`(?s)<div[^>]*class="[^"]*\bSearchSnippet\b[^"]*"[^>]*>.*?<h2[^>]*>\s*<a[^>]*href="/([^"]+)"[^>]*>.*?</a>\s*</h2>.*?<p[^>]*class="[^"]*\bSearchSnippet-synopsis\b[^"]*"[^>]*>(.*?)</p>.*?<div[^>]*class="[^"]*\bSearchSnippet-infoLabel\b[^"]*"[^>]*>(.*?)</div>`,
	)
	tag = regexp.MustCompile(`<[^>]+>`)
	ver = regexp.MustCompile(`\bv[^\s|]+`)
)

type Parser interface {
	Find(query string) ([]models.Result, error)
	Print(results []models.Result)
}

type parser struct {
	httpClient *http.Client
}

func New(httpClient *http.Client) Parser {
	return &parser{
		httpClient: httpClient,
	}
}

func (p *parser) Find(query string) ([]models.Result, error) {
	body, err := p.fetchSearchResults(query)
	if err != nil {
		return nil, fmt.Errorf("error: failed to fetch results: %v", err)
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
		return nil, fmt.Errorf("error: failed to parse results: %v", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no packages found for \"%s\"", query)
	}
	return results, nil
}

func (p *parser) Print(results []models.Result) {
	for i, res := range results {
		fmt.Printf("%d. %s\n", i+1, res.ImportPath)
		fmt.Printf("\tLast Version: %s\n", res.Version)
		fmt.Printf("\tSynopsis: %s\n", res.Synopsis)
	}
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

func (p *parser) Parse(r io.Reader) ([]models.Result, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read html: %w", err)
	}
	ms := resultRe.FindAllStringSubmatch(string(b), 1)
	out := make([]models.Result, 0, len(ms))
	for _, m := range ms {
		out = append(out, models.Result{
			ImportPath: m[1],
			Synopsis:   strings.TrimSpace(html.UnescapeString(tag.ReplaceAllString(m[2], ""))),
			Version:    ver.FindString(m[3]),
		})
	}
	return out, nil
}
