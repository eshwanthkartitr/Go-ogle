package crawler

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/eshwanth/distributed-search-engine/internal/docs"
	"github.com/eshwanth/distributed-search-engine/internal/telemetry"
)

// DocumentSink receives parsed documents for downstream indexing.
type DocumentSink interface {
	Consume(doc *docs.Document)
	Close()
}

// Crawler walks the web graph starting from a seed frontier.
type Crawler struct {
	Fetcher      Fetcher
	Parser       Parser
	Logger       telemetry.Logger
	Workers      int
	Politeness   time.Duration
	MaxPages     int
	visited      sync.Map
	visitedCount atomic.Int64
}

// New creates a crawler with sane defaults.
func New(fetcher Fetcher, parser Parser, logger telemetry.Logger) *Crawler {
	return &Crawler{
		Fetcher:    fetcher,
		Parser:     parser,
		Logger:     logger,
		Workers:    4,
		Politeness: 50 * time.Millisecond,
		MaxPages:   100,
	}
}

// Crawl starts concurrent workers that fetch and parse URLs and stream documents to the sink.
func (c *Crawler) Crawl(ctx context.Context, seeds []string, sink DocumentSink) {
	defer sink.Close()

	workCh := make(chan string, c.Workers*2)
	var workerWG sync.WaitGroup
	var pending sync.WaitGroup

	enqueue := func(url string) {
		if url == "" {
			return
		}
		if c.MaxPages > 0 && c.visitedCount.Load() >= int64(c.MaxPages) {
			return
		}
		if _, seen := c.visited.LoadOrStore(url, struct{}{}); seen {
			return
		}
		if c.MaxPages > 0 && c.visitedCount.Add(1) > int64(c.MaxPages) {
			return
		}
		pending.Add(1)
		select {
		case workCh <- url:
		case <-ctx.Done():
			pending.Done()
		}
	}

	for i := 0; i < c.Workers; i++ {
		workerWG.Add(1)
		go func(id int) {
			defer workerWG.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case url, ok := <-workCh:
					if !ok {
						return
					}
					c.handleURL(ctx, url, sink, enqueue)
					pending.Done()
					if c.Politeness > 0 {
						timer := time.NewTimer(c.Politeness)
						select {
						case <-ctx.Done():
							timer.Stop()
							return
						case <-timer.C:
						}
					}
				}
			}
		}(i)
	}

	for _, seed := range seeds {
		enqueue(seed)
	}

	done := make(chan struct{})
	go func() {
		pending.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
	case <-done:
	}

	close(workCh)
	workerWG.Wait()
}

func (c *Crawler) handleURL(ctx context.Context, target string, sink DocumentSink, enqueue func(string)) {
	body, err := c.Fetcher.Fetch(target)
	if err != nil {
		c.Logger.Error("fetch failed", err, "url", target)
		telemetry.IncCrawlerErrors()
		return
	}

	doc, links, err := c.Parser.Parse(target, body)
	if err != nil {
		c.Logger.Error("parse failed", err, "url", target)
		telemetry.IncCrawlerErrors()
		return
	}

	sink.Consume(doc)
	c.Logger.Info("crawled", "url", target, "out_links", len(links))
	telemetry.IncCrawlerDocuments()

	for _, link := range links {
		select {
		case <-ctx.Done():
			return
		default:
		}
		enqueue(link)
	}
}
