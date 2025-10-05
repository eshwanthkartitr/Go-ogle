package search_test

import (
	"testing"

	"github.com/eshwanth/distributed-search-engine/internal/docs"
	"github.com/eshwanth/distributed-search-engine/internal/index"
	"github.com/eshwanth/distributed-search-engine/internal/search"
	"github.com/eshwanth/distributed-search-engine/internal/semantic"
)

func TestSearchRanksRelevantDocumentHigher(t *testing.T) {
	idx := index.NewInvertedIndex()
	idx.AddDocument(&docs.Document{ID: "1", Title: "Vector Search", Content: "Vector search uses embeddings and approximate nearest neighbors."})
	idx.AddDocument(&docs.Document{ID: "2", Title: "Circuit Breakers", Content: "Circuit breakers protect distributed systems from cascading failures."})

	svc := search.NewService(idx, nil)
	results := svc.Search("vector search", 5)

	if len(results) == 0 {
		t.Fatalf("expected results, got none")
	}

	if results[0].DocID != "1" {
		t.Fatalf("expected doc1 to rank first, got %s", results[0].DocID)
	}
}

func TestSemanticBoostsNonLexicalDocument(t *testing.T) {
	idx := index.NewInvertedIndex()
	idx.AddDocument(&docs.Document{ID: "lex", Title: "Keyword Match", Content: "Classical keyword search"})
	idx.AddDocument(&docs.Document{ID: "sem", Title: "Embeddings", Content: "Dense vector representations for semantic retrieval"})

	semIdx := semantic.NewIndex(semantic.Options{Dimension: 64, HyperplaneCount: 16, Seed: 42})
	semIdx.AddDocument(&docs.Document{ID: "lex", Content: "Classical keyword search"})
	semIdx.AddDocument(&docs.Document{ID: "sem", Content: "Dense vector representations for semantic retrieval"})

	svc := search.NewService(idx, semIdx)
	svc.LexicalWeight = 0.2
	svc.SemanticWeight = 1.0

	results := svc.Search("dense retrieval", 5)
	if len(results) == 0 {
		t.Fatalf("expected results, got none")
	}
	if results[0].DocID != "sem" {
		t.Fatalf("expected semantic document to rank first, got %s", results[0].DocID)
	}
}
