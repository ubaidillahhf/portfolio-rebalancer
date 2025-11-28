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

// Mock portfolio service
type mockPortfolioService struct {
	createFunc func(ctx context.Context, p models.Portfolio) (*models.Portfolio, error)
	getFunc    func(ctx context.Context, userID string) (*models.Portfolio, error)
	updateFunc func(ctx context.Context, portfolio models.Portfolio) error
}

func (m *mockPortfolioService) CreatePortfolio(ctx context.Context, p models.Portfolio) (*models.Portfolio, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, p)
	}
	return &p, nil
}

func (m *mockPortfolioService) GetPortfolio(ctx context.Context, userID string) (*models.Portfolio, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, userID)
	}
	return nil, errors.New("not found")
}

func (m *mockPortfolioService) UpdatePortfolio(ctx context.Context, portfolio models.Portfolio) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, portfolio)
	}
	return nil
}

func TestHandlePortfolio(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET method",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST method",
			method:         http.MethodPost,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "PUT method not allowed",
			method:         http.MethodPut,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "DELETE method not allowed",
			method:         http.MethodDelete,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockPortfolioService{
				createFunc: func(ctx context.Context, p models.Portfolio) (*models.Portfolio, error) {
					return &p, nil
				},
				getFunc: func(ctx context.Context, userID string) (*models.Portfolio, error) {
					return &models.Portfolio{
						UserID: userID,
						Allocation: map[string]float64{
							"stocks": 60.0,
							"bonds":  40.0,
						},
					}, nil
				},
			}

			handler := &PortfolioHandler{
				portfolioService: mockService,
			}

			var req *http.Request
			if tt.method == http.MethodPost {
				body := map[string]interface{}{
					"user_id": "user1",
					"allocation": map[string]float64{
						"stocks": 60.0,
						"bonds":  40.0,
					},
				}
				jsonBody, _ := json.Marshal(body)
				req = httptest.NewRequest(tt.method, "/portfolio", bytes.NewBuffer(jsonBody))
			} else if tt.method == http.MethodGet {
				req = httptest.NewRequest(tt.method, "/portfolio?user_id=user1", nil)
			} else {
				req = httptest.NewRequest(tt.method, "/portfolio", nil)
			}

			w := httptest.NewRecorder()
			handler.HandlePortfolio(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandleCreatePortfolio(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockCreate     func(ctx context.Context, p models.Portfolio) (*models.Portfolio, error)
		expectedStatus int
	}{
		{
			name: "successful creation",
			requestBody: map[string]interface{}{
				"user_id": "user1",
				"allocation": map[string]float64{
					"stocks": 60.0,
					"bonds":  40.0,
				},
			},
			mockCreate: func(ctx context.Context, p models.Portfolio) (*models.Portfolio, error) {
				return &p, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			requestBody: map[string]interface{}{
				"user_id": "user1",
				"allocation": map[string]float64{
					"stocks": 60.0,
					"bonds":  40.0,
				},
			},
			mockCreate: func(ctx context.Context, p models.Portfolio) (*models.Portfolio, error) {
				return nil, errors.New("service error")
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockPortfolioService{
				createFunc: tt.mockCreate,
			}

			handler := &PortfolioHandler{
				portfolioService: mockService,
			}

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/portfolio", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.HandlePortfolio(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandleGetPortfolio(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockGet        func(ctx context.Context, userID string) (*models.Portfolio, error)
		expectedStatus int
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
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user_id",
			userID:         "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "portfolio not found",
			userID: "user1",
			mockGet: func(ctx context.Context, userID string) (*models.Portfolio, error) {
				return nil, errors.New("not found")
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockPortfolioService{
				getFunc: tt.mockGet,
			}

			handler := &PortfolioHandler{
				portfolioService: mockService,
			}

			url := "/portfolio"
			if tt.userID != "" {
				url += "?user_id=" + tt.userID
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			handler.HandlePortfolio(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
