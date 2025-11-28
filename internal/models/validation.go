package models

import (
	"fmt"
	"math"
)

// ValidateAllocation checks if allocation percentages sum to 100
func ValidateAllocation(allocation map[string]float64) error {
	if len(allocation) == 0 {
		return fmt.Errorf("allocation cannot be empty")
	}

	var total float64
	for asset, percent := range allocation {
		if percent < 0 {
			return fmt.Errorf("allocation for %s cannot be negative", asset)
		}
		total += percent
	}

	// Allow small floating point errors
	if math.Abs(total-100.0) > 0.01 {
		return fmt.Errorf("allocation must sum to 100%%, got %.2f%%", total)
	}

	return nil
}
