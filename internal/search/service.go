package search

import (
	"math"
	"sort"
	"strings"

	"github.com/eshwanth/distributed-search-engine/internal/index"
	"github.com/eshwanth/distributed-search-engine/internal/semantic"
)

// Result represents a ranked document for a query.
type Result struct {
	DocID   string  `json:"doc_id"`
	Score   float64 `json:"score"`
	Title   string  `json:"title"`
	Snippet string  `json:"snippet"`
	URL     string  `json:"url"`
}

// Service executes ranked search queries against the inverted index.
type Service struct {
	Index          *index.InvertedIndex
	Semantic       *semantic.Index
	K1             float64
	B              float64
	LexicalWeight  float64
	SemanticWeight float64
}

// NewService creates a search Service with BM25 defaults.
func NewService(idx *index.InvertedIndex, semanticIdx *semantic.Index) *Service {
	return &Service{
		Index:          idx,
		Semantic:       semanticIdx,
		K1:             1.5,
		B:              0.75,
		LexicalWeight:  1.0,
		SemanticWeight: 0.65,
	}
}

// Search tokenizes the query, scores documents, and returns the topK results.
func (s *Service) Search(query string, topK int) []Result {
	if s.Index == nil {
		return nil
	}
	tokens := index.Tokenize(query)
	if len(tokens) == 0 {
		return nil
	}

	docCount := float64(s.Index.DocumentCount())
	avgDocLen := s.Index.AverageDocumentLength()
	if avgDocLen == 0 {
		avgDocLen = 1
	}

	lexicalScores := make(map[string]float64)
	for _, term := range tokens {
		postings := s.Index.Postings(term)
		if len(postings) == 0 {
			continue
		}
		df := float64(s.Index.DocumentFrequency(term))
		if df == 0 {
			continue
		}
		idf := math.Log((docCount - df + 0.5) / (df + 0.5))
		if idf < 0 {
			idf = 0
		}
		for _, posting := range postings {
			doc, ok := s.Index.Document(posting.DocID)
			if !ok {
				continue
			}
			docLen := float64(len(doc.Tokens))
			if docLen == 0 {
				docLen = avgDocLen
			}
			numerator := posting.TF * (s.K1 + 1)
			denominator := posting.TF + s.K1*(1-s.B+s.B*(docLen/avgDocLen))
			score := idf * (numerator / denominator)
			lexicalScores[posting.DocID] += score
		}
	}

	semanticScores := make(map[string]float64)
	if s.Semantic != nil {
		semanticLimit := topK
		if semanticLimit < 10 {
			semanticLimit = 10
		}
		candidates := s.Semantic.Query(query, semanticLimit)
		for _, candidate := range candidates {
			semanticScores[candidate.DocID] = candidate.Score
		}
	}

	combined := make(map[string]float64)
	for docID, lexical := range lexicalScores {
		combined[docID] += s.LexicalWeight * lexical
	}
	for docID, sem := range semanticScores {
		combined[docID] += s.SemanticWeight * sem
	}

	results := make([]Result, 0, len(combined))
	for docID, score := range combined {
		doc, ok := s.Index.Document(docID)
		if !ok {
			continue
		}
		snippet := buildSnippet(doc.Content, tokens)
		results = append(results, Result{DocID: docID, Score: score, Title: doc.Title, URL: doc.URL, Snippet: snippet})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].DocID < results[j].DocID
		}
		return results[i].Score > results[j].Score
	})

	if topK > 0 && len(results) > topK {
		results = results[:topK]
	}
	return results
}

func buildSnippet(content string, tokens []string) string {
	if len(content) == 0 {
		return ""
	}

	lower := strings.ToLower(content)
	for _, token := range tokens {
		pos := strings.Index(lower, token)
		if pos >= 0 {
			start := pos - 40
			if start < 0 {
				start = 0
			}
			end := pos + 40
			if end > len(content) {
				end = len(content)
			}
			return content[start:end]
		}
	}
	if len(content) > 80 {
		return content[:80]
	}
	return content
}
