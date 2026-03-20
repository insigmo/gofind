package parser

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// Result represents a search result from pkg.go.dev.
type Result struct {
	ImportPath string
	Version    string
	Synopsis   string
}

// ParseFirstThree parses the HTML from the reader and returns the first three results.
func ParseFirstThree(reader io.Reader) ([]Result, error) {
	doc, err := html.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var results []Result
	var findSnippets func(*html.Node)
	findSnippets = func(n *html.Node) {
		if len(results) >= 3 {
			return
		}

		if n.Type == html.ElementNode && n.Data == "div" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "SearchSnippet") {
					res := scrapePackageFromNode(n)
					if res != nil {
						results = append(results, *res)
					}
					return // Don't look deeper into this snippet
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findSnippets(c)
		}
	}

	findSnippets(doc)
	return results, nil
}

func scrapePackageFromNode(n *html.Node) *Result {
	res := &Result{}

	// Find the link in h2
	h2 := findNode(n, func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.Data == "h2"
	})

	if h2 != nil {
		a := findNode(h2, func(node *html.Node) bool {
			return node.Type == html.ElementNode && node.Data == "a"
		})
		if a != nil {
			for _, attr := range a.Attr {
				if attr.Key == "href" {
					res.ImportPath = strings.TrimPrefix(attr.Val, "/")
				}
			}
		}
	}

	// Find synopsis
	synopsisNode := findNode(n, func(node *html.Node) bool {
		if node.Type == html.ElementNode && node.Data == "p" {
			for _, attr := range node.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "SearchSnippet-synopsis") {
					return true
				}
			}
		}
		return false
	})
	if synopsisNode != nil {
		res.Synopsis = strings.TrimSpace(getText(synopsisNode))
	}

	// Find version in SearchSnippet-infoLabel
	infoLabel := findNode(n, func(node *html.Node) bool {
		if node.Type == html.ElementNode && node.Data == "div" {
			for _, attr := range node.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "SearchSnippet-infoLabel") {
					return true
				}
			}
		}
		return false
	})
	if infoLabel != nil {
		text := getText(infoLabel)
		// The text usually looks like "Imported by 1,527 | v1.5.0 published on Jul 20, 2023 | MIT"
		parts := strings.Split(text, "|")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if strings.HasPrefix(trimmed, "v") {
				// Take the first word (version)
				versionParts := strings.Fields(trimmed)
				if len(versionParts) > 0 {
					res.Version = versionParts[0]
					break
				}
			}
		}
	}

	if res.ImportPath == "" {
		return nil
	}

	return res
}

func getText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		b.WriteString(getText(c))
	}
	return b.String()
}

func findNode(n *html.Node, match func(*html.Node) bool) *html.Node {
	if match(n) {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if res := findNode(c, match); res != nil {
			return res
		}
	}
	return nil
}
