package pipeline

import (
	"sync"

	"github.com/eshwanth/distributed-search-engine/internal/crawler"
	"github.com/eshwanth/distributed-search-engine/internal/docs"
	"github.com/eshwanth/distributed-search-engine/internal/index"
)

// IndexSink writes documents from the crawler into the inverted index.
type IndexSink struct {
	idx *index.InvertedIndex
	mu  sync.Mutex
}

// NewIndexSink creates a sink bound to the provided index.
func NewIndexSink(idx *index.InvertedIndex) *IndexSink {
	return &IndexSink{idx: idx}
}

// Consume indexes the received document.
func (s *IndexSink) Consume(doc *docs.Document) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.idx.AddDocument(doc)
}

// Close implements crawler.DocumentSink.
func (s *IndexSink) Close() {}

var _ crawler.DocumentSink = (*IndexSink)(nil)
