package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"portfolio-rebalancer/internal/models"
	"portfolio-rebalancer/internal/services"
)

type RebalanceHandler struct {
	rebalanceService services.RebalanceService
	portfolioService services.PortfolioService
}

// NewRebalanceHandler creates a new rebalance handler with injected dependencies
func NewRebalanceHandler(
	mux *http.ServeMux,
	rebalanceService services.RebalanceService,
	portfolioService services.PortfolioService,
) {
	handler := &RebalanceHandler{
		rebalanceService: rebalanceService,
		portfolioService: portfolioService,
	}

	// Register routes
	mux.HandleFunc("/rebalance", handler.HandleRebalance)
}

// HandleRebalance handles portfolio rebalance requests from 3rd party provider
// Sample Request (POST /rebalance):
//
//	{
//	    "user_id": "1",
//	    "new_allocation": {"stocks": 70, "bonds": 20, "gold": 10}
//	}
func (h *RebalanceHandler) HandleRebalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Only POST method is allowed")
		return
	}

	var req models.UpdatedPortfolio
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid JSON format in request body")
		return
	}

	// Validate user ID
	if req.UserID == "" {
		RespondWithError(w, http.StatusBadRequest, "user_id is required and cannot be empty")
		return
	}

	// Validate new allocation exists
	if len(req.NewAllocation) == 0 {
		RespondWithError(w, http.StatusBadRequest, "new_allocation is required and cannot be empty")
		return
	}

	// Validate new allocation percentages
	if err := models.ValidateAllocation(req.NewAllocation); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Fetch original portfolio
	portfolio, err := h.portfolioService.GetPortfolio(r.Context(), req.UserID)
	if err != nil {
		log.Printf("Failed to get portfolio for user %s: %v", req.UserID, err)
		RespondWithError(w, http.StatusNotFound, "User portfolio not found. Please create portfolio first using /portfolio endpoint")
		return
	}

	// Calculate rebalance transactions
	transactions := h.rebalanceService.CalculateRebalance(req.NewAllocation, portfolio.OriginalAllocation, req.UserID)

	// Publish transactions for async processing
	if err := h.rebalanceService.PublishRebalanceTransactions(r.Context(), transactions); err != nil {
		log.Printf("Failed to publish transactions for user %s: %v", req.UserID, err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to queue rebalance transactions. Please try again")
		return
	}

	if len(transactions) > 0 {
		log.Printf("Published %d transactions for user %s", len(transactions), req.UserID)
	} else {
		log.Printf("No rebalancing needed for user %s - portfolio already at target allocation", req.UserID)
	}

	// Update portfolio's current allocation
	portfolio.Allocation = req.NewAllocation
	if err := h.portfolioService.UpdatePortfolio(r.Context(), *portfolio); err != nil {
		log.Printf("Failed to update portfolio for user %s: %v", req.UserID, err)
		// Don't fail the request, transactions are already queued
	}

	// Return the calculated transactions
	response := map[string]interface{}{
		"user_id":           req.UserID,
		"transactions":      transactions,
		"transaction_count": len(transactions),
		"message":           "Rebalance transactions queued for processing",
	}

	if len(transactions) == 0 {
		response["message"] = "No rebalancing needed - portfolio already at target allocation"
	}

	RespondWithJSON(w, http.StatusOK, response)
}
