package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"portfolio-rebalancer/internal/messaging"
	"portfolio-rebalancer/internal/models"
	"portfolio-rebalancer/internal/repository"
	"time"
)

type RebalanceService interface {
	CalculateRebalance(currentAllocation, targetAllocation map[string]float64, userID string) []models.RebalanceTransaction
	PublishRebalanceTransactions(ctx context.Context, transactions []models.RebalanceTransaction) error
	ProcessTransactions(ctx context.Context, message []byte) error
}

type RebalanceServiceImpl struct {
	transactionRepo repository.TransactionRepository
	publisher       messaging.Publisher
}

// NewRebalanceService creates a new rebalance service instance
func NewRebalanceService(transactionRepo repository.TransactionRepository, publisher messaging.Publisher) RebalanceService {
	return &RebalanceServiceImpl{
		transactionRepo: transactionRepo,
		publisher:       publisher,
	}
}

func (s *RebalanceServiceImpl) CalculateRebalance(currentAllocation, targetAllocation map[string]float64, userID string) []models.RebalanceTransaction {
	var transactions []models.RebalanceTransaction
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Calculate the difference between current and target for each asset
	for asset, targetPercent := range targetAllocation {
		currentPercent := currentAllocation[asset]
		diff := currentPercent - targetPercent

		// Skip if difference is negligible (less than 0.01%)
		if math.Abs(diff) < 0.01 {
			continue
		}

		var action string
		var rebalancePercent float64

		if diff > 0 {
			// Current allocation is higher than target, need to SELL
			action = "SELL"
			rebalancePercent = diff
		} else {
			// Current allocation is lower than target, need to BUY
			action = "BUY"
			rebalancePercent = math.Abs(diff)
		}

		transactions = append(transactions, models.RebalanceTransaction{
			UserID:           userID,
			Action:           action,
			Asset:            asset,
			RebalancePercent: rebalancePercent,
			Timestamp:        timestamp,
		})
	}

	// Check for assets in current allocation that are not in target (should sell all)
	for asset, currentPercent := range currentAllocation {
		if _, exists := targetAllocation[asset]; !exists && currentPercent > 0.01 {
			transactions = append(transactions, models.RebalanceTransaction{
				UserID:           userID,
				Action:           "SELL",
				Asset:            asset,
				RebalancePercent: currentPercent,
				Timestamp:        timestamp,
			})
		}
	}

	return transactions
}

func (s *RebalanceServiceImpl) PublishRebalanceTransactions(ctx context.Context, transactions []models.RebalanceTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	payload, err := json.Marshal(transactions)
	if err != nil {
		return fmt.Errorf("failed to marshal transactions: %w", err)
	}

	if err := s.publisher.Publish(ctx, payload); err != nil {
		return fmt.Errorf("failed to publish transactions: %w", err)
	}

	return nil
}

// ProcessTransactions processes rebalance transactions from Kafka message
func (s *RebalanceServiceImpl) ProcessTransactions(ctx context.Context, message []byte) error {
	// Skip empty or invalid messages (e.g., health check pings)
	if len(message) == 0 {
		return nil
	}

	var transactions []models.RebalanceTransaction
	if err := json.Unmarshal(message, &transactions); err != nil {
		// Log but don't fail on invalid JSON (could be health check messages)
		log.Printf("Skipping invalid message (not JSON array): %s", string(message))
		return nil
	}

	// Skip if no transactions to process
	if len(transactions) == 0 {
		return nil
	}

	log.Printf("Processing %d rebalance transactions from consumer kafka", len(transactions))

	// Save each transaction to Elasticsearch
	for _, tx := range transactions {
		if err := s.transactionRepo.SaveTransaction(ctx, tx); err != nil {
			log.Printf("Failed to save transaction for user %s, asset %s: %v", tx.UserID, tx.Asset, err)
			// Continue processing other transactions even if one fails
			continue
		}
	}

	log.Printf("Successfully processed %d transactions", len(transactions))
	return nil
}
