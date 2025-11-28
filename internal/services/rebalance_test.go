package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"portfolio-rebalancer/internal/models"
)

// Mock transaction repository
type mockTransactionRepository struct {
	saveFunc func(ctx context.Context, tx models.RebalanceTransaction) error
}

func (m *mockTransactionRepository) SaveTransaction(ctx context.Context, tx models.RebalanceTransaction) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, tx)
	}
	return nil
}

// Mock publisher
type mockPublisher struct {
	publishFunc func(ctx context.Context, message []byte) error
}

func (m *mockPublisher) Publish(ctx context.Context, message []byte) error {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, message)
	}
	return nil
}

func TestCalculateRebalance(t *testing.T) {
	tests := []struct {
		name               string
		currentAllocation  map[string]float64
		targetAllocation   map[string]float64
		userID             string
		expectedTxCount    int
		expectedActions    map[string]string // asset -> action
	}{
		{
			name: "need to buy and sell",
			currentAllocation: map[string]float64{
				"stocks": 5.0,
				"bonds":  55.0,
				"gold":   40.0,
			},
			targetAllocation: map[string]float64{
				"stocks": 10.0,
				"bonds":  50.0,
				"gold":   40.0,
			},
			userID:          "user1",
			expectedTxCount: 2,
			expectedActions: map[string]string{
				"stocks": "BUY",
				"bonds":  "SELL",
			},
		},
		{
			name: "already balanced",
			currentAllocation: map[string]float64{
				"stocks": 60.0,
				"bonds":  30.0,
				"gold":   10.0,
			},
			targetAllocation: map[string]float64{
				"stocks": 60.0,
				"bonds":  30.0,
				"gold":   10.0,
			},
			userID:          "user1",
			expectedTxCount: 0,
		},
		{
			name: "within tolerance - no rebalance needed",
			currentAllocation: map[string]float64{
				"stocks": 60.005,
				"bonds":  29.995,
				"gold":   10.0,
			},
			targetAllocation: map[string]float64{
				"stocks": 60.0,
				"bonds":  30.0,
				"gold":   10.0,
			},
			userID:          "user1",
			expectedTxCount: 0,
		},
		{
			name: "sell all of removed asset",
			currentAllocation: map[string]float64{
				"stocks": 50.0,
				"bonds":  30.0,
				"gold":   20.0,
			},
			targetAllocation: map[string]float64{
				"stocks": 60.0,
				"bonds":  40.0,
			},
			userID:          "user1",
			expectedTxCount: 3,
			expectedActions: map[string]string{
				"stocks": "BUY",
				"bonds":  "BUY",
				"gold":   "SELL",
			},
		},
		{
			name: "all assets need rebalancing",
			currentAllocation: map[string]float64{
				"stocks": 70.0,
				"bonds":  20.0,
				"gold":   10.0,
			},
			targetAllocation: map[string]float64{
				"stocks": 60.0,
				"bonds":  30.0,
				"gold":   10.0,
			},
			userID:          "user1",
			expectedTxCount: 2,
			expectedActions: map[string]string{
				"stocks": "SELL",
				"bonds":  "BUY",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockTransactionRepository{}
			mockPub := &mockPublisher{}
			service := NewRebalanceService(mockRepo, mockPub)

			transactions := service.CalculateRebalance(tt.currentAllocation, tt.targetAllocation, tt.userID)

			if len(transactions) != tt.expectedTxCount {
				t.Errorf("expected %d transactions, got %d", tt.expectedTxCount, len(transactions))
			}

			// Verify actions
			for _, tx := range transactions {
				if expectedAction, exists := tt.expectedActions[tx.Asset]; exists {
					if tx.Action != expectedAction {
						t.Errorf("expected action %s for asset %s, got %s", expectedAction, tx.Asset, tx.Action)
					}
				}
			}
		})
	}
}

func TestPublishRebalanceTransactions(t *testing.T) {
	tests := []struct {
		name         string
		transactions []models.RebalanceTransaction
		mockPublish  func(ctx context.Context, message []byte) error
		expectError  bool
		errorMsg     string
	}{
		{
			name: "successful publish",
			transactions: []models.RebalanceTransaction{
				{
					UserID:           "user1",
					Action:           "BUY",
					Asset:            "stocks",
					RebalancePercent: 10.0,
					Timestamp:        "2025-01-01T00:00:00Z",
				},
			},
			expectError: false,
		},
		{
			name:         "empty transactions - no publish",
			transactions: []models.RebalanceTransaction{},
			expectError:  false,
		},
		{
			name: "publish error",
			transactions: []models.RebalanceTransaction{
				{
					UserID:           "user1",
					Action:           "BUY",
					Asset:            "stocks",
					RebalancePercent: 10.0,
					Timestamp:        "2025-01-01T00:00:00Z",
				},
			},
			mockPublish: func(ctx context.Context, message []byte) error {
				return errors.New("kafka error")
			},
			expectError: true,
			errorMsg:    "failed to publish transactions: kafka error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockTransactionRepository{}
			mockPub := &mockPublisher{
				publishFunc: tt.mockPublish,
			}
			service := NewRebalanceService(mockRepo, mockPub)

			err := service.PublishRebalanceTransactions(context.Background(), tt.transactions)

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

func TestProcessTransactions(t *testing.T) {
	tests := []struct {
		name        string
		message     []byte
		mockSave    func(ctx context.Context, tx models.RebalanceTransaction) error
		expectError bool
	}{
		{
			name: "successful processing",
			message: func() []byte {
				txs := []models.RebalanceTransaction{
					{
						UserID:           "user1",
						Action:           "BUY",
						Asset:            "stocks",
						RebalancePercent: 10.0,
						Timestamp:        "2025-01-01T00:00:00Z",
					},
				}
				data, _ := json.Marshal(txs)
				return data
			}(),
			expectError: false,
		},
		{
			name:        "empty message",
			message:     []byte{},
			expectError: false,
		},
		{
			name:        "invalid JSON",
			message:     []byte("invalid json"),
			expectError: false, // Should log and skip, not error
		},
		{
			name: "empty transaction array",
			message: func() []byte {
				txs := []models.RebalanceTransaction{}
				data, _ := json.Marshal(txs)
				return data
			}(),
			expectError: false,
		},
		{
			name: "save error - should continue",
			message: func() []byte {
				txs := []models.RebalanceTransaction{
					{
						UserID:           "user1",
						Action:           "BUY",
						Asset:            "stocks",
						RebalancePercent: 10.0,
						Timestamp:        "2025-01-01T00:00:00Z",
					},
				}
				data, _ := json.Marshal(txs)
				return data
			}(),
			mockSave: func(ctx context.Context, tx models.RebalanceTransaction) error {
				return errors.New("save failed")
			},
			expectError: false, // Should log error but not fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockTransactionRepository{
				saveFunc: tt.mockSave,
			}
			mockPub := &mockPublisher{}
			service := NewRebalanceService(mockRepo, mockPub)

			err := service.ProcessTransactions(context.Background(), tt.message)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}
