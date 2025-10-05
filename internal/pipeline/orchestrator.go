package pipeline

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/eshwanth/distributed-search-engine/internal/crawler"
	"github.com/eshwanth/distributed-search-engine/internal/telemetry"
)

// Orchestrator glues together the crawler and indexing pipeline.
type Orchestrator struct {
	Crawler *crawler.Crawler
	Sink    crawler.DocumentSink
	Logger  telemetry.Logger
}

// Run triggers a crawl for the provided seeds.
func (o *Orchestrator) Run(ctx context.Context, seeds []string) {
	start := time.Now()
	o.Logger.Info("pipeline_start", "seeds", len(seeds))
	o.Crawler.Crawl(ctx, seeds, o.Sink)
	o.Logger.Info("pipeline_complete", "duration_ms", time.Since(start).Milliseconds())
}

// LocalSeeds converts relative fixture names into file URLs under the provided directory.
func LocalSeeds(base string, filenames ...string) []string {
	seeds := make([]string, 0, len(filenames))
	for _, name := range filenames {
		abs, err := filepath.Abs(filepath.Join(base, name))
		if err != nil {
			continue
		}
		seeds = append(seeds, fmt.Sprintf("file://%s", abs))
	}
	return seeds
}

// NewCrawlerOrchestrator builds a crawler orchestrator with the provided sink.
func NewCrawlerOrchestrator(logger telemetry.Logger, sink crawler.DocumentSink) *Orchestrator {
	fetcher := &crawler.HTTPFetcher{}
	parser := &crawler.HTMLParser{}
	c := crawler.New(fetcher, parser, logger)
	c.Workers = 6
	c.MaxPages = 1000
	return &Orchestrator{Crawler: c, Sink: sink, Logger: logger}
}
