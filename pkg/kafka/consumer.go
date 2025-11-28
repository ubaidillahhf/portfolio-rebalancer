package kafka

import (
	"context"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
)

// StartConsumer starts a Kafka consumer in a background goroutine
// The handler function will be called for each message received
func StartConsumer(ctx context.Context, handler func(context.Context, kafka.Message) error) error {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	topic := os.Getenv("KAFKA_TOPIC")

	if kafkaBroker == "" || topic == "" {
		log.Println("Kafka consumer config not set; skipping consumer start")
		return nil
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{kafkaBroker},
		Topic:     topic,
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
		GroupID:   "portfolio-rebalancer-group",
	})

	go func() {
		defer reader.Close()
		log.Println("Kafka consumer started")

		for {
			select {
			case <-ctx.Done():
				log.Println("Kafka consumer shutting down")
				return
			default:
				msg, err := reader.ReadMessage(ctx)
				if err != nil {
					if err == context.Canceled {
						log.Println("Kafka consumer context canceled")
						return
					}
					log.Printf("Kafka read error: %v", err)
					continue
				}

				// Process message with handler
				if err := handler(ctx, msg); err != nil {
					log.Printf("Error processing Kafka message: %v", err)
				}
			}
		}
	}()

	return nil
}
