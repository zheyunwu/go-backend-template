package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// CustomValidator holds the validator instance
type CustomValidator struct {
	validate *validator.Validate
}

// NewCustomValidator creates a new CustomValidator
func NewCustomValidator() *CustomValidator {
	v := validator.New()

	// Register custom validation functions
	registerCustomValidators(v)

	return &CustomValidator{
		validate: v,
	}
}

// registerCustomValidators registers custom validation functions with the validator instance.
func registerCustomValidators(v *validator.Validate) {
	// Register empty_or_url validator
	v.RegisterValidation("empty_or_url", validateEmptyOrURL)
	v.RegisterValidation("empty_or_e164", validateEmptyOrE164)
}

// validateEmptyOrURL
func validateEmptyOrURL(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// Allow empty strings
	if value == "" {
		return true
	}
	// Validate if it's a valid URL
	_, err := url.ParseRequestURI(value)
	return err == nil
}

func validateEmptyOrE164(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// Allow empty strings
	if value == "" {
		return true
	}
	// Validate if it's a valid E.164 phone number
	return IsValidPhoneNumber(value)
}

func IsValidPhoneNumber(phone_number string) bool {
	e164Regex := `^\+[1-9]\d{1,14}$`
	re := regexp.MustCompile(e164Regex)
	phone_number = strings.ReplaceAll(phone_number, " ", "")

	return re.Find([]byte(phone_number)) != nil
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
		message := getCustomErrorMessage(fieldErr)
		errorMessages[field] = message
	}

	return errorMessages
}

// getCustomErrorMessage generates a custom error message based on the validation tag.
func getCustomErrorMessage(fieldErr validator.FieldError) string {
	field := fieldErr.Field()
	tag := fieldErr.Tag()

	switch tag {
	case "empty_or_url":
		return fmt.Sprintf("Field '%s' must be empty or a valid URL", field)
	case "empty_or_e164":
		return fmt.Sprintf("Field '%s' must be empty or a valid phone number in E.164 format", field)
	case "required":
		return fmt.Sprintf("Field '%s' is required", field)
	case "email":
		return fmt.Sprintf("Field '%s' must be a valid email address", field)
	case "min":
		return fmt.Sprintf("Field '%s' must be at least %s characters long", field, fieldErr.Param())
	case "max":
		return fmt.Sprintf("Field '%s' must be at most %s characters long", field, fieldErr.Param())
	case "url":
		return fmt.Sprintf("Field '%s' must be a valid URL", field)
	case "e164":
		return fmt.Sprintf("Field '%s' must be a valid phone number in E.164 format", field)
	case "datetime":
		return fmt.Sprintf("Field '%s' must be a valid date in format %s", field, fieldErr.Param())
	case "oneof":
		return fmt.Sprintf("Field '%s' must be one of: %s", field, fieldErr.Param())
	default:
		return fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", field, tag)
	}
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
