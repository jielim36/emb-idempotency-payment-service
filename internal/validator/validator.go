package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

// Greater than or equal
func DecimalGt(fl validator.FieldLevel) bool {
	param := fl.Param()
	if param == "" {
		return false
	}

	minimumValue, err := decimal.NewFromString(param)
	if err != nil {
		return false
	}

	value, ok := fl.Field().Interface().(decimal.Decimal)
	if !ok {
		return false
	}

	return value.GreaterThan(minimumValue)
}
