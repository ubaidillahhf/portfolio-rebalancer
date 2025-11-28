package models

import (
	"testing"
)

func TestValidateAllocation(t *testing.T) {
	tests := []struct {
		name        string
		allocation  map[string]float64
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid allocation - exactly 100%",
			allocation: map[string]float64{
				"stocks": 60.0,
				"bonds":  30.0,
				"gold":   10.0,
			},
			expectError: false,
		},
		{
			name: "valid allocation - within tolerance (99.995%)",
			allocation: map[string]float64{
				"stocks": 60.0,
				"bonds":  30.0,
				"gold":   9.995,
			},
			expectError: false,
		},
		{
			name: "valid allocation - within tolerance (100.005%)",
			allocation: map[string]float64{
				"stocks": 60.0,
				"bonds":  30.005,
				"gold":   10.0,
			},
			expectError: false,
		},
		{
			name: "invalid allocation - sum > 100%",
			allocation: map[string]float64{
				"stocks": 60.0,
				"bonds":  30.0,
				"gold":   15.0,
			},
			expectError: true,
			errorMsg:    "allocation must sum to 100%, got 105.00%",
		},
		{
			name: "invalid allocation - sum < 100%",
			allocation: map[string]float64{
				"stocks": 50.0,
				"bonds":  30.0,
				"gold":   10.0,
			},
			expectError: true,
			errorMsg:    "allocation must sum to 100%, got 90.00%",
		},
		{
			name: "invalid allocation - negative percentage",
			allocation: map[string]float64{
				"stocks": -10.0,
				"bonds":  60.0,
				"gold":   50.0,
			},
			expectError: true,
			errorMsg:    "allocation for stocks cannot be negative",
		},
		{
			name: "invalid allocation - exceeds 100%",
			allocation: map[string]float64{
				"stocks": 150.0,
			},
			expectError: true,
			errorMsg:    "allocation must sum to 100%, got 150.00%",
		},
		{
			name: "invalid allocation - empty",
			allocation: map[string]float64{},
			expectError: true,
			errorMsg:    "allocation cannot be empty",
		},
		{
			name: "valid allocation - single asset 100%",
			allocation: map[string]float64{
				"stocks": 100.0,
			},
			expectError: false,
		},
		{
			name: "valid allocation - multiple assets with decimals",
			allocation: map[string]float64{
				"stocks": 33.33,
				"bonds":  33.33,
				"gold":   33.34,
			},
			expectError: false,
		},
		{
			name: "invalid allocation - slightly outside tolerance (100.02%)",
			allocation: map[string]float64{
				"stocks": 60.0,
				"bonds":  30.02,
				"gold":   10.0,
			},
			expectError: true,
			errorMsg:    "allocation must sum to 100%, got 100.02%",
		},
		{
			name: "invalid allocation - zero percentage",
			allocation: map[string]float64{
				"stocks": 0.0,
				"bonds":  50.0,
				"gold":   50.0,
			},
			expectError: false, // Zero is allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAllocation(tt.allocation)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}
