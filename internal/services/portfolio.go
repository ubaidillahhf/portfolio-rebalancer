package services

import (
	"context"
	"fmt"
	"portfolio-rebalancer/internal/models"
	"portfolio-rebalancer/internal/repository"
)

type PortfolioService interface {
	CreatePortfolio(ctx context.Context, p models.Portfolio) (*models.Portfolio, error)
}

// PortfolioServiceImpl handles portfolio business logic
type PortfolioServiceImpl struct {
	portfolioRepository repository.PortfolioRepository
}

// NewPortfolioService creates a new portfolio service instance
func NewPortfolioService(portfolioRepository repository.PortfolioRepository) PortfolioService {
	return &PortfolioServiceImpl{
		portfolioRepository,
	}
}

// CreatePortfolio creates a new portfolio with validation
func (s *PortfolioServiceImpl) CreatePortfolio(ctx context.Context, p models.Portfolio) (*models.Portfolio, error) {
	// Validate user ID
	if p.UserID == "" {
		return nil, fmt.Errorf("user_id is required and cannot be empty")
	}

	// Validate allocation exists
	if len(p.Allocation) == 0 {
		return nil, fmt.Errorf("allocation is required and cannot be empty")
	}

	// Validate allocation percentages
	if err := models.ValidateAllocation(p.Allocation); err != nil {
		return nil, err
	}

	// Set original allocation to current allocation (this is the target to maintain)
	p.OriginalAllocation = make(map[string]float64)
	for k, v := range p.Allocation {
		p.OriginalAllocation[k] = v
	}

	// Save to storage
	if err := s.portfolioRepository.Save(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to save portfolio: %w", err)
	}

	return &p, nil
}
