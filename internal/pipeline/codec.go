package pipeline

import (
	"encoding/json"

	"github.com/eshwanth/distributed-search-engine/internal/docs"
)

// MarshalDocument encodes a document into JSON for transport.
func MarshalDocument(doc *docs.Document) ([]byte, error) {
	return json.Marshal(doc)
}

// UnmarshalDocument decodes a document payload back into a Document.
func UnmarshalDocument(payload []byte) (*docs.Document, error) {
	var doc docs.Document
	if err := json.Unmarshal(payload, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}
