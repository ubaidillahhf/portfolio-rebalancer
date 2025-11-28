package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"portfolio-rebalancer/internal/models"
	"portfolio-rebalancer/internal/services"
)

type PortfolioHandler struct {
	portfolioService services.PortfolioService
}

// NewPortfolioHandler creates a new portfolio handler with injected dependencies
func NewPortfolioHandler(mux *http.ServeMux, portfolioService services.PortfolioService) {
	handler := &PortfolioHandler{
		portfolioService: portfolioService,
	}

	// Register routes
	mux.HandleFunc("/portfolio", handler.HandlePortfolio)
}

// Uses HTTP method to determine action - proper REST design
func (h *PortfolioHandler) HandlePortfolio(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetPortfolio(w, r)
	case http.MethodPost:
		h.handleCreatePortfolio(w, r)
	default:
		RespondWithError(w, http.StatusMethodNotAllowed, "Only GET and POST methods are allowed")
	}
}

// HandlePortfolio handles new portfolio creation requests (feel free to update the request parameter/model)
// Sample Request (POST /portfolio):
//
//	{
//	    "user_id": "1",
//	    "allocation": {"stocks": 60, "bonds": 30, "gold": 10}
//	}
func (h *PortfolioHandler) handleCreatePortfolio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var p models.Portfolio
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	createdPortfolio, err := h.portfolioService.CreatePortfolio(r.Context(), p)
	if err != nil {
		log.Printf("Failed to create portfolio for user %s: %v", p.UserID, err)
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	RespondWithJSON(w, http.StatusCreated, createdPortfolio)
}

// handleGetPortfolio retrieves a user's portfolio
// GET /portfolio?user_id=1
func (h *PortfolioHandler) handleGetPortfolio(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		RespondWithError(w, http.StatusBadRequest, "user_id query parameter is required")
		return
	}

	portfolio, err := h.portfolioService.GetPortfolio(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get portfolio for user %s: %v", userID, err)
		RespondWithError(w, http.StatusNotFound, "Portfolio not found for user "+userID)
		return
	}

	log.Printf("Portfolio retrieved for user %s", userID)
	RespondWithJSON(w, http.StatusOK, portfolio)
}
