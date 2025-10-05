package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/eshwanth/distributed-search-engine/internal/pipeline"
	"github.com/eshwanth/distributed-search-engine/internal/telemetry"
)

func main() {
	telemetry.RegisterMetrics()
	logger := telemetry.NewStdLogger()

	brokersEnv := envOrDefault("KAFKA_BROKERS", "localhost:9092")
	brokers := splitAndTrim(brokersEnv)
	topic := envOrDefault("KAFKA_DOCUMENT_TOPIC", "documents")

	sink := pipeline.NewKafkaSink(brokers, topic, logger)
	defer sink.Close()

	orch := pipeline.NewCrawlerOrchestrator(logger, sink)

	seedDir := envOrDefault("SEED_DIR", filepath.Join("testdata", "pages"))
	seedFiles := envOrDefault("SEED_FILES", "distributed-systems.html,resilient-search.html,ranking-ml.html,vector-search.html")
	seeds := pipeline.LocalSeeds(seedDir, splitAndTrim(seedFiles)...)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	orch.Run(ctx, seeds)
	logger.Info("crawler_complete", "seeds", len(seeds), "topic", topic, "brokers", brokersEnv)
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
