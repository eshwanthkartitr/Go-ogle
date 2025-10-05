package pipeline_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/eshwanth/distributed-search-engine/internal/pipeline"
)

func TestLocalSeeds(t *testing.T) {
	base := filepath.Join("testdata", "pages")
	seeds := pipeline.LocalSeeds(base, "distributed-systems.html")
	if len(seeds) != 1 {
		t.Fatalf("expected 1 seed, got %d", len(seeds))
	}
	if !strings.HasPrefix(seeds[0], "file://") {
		t.Fatalf("expected file scheme, got %s", seeds[0])
	}
}
