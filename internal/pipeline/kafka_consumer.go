package pipeline

import (
	"context"
	"errors"
	"time"

	"github.com/eshwanth/distributed-search-engine/internal/docs"
	"github.com/eshwanth/distributed-search-engine/internal/telemetry"
	"github.com/segmentio/kafka-go"
)

// DocumentHandler processes a document delivered from the transport layer.
type DocumentHandler func(*docs.Document) error

// DocumentConsumer represents a streaming consumer for indexed documents.
type DocumentConsumer interface {
	Consume(ctx context.Context, handler DocumentHandler) error
	Close() error
}

// KafkaConsumer consumes documents from a Kafka topic.
type KafkaConsumer struct {
	reader *kafka.Reader
	logger telemetry.Logger
}

// NewKafkaConsumer creates a new Kafka consumer bound to the provided topic and group.
func NewKafkaConsumer(brokers []string, topic, groupID string, logger telemetry.Logger) *KafkaConsumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			Topic:          topic,
			GroupID:        groupID,
			CommitInterval: time.Second,
		}),
		logger: logger,
	}
}

// Consume continuously reads messages and invokes the handler until the context is canceled or an error occurs.
func (k *KafkaConsumer) Consume(ctx context.Context, handler DocumentHandler) error {
	for {
		m, err := k.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
		doc, err := UnmarshalDocument(m.Value)
		if err != nil {
			k.logger.Error("kafka_unmarshal_failed", err)
			continue
		}
		if err := handler(doc); err != nil {
			k.logger.Error("document_handler_failed", err, "doc_id", doc.ID)
		}
	}
}

// Close closes the reader.
func (k *KafkaConsumer) Close() error {
	return k.reader.Close()
}

var _ DocumentConsumer = (*KafkaConsumer)(nil)
