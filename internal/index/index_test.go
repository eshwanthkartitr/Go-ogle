package index_test

import (
	"testing"

	"github.com/eshwanth/distributed-search-engine/internal/docs"
	"github.com/eshwanth/distributed-search-engine/internal/index"
)

func TestAddDocumentAndPostings(t *testing.T) {
	idx := index.NewInvertedIndex()
	doc := &docs.Document{ID: "doc1", Content: "Distributed systems need consistency and availability."}
	idx.AddDocument(doc)

	if got := idx.DocumentCount(); got != 1 {
		t.Fatalf("expected document count 1, got %d", got)
	}

	postings := idx.Postings("distributed")
	if len(postings) != 1 {
		t.Fatalf("expected postings length 1, got %d", len(postings))
	}

	if postings[0].DocID != "doc1" {
		t.Fatalf("expected DocID doc1, got %s", postings[0].DocID)
	}
}

func TestTokenize(t *testing.T) {
	tokens := index.Tokenize("Tail latency hurts Search.")
	expected := []string{"tail", "latency", "hurts", "search"}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, token := range tokens {
		if token != expected[i] {
			t.Fatalf("expected token %s at position %d, got %s", expected[i], i, token)
		}
	}
}
