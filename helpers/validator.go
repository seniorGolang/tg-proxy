package helpers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {

	validate = validator.New()
}

func ValidateStruct(s any) (err error) {

	if err = validate.Struct(s); err != nil {
		var validationErrs validator.ValidationErrors
		errors.As(err, &validationErrs)
		validationErrors := make([]string, 0, len(validationErrs))
		for _, validationErr := range validationErrs {
			field := validationErr.Field()
			tag := validationErr.Tag()
			var msg string
			switch tag {
			case "required":
				msg = fmt.Sprintf("%s is required", field)
			case "min":
				msg = fmt.Sprintf("%s must be at least %s characters", field, validationErr.Param())
			case "max":
				msg = fmt.Sprintf("%s must be at most %s characters", field, validationErr.Param())
			case "url":
				msg = fmt.Sprintf("%s must be a valid URL", field)
			default:
				msg = fmt.Sprintf("%s is invalid", field)
			}
			validationErrors = append(validationErrors, msg)
		}
		return fmt.Errorf("validation failed: %s", strings.Join(validationErrors, "; "))
	}

	return
}

func ValidateAlias(alias string) (err error) {

	if alias == "" {
		return fmt.Errorf("alias is required")
	}
	if len(alias) < 1 || len(alias) > 255 {
		return fmt.Errorf("alias must be between 1 and 255 characters")
	}

	return nil
}
