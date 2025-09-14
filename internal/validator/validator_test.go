package validator_test

import (
	"fmt"
	"testing"

	// import your actual validator package
	customValidator "payment-service/internal/validator"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDecimalGt(t *testing.T) {
	// Register validators first
	customValidator.RegisterValidators()
	validatorEngine := customValidator.GetValidator()

	// Test struct for validation
	type TestStruct struct {
		Amount decimal.Decimal `binding:"decimalGt=10.5"` // Note: use the exact tag name you registered
	}

	tests := []struct {
		name     string
		amount   decimal.Decimal
		expected bool
	}{
		{
			name:     "value greater than threshold",
			amount:   decimal.NewFromFloat(15.75),
			expected: true,
		},
		{
			name:     "value less than threshold",
			amount:   decimal.NewFromFloat(5.25),
			expected: false,
		},
		{
			name:     "value equal to threshold",
			amount:   decimal.NewFromFloat(10.5),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := TestStruct{Amount: tt.amount}
			err := validatorEngine.Struct(testStruct)
			errMsg := fmt.Sprintf("Amount: %s, Expected: %t, CurrentError: %v", tt.amount.String(), tt.expected, err)

			if tt.expected {
				assert.NoError(t, err, errMsg)
			} else {
				assert.Error(t, err, errMsg)
				assert.True(t, isValidationError(err), "error type is not validation error")
			}
		})
	}
}

func isValidationError(err error) bool {
	_, ok := err.(validator.ValidationErrors)
	return ok
}
