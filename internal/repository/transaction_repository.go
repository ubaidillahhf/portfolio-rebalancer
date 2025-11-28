package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"portfolio-rebalancer/internal/models"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

type TransactionRepository interface {
	SaveTransaction(ctx context.Context, tx models.RebalanceTransaction) error
}

// TransactionRepository implements TransactionRepository using Elasticsearch
type TransactionRepositoryImpl struct {
	client *elasticsearch.Client
}

// NewTransactionRepository creates a new Elasticsearch portfolio repository
func NewTransactionRepository(client *elasticsearch.Client) TransactionRepository {
	return &TransactionRepositoryImpl{
		client: client,
	}
}

// SaveTransaction saves a rebalance transaction to Elasticsearch
func (r *TransactionRepositoryImpl) SaveTransaction(ctx context.Context, tx models.RebalanceTransaction) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	body, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	// Use timestamp + userID + asset as document ID for uniqueness
	docID := fmt.Sprintf("%s_%s_%s", tx.UserID, tx.Asset, tx.Timestamp)

	res, err := r.client.Index("rebalance_transactions", bytes.NewReader(body),
		r.client.Index.WithDocumentID(docID),
		r.client.Index.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error saving rebalance transaction: %s", res.String())
	}

	log.Printf("Rebalance transaction saved: User=%s, Action=%s, Asset=%s, Percent=%.2f%%",
		tx.UserID, tx.Action, tx.Asset, tx.RebalancePercent)
	return nil
}
