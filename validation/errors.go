package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve.Errors {
		messages = append(messages, err.Message)
	}
	return strings.Join(messages, "; ")
}

// FormatValidationErrors converts validator errors to user-friendly messages
func FormatValidationErrors(err error) error {
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	var errors []ValidationError

	for _, fieldErr := range validationErrors {
		errors = append(errors, ValidationError{
			Field:   fieldErr.Field(),
			Tag:     fieldErr.Tag(),
			Value:   fmt.Sprintf("%v", fieldErr.Value()),
			Message: getErrorMessage(fieldErr),
		})
	}

	return ValidationErrors{Errors: errors}
}

// getErrorMessage returns a user-friendly error message for a field error
//
//nolint:gocyclo,funlen // This function has many cases for different validation tags, which is necessary
func getErrorMessage(fe validator.FieldError) string {
	field := fe.Field()

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, fe.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, fe.Param())
	case "eq":
		return fmt.Sprintf("%s must be equal to %s", field, fe.Param())
	case "ne":
		return fmt.Sprintf("%s must not be equal to %s", field, fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, fe.Param())
	case "alpha":
		return fmt.Sprintf("%s must contain only alphabetic characters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	case "numeric":
		return fmt.Sprintf("%s must be a valid number", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "uuid4":
		return fmt.Sprintf("%s must be a valid UUID v4", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	case "phone":
		return fmt.Sprintf("%s must be a valid phone number", field)
	case "slug":
		return fmt.Sprintf("%s must be a valid slug (lowercase, alphanumeric, hyphens only)", field)
	case "password_strong":
		return fmt.Sprintf("%s must contain at least 8 characters with uppercase, lowercase, number, and special character", field)
	case "tenant_id":
		return fmt.Sprintf("%s must be a valid tenant ID (3-50 alphanumeric characters)", field)
	case "safe_string":
		return fmt.Sprintf("%s contains potentially unsafe characters", field)
	case "hex_color":
		return fmt.Sprintf("%s must be a valid hex color code", field)
	case "domain":
		return fmt.Sprintf("%s must be a valid domain name", field)
	case "json_string":
		return fmt.Sprintf("%s must be valid JSON", field)
	case "username":
		return fmt.Sprintf("%s must be 3-20 characters and contain only letters, numbers, underscore, or hyphen", field)
	case "ipv4":
		return fmt.Sprintf("%s must be a valid IPv4 address", field)
	case "ipv6":
		return fmt.Sprintf("%s must be a valid IPv6 address", field)
	case "mac_address":
		return fmt.Sprintf("%s must be a valid MAC address", field)
	case "latitude":
		return fmt.Sprintf("%s must be a valid latitude (-90 to 90)", field)
	case "longitude":
		return fmt.Sprintf("%s must be a valid longitude (-180 to 180)", field)
	case "semver":
		return fmt.Sprintf("%s must be a valid semantic version (e.g., 1.0.0)", field)
	case "credit_card":
		return fmt.Sprintf("%s must be a valid credit card number", field)
	case "currency_code":
		return fmt.Sprintf("%s must be a valid ISO 4217 currency code", field)
	case "language_code":
		return fmt.Sprintf("%s must be a valid ISO 639-1 language code", field)
	case "file_path":
		return fmt.Sprintf("%s must be a valid and safe file path", field)
	case "string_length":
		return fmt.Sprintf("%s must be between %s characters", field, fe.Param())
	case "array_length":
		return fmt.Sprintf("%s must contain between %s items", field, fe.Param())
	case "numeric_range":
		return fmt.Sprintf("%s must be between %s", field, fe.Param())
	case "alpha_numeric_spaces":
		if fe.Param() != "" {
			return fmt.Sprintf("%s must be alphanumeric with spaces and between %s characters", field, fe.Param())
		}
		return fmt.Sprintf("%s must contain only letters, numbers, and spaces", field)
	case "required_if":
		return fmt.Sprintf("%s is required when %s", field, fe.Param())
	case "required_unless":
		return fmt.Sprintf("%s is required unless %s", field, fe.Param())
	case "field_equals":
		return fmt.Sprintf("%s must equal %s", field, fe.Param())
	case "one_of":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	case "excluded":
		return fmt.Sprintf("%s cannot be: %s", field, fe.Param())
	case "contains":
		return fmt.Sprintf("%s must contain '%s'", field, fe.Param())
	case "starts_with":
		return fmt.Sprintf("%s must start with '%s'", field, fe.Param())
	case "ends_with":
		return fmt.Sprintf("%s must end with '%s'", field, fe.Param())
	default:
		return fmt.Sprintf("%s failed validation for '%s'", field, fe.Tag())
	}
}
