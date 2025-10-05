package pipeline

import (
	"context"
	"time"

	"github.com/eshwanth/distributed-search-engine/internal/docs"
	"github.com/eshwanth/distributed-search-engine/internal/index"
	"github.com/eshwanth/distributed-search-engine/internal/semantic"
	"github.com/eshwanth/distributed-search-engine/internal/telemetry"
)

// IndexUpdater consumes documents and updates both lexical and semantic indexes.
type IndexUpdater struct {
	Consumer        DocumentConsumer
	Index           *index.InvertedIndex
	Semantic        *semantic.Index
	SnapshotPath    string
	SnapshotEvery   time.Duration
	Logger          telemetry.Logger
	lastSnapshotRun time.Time
}

// Run starts consuming documents until context cancellation or consumer error.
func (u *IndexUpdater) Run(ctx context.Context) error {
	if u.Consumer == nil {
		return nil
	}
	return u.Consumer.Consume(ctx, func(doc *docs.Document) error {
		u.Index.AddDocument(doc)
		if u.Semantic != nil {
			u.Semantic.AddDocument(doc)
		}
		telemetry.IncIndexUpdates()

		if u.SnapshotPath != "" && u.SnapshotEvery > 0 {
			now := time.Now()
			if u.lastSnapshotRun.IsZero() || now.Sub(u.lastSnapshotRun) >= u.SnapshotEvery {
				if err := index.WriteSnapshot(u.Index, u.SnapshotPath); err != nil {
					u.Logger.Error("write_snapshot_failed", err, "path", u.SnapshotPath)
				} else {
					u.Logger.Info("snapshot_written", "path", u.SnapshotPath)
					u.lastSnapshotRun = now
				}
			}
		}
		return nil
	})
}
