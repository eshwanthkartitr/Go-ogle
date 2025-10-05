package index

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/eshwanth/distributed-search-engine/internal/docs"
)

// Snapshot contains a serializable view of the inverted index.
type Snapshot struct {
	Documents []*SnapshotDocument `json:"documents"`
}

// SnapshotDocument is a lightweight representation persisted to disk.
type SnapshotDocument struct {
	ID      string   `json:"id"`
	URL     string   `json:"url"`
	Title   string   `json:"title"`
	Tokens  []string `json:"tokens"`
	Content string   `json:"content"`
}

// WriteSnapshot exports the index and documents to the provided path.
func WriteSnapshot(idx *InvertedIndex, path string) error {
	docs := idx.Documents()
	snapshot := Snapshot{Documents: make([]*SnapshotDocument, 0, len(docs))}
	for _, doc := range docs {
		snapshot.Documents = append(snapshot.Documents, &SnapshotDocument{
			ID:      doc.ID,
			URL:     doc.URL,
			Title:   doc.Title,
			Tokens:  doc.Tokens,
			Content: doc.Content,
		})
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(&snapshot)
}

// LoadSnapshot reads documents from disk allowing callers to hydrate indexes.
func LoadSnapshot(path string) ([]*docs.Document, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var snapshot Snapshot
	if err := json.NewDecoder(file).Decode(&snapshot); err != nil {
		return nil, err
	}
	docsOut := make([]*docs.Document, 0, len(snapshot.Documents))
	for _, entry := range snapshot.Documents {
		docsOut = append(docsOut, &docs.Document{
			ID:      entry.ID,
			URL:     entry.URL,
			Title:   entry.Title,
			Tokens:  entry.Tokens,
			Content: entry.Content,
		})
	}
	return docsOut, nil
}
