package validator

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func IranianPhoneValidator(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	return len(phone) == 11 && strings.HasPrefix(phone, "09")
}

func RegisterIranianPhoneValidator(v *validator.Validate) error {
	return v.RegisterValidation("iranian_phone", IranianPhoneValidator)
}
