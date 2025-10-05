package crawler

import "testing"

func TestHTMLParserExtractsTitleAndLinks(t *testing.T) {
	const html = `<!DOCTYPE html><html><head><title>Sample</title></head><body><a href="next.html">Next</a></body></html>`
	parser := &HTMLParser{}
	doc, links, err := parser.Parse("file:///tmp/sample.html", html)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Title != "Sample" {
		t.Fatalf("expected title Sample, got %s", doc.Title)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0] != "file:///tmp/next.html" {
		t.Fatalf("expected resolved link, got %s", links[0])
	}
}
