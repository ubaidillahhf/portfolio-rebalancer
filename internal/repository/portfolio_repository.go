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

type PortfolioRepository interface {
	Save(ctx context.Context, portfolio models.Portfolio) error
}

// PortfolioRepository implements PortfolioRepository using Elasticsearch
type PortfolioRepositoryImpl struct {
	client *elasticsearch.Client
}

// NewPortfolioRepository creates a new Elasticsearch portfolio repository
func NewPortfolioRepository(client *elasticsearch.Client) PortfolioRepository {
	return &PortfolioRepositoryImpl{
		client: client,
	}
}

// Save saves a portfolio to Elasticsearch
func (r *PortfolioRepositoryImpl) Save(ctx context.Context, p models.Portfolio) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	body, err := json.Marshal(p)
	if err != nil {
		return err
	}

	res, err := r.client.Index("portfolios", bytes.NewReader(body),
		r.client.Index.WithDocumentID(p.UserID),
		r.client.Index.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error saving portfolio: %s", res.String())
	}

	log.Printf("Portfolio saved for user %s", p.UserID)
	return nil
}
