package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/eshwanth/distributed-search-engine/internal/api"
	"github.com/eshwanth/distributed-search-engine/internal/index"
	"github.com/eshwanth/distributed-search-engine/internal/pipeline"
	"github.com/eshwanth/distributed-search-engine/internal/search"
	"github.com/eshwanth/distributed-search-engine/internal/semantic"
	"github.com/eshwanth/distributed-search-engine/internal/telemetry"
)

func main() {
	telemetry.RegisterMetrics()
	logger := telemetry.NewStdLogger()

	brokers := splitAndTrim(envOrDefault("KAFKA_BROKERS", "localhost:9092"))
	topic := envOrDefault("KAFKA_DOCUMENT_TOPIC", "documents")
	group := envOrDefault("SEARCH_GROUP", "search-api")

	idx := index.NewInvertedIndex()
	sem := semantic.NewIndex(semantic.Options{Dimension: 128, HyperplaneCount: 24})

	snapshotPath := envOrDefault("SNAPSHOT_PATH", filepath.Join("data", "index.snapshot.json"))
	if info, err := os.Stat(snapshotPath); err == nil && !info.IsDir() {
		if docs, err := index.LoadSnapshot(snapshotPath); err == nil {
			for _, doc := range docs {
				idx.AddDocument(doc)
				sem.AddDocument(doc)
			}
			logger.Info("snapshot_loaded", "path", snapshotPath, "documents", len(docs))
		} else {
			logger.Error("snapshot_load_failed", err, "path", snapshotPath)
		}
	}

	consumer := pipeline.NewKafkaConsumer(brokers, topic, group, logger)
	defer consumer.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	updater := &pipeline.IndexUpdater{
		Consumer:      consumer,
		Index:         idx,
		Semantic:      sem,
		SnapshotPath:  "",
		SnapshotEvery: 0,
		Logger:        logger,
	}

	go func() {
		if err := updater.Run(ctx); err != nil {
			logger.Error("search_index_updater_failed", err)
		}
	}()

	service := search.NewService(idx, sem)
	server := &api.Server{Search: service, Logger: logger}

	addr := envOrDefault("SEARCH_HTTP_ADDR", ":8080")
	metricsAddr := envOrDefault("METRICS_ADDR", ":9102")
	go func() {
		logger.Info("metrics_listening", "addr", metricsAddr)
		if err := http.ListenAndServe(metricsAddr, telemetry.MetricsHandler()); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics_server_failed", err)
		}
	}()

	logger.Info("search_api_listening", "addr", addr, "topic", topic, "group", group)
	if err := server.Start(addr); err != nil {
		logger.Error("api_shutdown", err)
	}
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
