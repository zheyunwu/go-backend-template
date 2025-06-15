package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// CustomValidator holds the validator instance
type CustomValidator struct {
	validate *validator.Validate
}

// NewCustomValidator creates a new CustomValidator
func NewCustomValidator() *CustomValidator {
	return &CustomValidator{
		validate: validator.New(),
	}
}

// ValidateStruct validates a struct and returns a map of field to error message
// or nil if validation passes.
func (cv *CustomValidator) ValidateStruct(payload interface{}) map[string]string {
	err := cv.validate.Struct(payload)
	if err == nil {
		return nil
	}

	validationErrors := err.(validator.ValidationErrors)
	errorMessages := make(map[string]string)

	for _, fieldErr := range validationErrors {
		field := fieldErr.Field()
		// Use a more descriptive error message if possible
		// For now, using a generic message based on the tag
		errorMessages[field] = fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", field, fieldErr.Tag())
	}

	return errorMessages
}

// FormatValidationErrors formats validation errors into a single string.
// This can be used for logging or simple error responses.
func FormatValidationErrors(errors map[string]string) string {
	if errors == nil {
		return ""
	}
	var errorStrings []string
	for field, message := range errors {
		errorStrings = append(errorStrings, fmt.Sprintf("%s: %s", field, message))
	}
	return strings.Join(errorStrings, "; ")
}
