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

	// Register routes with logging middleware
	mux.HandleFunc("/portfolio", handler.HandlePortfolio)
	mux.HandleFunc("/rebalance", handler.HandleRebalance)
}

// Uses HTTP method to determine action - proper REST design
func (h *PortfolioHandler) HandlePortfolio(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
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

// HandleRebalance handles portfolio rebalance requests from 3rd party provider (feel free to update the request parameter/model)
// Sample Request (POST /rebalance):
//
//	{
//	    "user_id": "1",
//	    "new_allocation": {"stocks": 70, "bonds": 20, "gold": 10}
//	}
func (h *PortfolioHandler) HandleRebalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.UpdatedPortfolio
	json.NewDecoder(r.Body).Decode(&req)

	log.Println("HandleRebalance==", req)

	// TODO: Add Logic here

	w.WriteHeader(http.StatusOK)
}
