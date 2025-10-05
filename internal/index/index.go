package index

import (
	"sort"
	"sync"

	"github.com/eshwanth/distributed-search-engine/internal/docs"
)

// Posting represents a term occurrence in a document.
type Posting struct {
	DocID string
	TF    float64
}

// InvertedIndex stores postings lists and document statistics.
type InvertedIndex struct {
	mu          sync.RWMutex
	documents   map[string]*docs.Document
	postings    map[string]map[string]*Posting
	docLengths  map[string]int
	docTerms    map[string][]string
	totalTokens int
}

// NewInvertedIndex constructs an empty index.
func NewInvertedIndex() *InvertedIndex {
	return &InvertedIndex{
		documents:  make(map[string]*docs.Document),
		postings:   make(map[string]map[string]*Posting),
		docLengths: make(map[string]int),
		docTerms:   make(map[string][]string),
	}
}

// AddDocument tokenizes and inserts the document into the index.
func (idx *InvertedIndex) AddDocument(doc *docs.Document) {
	tokens := doc.Tokens
	if len(tokens) == 0 {
		tokens = Tokenize(doc.Content)
		doc.Tokens = tokens
	}

	termCounts := make(map[string]int)
	for _, token := range tokens {
		termCounts[token]++
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	if existingTerms, ok := idx.docTerms[doc.ID]; ok {
		for _, term := range existingTerms {
			if postingList, ok := idx.postings[term]; ok {
				delete(postingList, doc.ID)
				if len(postingList) == 0 {
					delete(idx.postings, term)
				}
			}
		}
		if length, ok := idx.docLengths[doc.ID]; ok {
			idx.totalTokens -= length
		}
	}

	totalTerms := 0
	termsOrdered := make([]string, 0, len(termCounts))
	for term, count := range termCounts {
		totalTerms += count
		termsOrdered = append(termsOrdered, term)
		postingList, ok := idx.postings[term]
		if !ok {
			postingList = make(map[string]*Posting)
			idx.postings[term] = postingList
		}
		postingList[doc.ID] = &Posting{DocID: doc.ID, TF: float64(count)}
	}

	idx.documents[doc.ID] = doc
	idx.docLengths[doc.ID] = len(tokens)
	idx.docTerms[doc.ID] = termsOrdered
	idx.totalTokens += totalTerms
}

// Document retrieves a stored document by ID.
func (idx *InvertedIndex) Document(id string) (*docs.Document, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	d, ok := idx.documents[id]
	return d, ok
}

// Postings returns a copy of the postings list for a term.
func (idx *InvertedIndex) Postings(term string) []Posting {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	entry, ok := idx.postings[term]
	if !ok {
		return nil
	}
	results := make([]Posting, 0, len(entry))
	for _, posting := range entry {
		results = append(results, *posting)
	}
	return results
}
func (idx *InvertedIndex) DocumentFrequency(term string) int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.postings[term])
}

// DocumentCount returns the number of indexed documents.
func (idx *InvertedIndex) DocumentCount() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.documents)
}

// AverageDocumentLength returns the average token length across indexed documents.
func (idx *InvertedIndex) AverageDocumentLength() float64 {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	if len(idx.docLengths) == 0 {
		return 0
	}
	return float64(idx.totalTokens) / float64(len(idx.docLengths))
}

// Documents returns all documents sorted by ID for deterministic ordering.
func (idx *InvertedIndex) Documents() []*docs.Document {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	ids := make([]string, 0, len(idx.documents))
	for id := range idx.documents {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]*docs.Document, 0, len(ids))
	for _, id := range ids {
		result = append(result, idx.documents[id])
	}
	return result
}
