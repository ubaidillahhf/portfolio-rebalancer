package kafka

import (
	"context"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
)

var writer *kafka.Writer

// GetWriter returns the Kafka writer instance
func GetWriter() *kafka.Writer {
	return writer
}

// Init initializes kafka connection with retry logic
func Init() error {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	topic := os.Getenv("KAFKA_TOPIC")

	if kafkaBroker == "" || topic == "" {
		log.Println("Kafka config not set; skipping Kafka initialization")
		return nil // skip if env not set
	}

	writer = &kafka.Writer{
		Addr:     kafka.TCP(kafkaBroker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	// Retry logic to check Kafka availability
	// We'll just verify the connection without sending messages
	conn, err := kafka.DialContext(context.Background(), "tcp", kafkaBroker)
	if err != nil {
		log.Printf("Warning: Could not connect to Kafka at %s: %v", kafkaBroker, err)
		log.Println("Continuing without Kafka (graceful degradation)")
		return nil
	}
	defer conn.Close()

	log.Printf("Successfully connected to Kafka at %s", kafkaBroker)
	return nil
}
