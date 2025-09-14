package validator

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var validatorEngine *validator.Validate

const (
	DecimalGreaterThan string = "decimalGt"
)

func RegisterValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation(DecimalGreaterThan, DecimalGt)
		validatorEngine = v
	}
}

func GetValidator() *validator.Validate {
	return validatorEngine
}
