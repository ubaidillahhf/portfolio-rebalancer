package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"portfolio-rebalancer/internal/handlers"
	"portfolio-rebalancer/internal/messaging"
	"portfolio-rebalancer/internal/repository"
	"portfolio-rebalancer/internal/services"
	"portfolio-rebalancer/pkg/elasticsearch"
	"portfolio-rebalancer/pkg/kafka"

	"github.com/joho/godotenv"
	kafkago "github.com/segmentio/kafka-go"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Initialize Elasticsearch for storing portfolios and transactions
	// Retries connection up to 5 times with 5-second delays
	if err := elasticsearch.Init(); err != nil {
		log.Fatalf("Failed to initialize Elasticsearch: %v", err)
	}

	// Get Elasticsearch client
	esClient := elasticsearch.GetClient()

	// Initialize repositories
	portfolioRepo := repository.NewPortfolioRepository(esClient)
	transactionRepo := repository.NewTransactionRepository(esClient)

	// Initialize Kafka producer for async transaction processing
	// Non-fatal if Kafka is unavailable (graceful degradation)
	var publisher messaging.Publisher
	if err := kafka.Init(); err != nil {
		log.Printf("Warning: Failed to initialize Kafka: %v", err)
	} else {
		publisher = kafka.NewPublisher(kafka.GetWriter())
	}

	// Services
	portfolioService := services.NewPortfolioService(portfolioRepo)
	rebalanceService := services.NewRebalanceService(transactionRepo, publisher)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Kafka consumer to process rebalance transactions asynchronously
	// Runs in background goroutine, processing messages from the queue
	if err := kafka.StartConsumer(ctx, func(ctx context.Context, msg kafkago.Message) error {
		return rebalanceService.ProcessTransactions(ctx, msg.Value)
	}); err != nil {
		log.Printf("Warning: Failed to start Kafka consumer: %v", err)
	}

	// Create HTTP router
	mux := http.NewServeMux()

	// Register handlers - each with single responsibility
	handlers.NewPortfolioHandler(mux, portfolioService)
	handlers.NewRebalanceHandler(mux, rebalanceService, portfolioService)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Println("Server started at :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down server gracefully...")

	// Cancel context to stop Kafka consumer
	cancel()

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server gracefully
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close Kafka writer if it exists
	if writer := kafka.GetWriter(); writer != nil {
		if err := writer.Close(); err != nil {
			log.Printf("Error closing Kafka writer: %v", err)
		} else {
			log.Println("Kafka writer closed")
		}
	}

	log.Println("Server stopped gracefully")
}
