package main

import (
	"log"
	"net/http"
	"portfolio-rebalancer/internal/handlers"
	"portfolio-rebalancer/internal/repository"
	"portfolio-rebalancer/internal/services"
	"portfolio-rebalancer/internal/storage"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Initialize Elasticsearch for storing portfolios and transactions
	// Retries connection up to 5 times with 5-second delays
	if err := storage.InitElastic(); err != nil {
		log.Fatalf("Failed to initialize Elasticsearch: %v", err)
	}

	// Get Elasticsearch client
	esClient := storage.GetElasticsearchClient()

	// Initialize repositories
	portfolioRepo := repository.NewPortfolioRepository(esClient)

	portfolioService := services.NewPortfolioService(portfolioRepo)

	// Create HTTP router
	mux := http.NewServeMux()

	handlers.NewPortfolioHandler(mux, portfolioService)

	//http.HandleFunc("/portfolio", handlers.HandlePortfolio)
	//http.HandleFunc("/rebalance", handlers.HandleRebalance)

	server := &http.Server{
		Addr:         ":8081",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("Server started at :8081")
	log.Fatal(server.ListenAndServe())
}
