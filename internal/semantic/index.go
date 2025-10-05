package semantic

import (
	"hash/fnv"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/eshwanth/distributed-search-engine/internal/docs"
	"github.com/eshwanth/distributed-search-engine/internal/index"
)

// Vector represents an embedding vector.
type Vector []float64

// dot returns the dot product of two vectors.
func dot(a, b Vector) float64 {
	length := len(a)
	if len(b) < length {
		length = len(b)
	}
	var sum float64
	for i := 0; i < length; i++ {
		sum += a[i] * b[i]
	}
	return sum
}

func magnitude(a Vector) float64 {
	var sum float64
	for _, v := range a {
		sum += v * v
	}
	return math.Sqrt(sum)
}

// cosineSimilarity returns cosine similarity between vectors.
func cosineSimilarity(a, b Vector) float64 {
	mag := magnitude(a) * magnitude(b)
	if mag == 0 {
		return 0
	}
	return dot(a, b) / mag
}

// Embedder builds semantic vectors for documents and queries.
type Embedder interface {
	EmbedTokens(tokens []string) Vector
	EmbedText(text string) Vector
}

// HashingEmbedder creates embeddings using a hashing trick for fixed-size vectors.
type HashingEmbedder struct {
	dimension int
}

// NewHashingEmbedder constructs a hashing embedder with the provided dimension.
func NewHashingEmbedder(dimension int) *HashingEmbedder {
	return &HashingEmbedder{dimension: dimension}
}

// EmbedTokens converts tokens into a hashed bag-of-words vector.
func (h *HashingEmbedder) EmbedTokens(tokens []string) Vector {
	vec := make(Vector, h.dimension)
	for _, token := range tokens {
		i := hashToken(token) % uint32(h.dimension)
		vec[i] += 1
	}
	return vec
}

// EmbedText tokenizes the text using the index tokenizer and embeds it.
func (h *HashingEmbedder) EmbedText(text string) Vector {
	tokens := index.Tokenize(text)
	return h.EmbedTokens(tokens)
}

func hashToken(token string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(token))
	return h.Sum32()
}

// Index builds an approximate nearest neighbor lookup using random hyperplanes.
type Index struct {
	mu          sync.RWMutex
	embedder    Embedder
	hyperplanes []Vector
	buckets     map[string]map[string]struct{}
	vectors     map[string]Vector
}

// Options controls semantic index construction.
type Options struct {
	Dimension       int
	HyperplaneCount int
	Seed            int64
}

// NewIndex constructs a semantic index using hashing embeddings and random projections.
func NewIndex(opts Options) *Index {
	if opts.Dimension == 0 {
		opts.Dimension = 128
	}
	if opts.HyperplaneCount == 0 {
		opts.HyperplaneCount = 24
	}
	if opts.Seed == 0 {
		opts.Seed = time.Now().UnixNano()
	}
	embedder := NewHashingEmbedder(opts.Dimension)
	hyperplanes := make([]Vector, opts.HyperplaneCount)
	rng := rand.New(rand.NewSource(opts.Seed))
	for i := range hyperplanes {
		plane := make(Vector, opts.Dimension)
		for j := range plane {
			plane[j] = rng.NormFloat64()
		}
		hyperplanes[i] = plane
	}
	return &Index{
		embedder:    embedder,
		hyperplanes: hyperplanes,
		buckets:     make(map[string]map[string]struct{}),
		vectors:     make(map[string]Vector),
	}
}

// signature computes the random projection signature for ANN bucketing.
func (i *Index) signature(vec Vector) string {
	var sb strings.Builder
	for _, plane := range i.hyperplanes {
		if dot(vec, plane) >= 0 {
			sb.WriteByte('1')
		} else {
			sb.WriteByte('0')
		}
	}
	return sb.String()
}

// AddDocument indexes the document into semantic storage.
func (i *Index) AddDocument(doc *docs.Document) {
	vec := i.embedder.EmbedText(doc.Content)
	if len(vec) == 0 {
		return
	}
	sig := i.signature(vec)
	i.mu.Lock()
	defer i.mu.Unlock()
	i.vectors[doc.ID] = vec
	bucket, ok := i.buckets[sig]
	if !ok {
		bucket = make(map[string]struct{})
		i.buckets[sig] = bucket
	}
	bucket[doc.ID] = struct{}{}
}

// Result represents a semantic retrieval candidate.
type Result struct {
	DocID string
	Score float64
}

// Query returns approximate nearest neighbors for the provided query text.
func (i *Index) Query(query string, topK int) []Result {
	vec := i.embedder.EmbedText(query)
	if len(vec) == 0 {
		return nil
	}
	sig := i.signature(vec)
	i.mu.RLock()
	candidates := make([]string, 0)
	for docID := range i.buckets[sig] {
		candidates = append(candidates, docID)
	}
	// Fallback to limited global scan if bucket is sparse.
	if len(candidates) < topK {
		for docID := range i.vectors {
			candidates = append(candidates, docID)
			if len(candidates) >= 5*topK && topK > 0 {
				break
			}
		}
	}
	seen := make(map[string]struct{}, len(candidates))
	unique := make([]string, 0, len(candidates))
	for _, id := range candidates {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	scores := make([]Result, 0, len(unique))
	for _, id := range unique {
		vecDoc := i.vectors[id]
		if vecDoc == nil {
			continue
		}
		scores = append(scores, Result{DocID: id, Score: cosineSimilarity(vec, vecDoc)})
	}
	i.mu.RUnlock()

	sort.Slice(scores, func(a, b int) bool {
		if scores[a].Score == scores[b].Score {
			return scores[a].DocID < scores[b].DocID
		}
		return scores[a].Score > scores[b].Score
	})

	if topK > 0 && len(scores) > topK {
		scores = scores[:topK]
	}
	return scores
}

// DocumentVector returns the stored vector for a given document if present.
func (i *Index) DocumentVector(id string) (Vector, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	vec, ok := i.vectors[id]
	return vec, ok
}

// Reset clears the semantic index (useful for tests).
func (i *Index) Reset() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.vectors = make(map[string]Vector)
	i.buckets = make(map[string]map[string]struct{})
}
