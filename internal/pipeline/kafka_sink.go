package pipeline

import (
	"context"
	"time"

	"github.com/eshwanth/distributed-search-engine/internal/crawler"
	"github.com/eshwanth/distributed-search-engine/internal/docs"
	"github.com/eshwanth/distributed-search-engine/internal/telemetry"
	"github.com/segmentio/kafka-go"
)

// KafkaSink implements crawler.DocumentSink by publishing JSON documents to Kafka.
type KafkaSink struct {
	writer *kafka.Writer
	logger telemetry.Logger
}

// NewKafkaSink creates a sink that writes messages to the given topic.
func NewKafkaSink(brokers []string, topic string, logger telemetry.Logger) *KafkaSink {
	return &KafkaSink{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			RequiredAcks: kafka.RequireAll,
			BatchTimeout: 10 * time.Millisecond,
		},
		logger: logger,
	}
}

// Consume publishes the document to Kafka.
func (k *KafkaSink) Consume(doc *docs.Document) {
	payload, err := MarshalDocument(doc)
	if err != nil {
		k.logger.Error("marshal_document_failed", err, "doc_id", doc.ID)
		return
	}
	msg := kafka.Message{Value: payload}
	if err := k.writer.WriteMessages(context.Background(), msg); err != nil {
		k.logger.Error("kafka_write_failed", err, "doc_id", doc.ID)
	}
}

// Close flushes and closes the underlying writer.
func (k *KafkaSink) Close() {
	if err := k.writer.Close(); err != nil {
		k.logger.Error("kafka_writer_close_failed", err)
	}
}

var _ crawler.DocumentSink = (*KafkaSink)(nil)
