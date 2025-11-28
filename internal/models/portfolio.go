package models

type Portfolio struct {
	UserID             string             `json:"user_id"`
	Allocation         map[string]float64 `json:"allocation"`          // Current user allocation in percentage terms
	OriginalAllocation map[string]float64 `json:"original_allocation"` // Target allocation to maintain
}

type UpdatedPortfolio struct {
	UserID        string             `json:"user_id"`
	NewAllocation map[string]float64 `json:"new_allocation"` // Updated user allocation from provider in percentage terms
}

type RebalanceTransaction struct {
	UserID           string  `json:"user_id"`
	Action           string  `json:"action"`            // BUY or SELL
	Asset            string  `json:"asset"`             // stocks, bonds, gold, etc.
	RebalancePercent float64 `json:"rebalance_percent"` // percentage to buy/sell
	Timestamp        string  `json:"timestamp"`
}
