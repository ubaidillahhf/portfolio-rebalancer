package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// Publisher implements messaging.Publisher interface using Kafka
type Publisher struct {
	writer *kafka.Writer
}

// NewPublisher creates a new Kafka publisher
func NewPublisher(writer *kafka.Writer) *Publisher {
	return &Publisher{
		writer: writer,
	}
}

// Publish publishes a message to Kafka
func (p *Publisher) Publish(ctx context.Context, message []byte) error {
	if p.writer == nil {
		return nil // Graceful degradation if Kafka not available
	}

	msg := kafka.Message{
		Value: message,
	}

	return p.writer.WriteMessages(ctx, msg)
}
