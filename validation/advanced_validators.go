package validation

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidatorFunc is a custom validation function type
type ValidatorFunc func(value interface{}) bool

// ValidatorFuncWithParam is a custom validation function with parameter
type ValidatorFuncWithParam func(value interface{}, param string) bool

// ConditionalValidator allows conditional validation based on other fields
type ConditionalValidator struct {
	Condition func(interface{}) bool
	Rules     string
}

// validateStringLength validates string length with min and max parameters
// Usage: validate:"string_length=3:100" (min 3, max 100)
func validateStringLength(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	param := fl.Param()

	// Parse min:max from parameter
	parts := strings.Split(param, ":")
	if len(parts) != 2 {
		return false
	}

	minLen, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}

	maxLen, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}

	length := len(str)
	return length >= minLen && length <= maxLen
}

// validateArrayLength validates array/slice length with min and max parameters
// Usage: validate:"array_length=1:10" (min 1 item, max 10 items)
func validateArrayLength(fl validator.FieldLevel) bool {
	field := fl.Field()

	if field.Kind() != reflect.Slice && field.Kind() != reflect.Array {
		return false
	}

	param := fl.Param()
	parts := strings.Split(param, ":")
	if len(parts) != 2 {
		return false
	}

	minLen, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}

	maxLen, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}

	length := field.Len()
	return length >= minLen && length <= maxLen
}

// validateNumericRange validates numeric values within a range
// Usage: validate:"numeric_range=1:100" (between 1 and 100)
func validateNumericRange(fl validator.FieldLevel) bool {
	field := fl.Field()
	param := fl.Param()

	parts := strings.Split(param, ":")
	if len(parts) != 2 {
		return false
	}

	minVal, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return false
	}

	maxVal, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return false
	}

	var value float64
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = float64(field.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value = float64(field.Uint())
	case reflect.Float32, reflect.Float64:
		value = field.Float()
	default:
		return false
	}

	return value >= minVal && value <= maxVal
}

// validateRequiredIf validates field is required if another field has specific value
// Usage: validate:"required_if=Status active" (required if Status == "active")
func validateRequiredIf(fl validator.FieldLevel) bool {
	param := fl.Param()
	parts := strings.Fields(param)
	if len(parts) < 2 {
		return false
	}

	fieldName := parts[0]
	expectedValue := strings.Join(parts[1:], " ")

	// Get the parent struct
	parent := fl.Parent()
	if parent.Kind() != reflect.Struct {
		return false
	}

	// Get the condition field
	conditionField := parent.FieldByName(fieldName)
	if !conditionField.IsValid() {
		return false
	}

	// Check if condition is met
	conditionMet := fmt.Sprintf("%v", conditionField.Interface()) == expectedValue

	// If condition is met, field must not be empty
	if conditionMet {
		field := fl.Field()
		return !isFieldEmpty(field)
	}

	return true
}

// validateRequiredUnless validates field is required unless another field has specific value
// Usage: validate:"required_unless=Status inactive" (required unless Status == "inactive")
func validateRequiredUnless(fl validator.FieldLevel) bool {
	param := fl.Param()
	parts := strings.Fields(param)
	if len(parts) < 2 {
		return false
	}

	fieldName := parts[0]
	expectedValue := strings.Join(parts[1:], " ")

	// Get the parent struct
	parent := fl.Parent()
	if parent.Kind() != reflect.Struct {
		return false
	}

	// Get the condition field
	conditionField := parent.FieldByName(fieldName)
	if !conditionField.IsValid() {
		return false
	}

	// Check if condition is met
	conditionMet := fmt.Sprintf("%v", conditionField.Interface()) == expectedValue

	// If condition is not met, field must not be empty
	if !conditionMet {
		field := fl.Field()
		return !isFieldEmpty(field)
	}

	return true
}

// validateFieldEquals validates that field equals another field's value
// Usage: validate:"field_equals=Password" (must equal Password field)
func validateFieldEquals(fl validator.FieldLevel) bool {
	fieldName := fl.Param()

	// Get the parent struct
	parent := fl.Parent()
	if parent.Kind() != reflect.Struct {
		return false
	}

	// Get the comparison field
	compareField := parent.FieldByName(fieldName)
	if !compareField.IsValid() {
		return false
	}

	// Compare values
	return fl.Field().Interface() == compareField.Interface()
}

// validateOneOf validates that field value is one of the specified values
// Usage: validate:"one_of=admin user guest" (must be one of: admin, user, guest)
func validateOneOf(fl validator.FieldLevel) bool {
	value := fmt.Sprintf("%v", fl.Field().Interface())
	param := fl.Param()

	allowedValues := strings.Fields(param)
	for _, allowed := range allowedValues {
		if value == allowed {
			return true
		}
	}

	return false
}

// validateExcluded validates that field value is not one of the specified values
// Usage: validate:"excluded=admin root system" (cannot be: admin, root, system)
func validateExcluded(fl validator.FieldLevel) bool {
	value := fmt.Sprintf("%v", fl.Field().Interface())
	param := fl.Param()

	excludedValues := strings.Fields(param)
	for _, excluded := range excludedValues {
		if value == excluded {
			return false
		}
	}

	return true
}

// validateAlphaNumericSpaces validates alphanumeric characters with spaces
// Can accept parameter for min:max length
// Usage: validate:"alpha_numeric_spaces=3:50"
func validateAlphaNumericSpaces(fl validator.FieldLevel) bool {
	str := fl.Field().String()

	// Check characters
	for _, char := range str {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == ' ') {
			return false
		}
	}

	// Check length if parameter provided
	param := fl.Param()
	if param != "" {
		parts := strings.Split(param, ":")
		if len(parts) == 2 {
			minLen, err1 := strconv.Atoi(parts[0])
			maxLen, err2 := strconv.Atoi(parts[1])
			if err1 == nil && err2 == nil {
				length := len(str)
				return length >= minLen && length <= maxLen
			}
		}
	}

	return true
}

// validateContains validates that string contains a specific substring
// Usage: validate:"contains=@example.com"
func validateContains(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	param := fl.Param()
	return strings.Contains(str, param)
}

// validateStartsWith validates that string starts with a specific prefix
// Usage: validate:"starts_with=https://"
func validateStartsWith(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	param := fl.Param()
	return strings.HasPrefix(str, param)
}

// validateEndsWith validates that string ends with a specific suffix
// Usage: validate:"ends_with=.com"
func validateEndsWith(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	param := fl.Param()
	return strings.HasSuffix(str, param)
}

// isFieldEmpty checks if a field is empty
func isFieldEmpty(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.String:
		return field.String() == ""
	case reflect.Slice, reflect.Map, reflect.Array:
		return field.Len() == 0
	case reflect.Bool:
		return !field.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return field.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return field.IsNil()
	}
	return false
}

// registerAdvancedValidators registers advanced validators with parameters
func registerAdvancedValidators(v *validator.Validate) {
	// Parameterized validators
	_ = v.RegisterValidation("string_length", validateStringLength)
	_ = v.RegisterValidation("array_length", validateArrayLength)
	_ = v.RegisterValidation("numeric_range", validateNumericRange)
	_ = v.RegisterValidation("alpha_numeric_spaces", validateAlphaNumericSpaces)

	// Conditional validators
	_ = v.RegisterValidation("required_if", validateRequiredIf)
	_ = v.RegisterValidation("required_unless", validateRequiredUnless)

	// Cross-field validators
	_ = v.RegisterValidation("field_equals", validateFieldEquals)

	// Value validators with parameters
	_ = v.RegisterValidation("one_of", validateOneOf)
	_ = v.RegisterValidation("excluded", validateExcluded)

	// String validators with parameters
	_ = v.RegisterValidation("contains", validateContains)
	_ = v.RegisterValidation("starts_with", validateStartsWith)
	_ = v.RegisterValidation("ends_with", validateEndsWith)
}
