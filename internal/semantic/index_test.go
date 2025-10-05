package semantic_test

import (
	"testing"

	"github.com/eshwanth/distributed-search-engine/internal/docs"
	"github.com/eshwanth/distributed-search-engine/internal/semantic"
)

func TestSemanticIndexQueryReturnsNearestDocument(t *testing.T) {
	idx := semantic.NewIndex(semantic.Options{Dimension: 64, HyperplaneCount: 16, Seed: 99})
	idx.AddDocument(&docs.Document{ID: "1", Content: "Vector embeddings enable semantic search."})
	idx.AddDocument(&docs.Document{ID: "2", Content: "Caching strategies reduce tail latency."})

	results := idx.Query("semantic embeddings", 2)
	if len(results) == 0 {
		t.Fatalf("expected results, got none")
	}
	if results[0].DocID != "1" {
		t.Fatalf("expected embedding document to rank first, got %s", results[0].DocID)
	}
}
