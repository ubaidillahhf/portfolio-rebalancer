package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"portfolio-rebalancer/internal/models"
)

// Mock rebalance service
type mockRebalanceService struct {
	calculateFunc func(currentAllocation, targetAllocation map[string]float64, userID string) []models.RebalanceTransaction
	publishFunc   func(ctx context.Context, transactions []models.RebalanceTransaction) error
	processFunc   func(ctx context.Context, message []byte) error
}

func (m *mockRebalanceService) CalculateRebalance(currentAllocation, targetAllocation map[string]float64, userID string) []models.RebalanceTransaction {
	if m.calculateFunc != nil {
		return m.calculateFunc(currentAllocation, targetAllocation, userID)
	}
	return []models.RebalanceTransaction{}
}

func (m *mockRebalanceService) PublishRebalanceTransactions(ctx context.Context, transactions []models.RebalanceTransaction) error {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, transactions)
	}
	return nil
}

func (m *mockRebalanceService) ProcessTransactions(ctx context.Context, message []byte) error {
	if m.processFunc != nil {
		return m.processFunc(ctx, message)
	}
	return nil
}

func TestHandleRebalance(t *testing.T) {
	tests := []struct {
		name              string
		method            string
		requestBody       interface{}
		mockPortfolio     *models.Portfolio
		mockPortfolioErr  error
		mockCalculate     []models.RebalanceTransaction
		mockPublishErr    error
		mockUpdateErr     error
		expectedStatus    int
	}{
		{
			name:   "successful rebalance",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"user_id": "user1",
				"new_allocation": map[string]float64{
					"stocks": 70.0,
					"bonds":  30.0,
				},
			},
			mockPortfolio: &models.Portfolio{
				UserID: "user1",
				Allocation: map[string]float64{
					"stocks": 60.0,
					"bonds":  40.0,
				},
				OriginalAllocation: map[string]float64{
					"stocks": 60.0,
					"bonds":  40.0,
				},
			},
			mockCalculate: []models.RebalanceTransaction{
				{
					UserID:           "user1",
					Action:           "BUY",
					Asset:            "stocks",
					RebalancePercent: 10.0,
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "method not allowed",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "invalid JSON",
			method:         http.MethodPost,
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "empty user_id",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"user_id": "",
				"new_allocation": map[string]float64{
					"stocks": 60.0,
					"bonds":  40.0,
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "empty new_allocation",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"user_id":        "user1",
				"new_allocation": map[string]float64{},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid allocation percentages",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"user_id": "user1",
				"new_allocation": map[string]float64{
					"stocks": 60.0,
					"bonds":  30.0,
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "portfolio not found",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"user_id": "user1",
				"new_allocation": map[string]float64{
					"stocks": 60.0,
					"bonds":  40.0,
				},
			},
			mockPortfolioErr: errors.New("not found"),
			expectedStatus:   http.StatusNotFound,
		},
		{
			name:   "publish error",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"user_id": "user1",
				"new_allocation": map[string]float64{
					"stocks": 60.0,
					"bonds":  40.0,
				},
			},
			mockPortfolio: &models.Portfolio{
				UserID: "user1",
				Allocation: map[string]float64{
					"stocks": 60.0,
					"bonds":  40.0,
				},
				OriginalAllocation: map[string]float64{
					"stocks": 60.0,
					"bonds":  40.0,
				},
			},
			mockCalculate: []models.RebalanceTransaction{
				{
					UserID: "user1",
					Action: "BUY",
					Asset:  "stocks",
				},
			},
			mockPublishErr: errors.New("kafka error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "no rebalancing needed",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"user_id": "user1",
				"new_allocation": map[string]float64{
					"stocks": 60.0,
					"bonds":  40.0,
				},
			},
			mockPortfolio: &models.Portfolio{
				UserID: "user1",
				Allocation: map[string]float64{
					"stocks": 60.0,
					"bonds":  40.0,
				},
				OriginalAllocation: map[string]float64{
					"stocks": 60.0,
					"bonds":  40.0,
				},
			},
			mockCalculate:  []models.RebalanceTransaction{},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPortfolioSvc := &mockPortfolioService{
				getFunc: func(ctx context.Context, userID string) (*models.Portfolio, error) {
					if tt.mockPortfolioErr != nil {
						return nil, tt.mockPortfolioErr
					}
					return tt.mockPortfolio, nil
				},
				updateFunc: func(ctx context.Context, portfolio models.Portfolio) error {
					return tt.mockUpdateErr
				},
			}

			mockRebalanceSvc := &mockRebalanceService{
				calculateFunc: func(currentAllocation, targetAllocation map[string]float64, userID string) []models.RebalanceTransaction {
					return tt.mockCalculate
				},
				publishFunc: func(ctx context.Context, transactions []models.RebalanceTransaction) error {
					return tt.mockPublishErr
				},
			}

			handler := &RebalanceHandler{
				rebalanceService: mockRebalanceSvc,
				portfolioService: mockPortfolioSvc,
			}

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else if tt.requestBody != nil {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(tt.method, "/rebalance", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.HandleRebalance(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
