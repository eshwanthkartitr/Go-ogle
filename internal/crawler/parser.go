package crawler

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/eshwanth/distributed-search-engine/internal/docs"
	"golang.org/x/net/html"
)

// Parser extracts structured data and discovered links from raw HTML.
type Parser interface {
	Parse(baseURL string, htmlBody string) (*docs.Document, []string, error)
}

// HTMLParser parses HTML documents extracting their title, textual content, and hyperlinks.
type HTMLParser struct{}

// Parse returns a Document alongside discovered links.
func (p *HTMLParser) Parse(baseURL string, htmlBody string) (*docs.Document, []string, error) {
	node, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return nil, nil, err
	}

	title := extractTitle(node)
	text := extractText(node)
	links := extractLinks(node, baseURL)

	sum := sha1.Sum([]byte(baseURL))
	id := fmt.Sprintf("%x", sum[:])

	doc := &docs.Document{
		ID:      id,
		URL:     baseURL,
		Title:   title,
		Content: text,
	}

	return doc, links, nil
}

func extractTitle(node *html.Node) string {
	if node.Type == html.ElementNode && node.Data == "title" && node.FirstChild != nil {
		return strings.TrimSpace(node.FirstChild.Data)
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if title := extractTitle(child); title != "" {
			return title
		}
	}
	return ""
}

func extractText(node *html.Node) string {
	var buf bytes.Buffer
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				buf.WriteString(text)
				buf.WriteByte(' ')
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			traverse(child)
		}
	}
	traverse(node)
	return buf.String()
}

func extractLinks(node *html.Node, base string) []string {
	var links []string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					links = append(links, resolveLink(base, attr.Val))
					break
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			traverse(child)
		}
	}
	traverse(node)
	return links
}

func resolveLink(base string, href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") || strings.HasPrefix(href, "file://") {
		return href
	}

	return joinPath(base, href)
}

func joinPath(base string, link string) string {
	if base == "" {
		return link
	}
	if strings.HasPrefix(base, "file://") {
		index := strings.LastIndex(base, "/")
		if index == -1 {
			return base + "/" + link
		}
		return base[:index+1] + link
	}
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(link, "/")
}
