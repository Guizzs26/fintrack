package validatorx

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// FieldError contains structured information about a single validation error
// This structure is designed to be returned to the API client
type FieldError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

// ValidationError is a custom error type that wraps one or more FieldErrors
// This allows us to return all validation failures at once
type ValidationError struct {
	Errors []FieldError
}

// Error implements the error interface for ValidationError
func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation failed with %d error(s)", len(ve.Errors))
}

// Validator is a custom validator for Echo that uses the go-playground/validator library
type Validator struct {
	validator *validator.Validate
}

// NewValidator creates a new instance of Validator
func NewValidator() *Validator {
	return &Validator{validator: validator.New()}
}

// Validate implements the echo.Validator interface
// It performs struct validation and, if it fails, returns a custom ValidationError
// containing detailed information about each field error
func (v *Validator) Validate(i any) error {
	err := v.validator.Struct(i)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		out := ValidationError{
			Errors: make([]FieldError, len(validationErrors)),
		}

		for i, fe := range validationErrors {
			out.Errors[i] = FieldError{
				Field:   fe.Field(),
				Tag:     fe.Tag(),
				Message: msgForTag(fe.Tag(), fe.Param()),
			}
		}
		return out
	}
	return err
}

func msgForTag(tag, param string) string {
	switch tag {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("This field must be at least %s characters long", param)
	case "max":
		return fmt.Sprintf("This field must not exceed %s characters", param)
	default:
		return fmt.Sprintf("Failed validation on rule: %s", tag)
	}
}
