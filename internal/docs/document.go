package docs

import "time"

// Document represents a crawled web page that can be indexed.
type Document struct {
	ID        string
	URL       string
	Title     string
	Content   string
	Tokens    []string
	FetchedAt time.Time
}
