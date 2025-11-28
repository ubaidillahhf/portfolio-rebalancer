package services

import (
	"context"
	"errors"
	"testing"

	"portfolio-rebalancer/internal/models"
)

// Mock repository
type mockPortfolioRepository struct {
	saveFunc      func(ctx context.Context, portfolio models.Portfolio) error
	getByUserIDFunc func(ctx context.Context, userID string) (*models.Portfolio, error)
}

func (m *mockPortfolioRepository) Save(ctx context.Context, portfolio models.Portfolio) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, portfolio)
	}
	return nil
}

func (m *mockPortfolioRepository) GetByUserID(ctx context.Context, userID string) (*models.Portfolio, error) {
	if m.getByUserIDFunc != nil {
		return m.getByUserIDFunc(ctx, userID)
	}
	return nil, errors.New("not found")
}

func TestCreatePortfolio(t *testing.T) {
	tests := []struct {
		name        string
		portfolio   models.Portfolio
		mockSave    func(ctx context.Context, portfolio models.Portfolio) error
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid portfolio creation",
			portfolio: models.Portfolio{
				UserID: "user1",
				Allocation: map[string]float64{
					"stocks": 60.0,
					"bonds":  30.0,
					"gold":   10.0,
				},
			},
			mockSave:    nil,
			expectError: false,
		},
		{
			name: "empty user ID",
			portfolio: models.Portfolio{
				UserID: "",
				Allocation: map[string]float64{
					"stocks": 60.0,
					"bonds":  40.0,
				},
			},
			expectError: true,
			errorMsg:    "user_id is required and cannot be empty",
		},
		{
			name: "empty allocation",
			portfolio: models.Portfolio{
				UserID:     "user1",
				Allocation: map[string]float64{},
			},
			expectError: true,
			errorMsg:    "allocation is required and cannot be empty",
		},
		{
			name: "invalid allocation - sum not 100%",
			portfolio: models.Portfolio{
				UserID: "user1",
				Allocation: map[string]float64{
					"stocks": 60.0,
					"bonds":  30.0,
				},
			},
			expectError: true,
			errorMsg:    "allocation must sum to 100%, got 90.00%",
		},
		{
			name: "repository save error",
			portfolio: models.Portfolio{
				UserID: "user1",
				Allocation: map[string]float64{
					"stocks": 60.0,
					"bonds":  40.0,
				},
			},
			mockSave: func(ctx context.Context, portfolio models.Portfolio) error {
				return errors.New("database error")
			},
			expectError: true,
			errorMsg:    "failed to save portfolio: database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockPortfolioRepository{
				saveFunc: tt.mockSave,
			}
			service := NewPortfolioService(mockRepo)

			result, err := service.CreatePortfolio(context.Background(), tt.portfolio)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
					return
				}
				if result == nil {
					t.Errorf("expected result but got nil")
					return
				}
				// Verify original allocation is set
				if len(result.OriginalAllocation) == 0 {
					t.Errorf("expected original allocation to be set")
				}
			}
		})
	}
}

func TestGetPortfolio(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		mockGet     func(ctx context.Context, userID string) (*models.Portfolio, error)
		expectError bool
		errorMsg    string
	}{
		{
			name:   "successful get",
			userID: "user1",
			mockGet: func(ctx context.Context, userID string) (*models.Portfolio, error) {
				return &models.Portfolio{
					UserID: userID,
					Allocation: map[string]float64{
						"stocks": 60.0,
						"bonds":  40.0,
					},
				}, nil
			},
			expectError: false,
		},
		{
			name:        "empty user ID",
			userID:      "",
			expectError: true,
			errorMsg:    "user_id is required and cannot be empty",
		},
		{
			name:   "portfolio not found",
			userID: "user1",
			mockGet: func(ctx context.Context, userID string) (*models.Portfolio, error) {
				return nil, errors.New("not found")
			},
			expectError: true,
			errorMsg:    "portfolio not found for user user1: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockPortfolioRepository{
				getByUserIDFunc: tt.mockGet,
			}
			service := NewPortfolioService(mockRepo)

			result, err := service.GetPortfolio(context.Background(), tt.userID)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
					return
				}
				if result == nil {
					t.Errorf("expected result but got nil")
				}
			}
		})
	}
}

func TestUpdatePortfolio(t *testing.T) {
	tests := []struct {
		name        string
		portfolio   models.Portfolio
		mockSave    func(ctx context.Context, portfolio models.Portfolio) error
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful update",
			portfolio: models.Portfolio{
				UserID: "user1",
				Allocation: map[string]float64{
					"stocks": 70.0,
					"bonds":  30.0,
				},
			},
			expectError: false,
		},
		{
			name: "empty user ID",
			portfolio: models.Portfolio{
				UserID: "",
			},
			expectError: true,
			errorMsg:    "user_id is required and cannot be empty",
		},
		{
			name: "invalid allocation",
			portfolio: models.Portfolio{
				UserID: "user1",
				Allocation: map[string]float64{
					"stocks": 60.0,
				},
			},
			expectError: true,
			errorMsg:    "allocation must sum to 100%, got 60.00%",
		},
		{
			name: "repository error",
			portfolio: models.Portfolio{
				UserID: "user1",
				Allocation: map[string]float64{
					"stocks": 60.0,
					"bonds":  40.0,
				},
			},
			mockSave: func(ctx context.Context, portfolio models.Portfolio) error {
				return errors.New("save failed")
			},
			expectError: true,
			errorMsg:    "failed to update portfolio: save failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockPortfolioRepository{
				saveFunc: tt.mockSave,
			}
			service := NewPortfolioService(mockRepo)

			err := service.UpdatePortfolio(context.Background(), tt.portfolio)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}
