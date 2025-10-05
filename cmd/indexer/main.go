package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/eshwanth/distributed-search-engine/internal/index"
	"github.com/eshwanth/distributed-search-engine/internal/pipeline"
	"github.com/eshwanth/distributed-search-engine/internal/semantic"
	"github.com/eshwanth/distributed-search-engine/internal/telemetry"
)

func main() {
	telemetry.RegisterMetrics()
	logger := telemetry.NewStdLogger()

	brokers := splitAndTrim(envOrDefault("KAFKA_BROKERS", "localhost:9092"))
	topic := envOrDefault("KAFKA_DOCUMENT_TOPIC", "documents")
	group := envOrDefault("INDEXER_GROUP", "indexer-service")

	consumer := pipeline.NewKafkaConsumer(brokers, topic, group, logger)
	defer consumer.Close()

	idx := index.NewInvertedIndex()
	sem := semantic.NewIndex(semantic.Options{Dimension: 128, HyperplaneCount: 24})

	snapshotPath := envOrDefault("SNAPSHOT_PATH", filepath.Join("data", "index.snapshot.json"))
	snapshotInterval := envDuration("SNAPSHOT_INTERVAL", time.Minute)

	updater := &pipeline.IndexUpdater{
		Consumer:      consumer,
		Index:         idx,
		Semantic:      sem,
		SnapshotPath:  snapshotPath,
		SnapshotEvery: snapshotInterval,
		Logger:        logger,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	metricsAddr := envOrDefault("METRICS_ADDR", ":9101")
	go func() {
		logger.Info("indexer_metrics_listening", "addr", metricsAddr)
		if err := http.ListenAndServe(metricsAddr, telemetry.MetricsHandler()); err != nil && err != http.ErrServerClosed {
			logger.Error("indexer_metrics_failed", err)
		}
	}()

	logger.Info("indexer_started", "topic", topic, "group", group, "snapshot", snapshotPath)
	if err := updater.Run(ctx); err != nil {
		logger.Error("indexer_failed", err)
	}
	logger.Info("indexer_shutdown")
}

func envOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func splitAndTrim(csv string) []string {
	parts := strings.Split(csv, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		dur, err := time.ParseDuration(val)
		if err == nil {
			return dur
		}
	}
	return fallback
}
