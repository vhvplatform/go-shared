package validation

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps go-playground validator with custom validators
type Validator struct {
	validate *validator.Validate
}

// New creates a new validator instance with custom validators registered
func New() *Validator {
	v := validator.New()

	// Use JSON tag names in error messages
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom validators
	registerCustomValidators(v)

	// Register advanced validators with parameters
	registerAdvancedValidators(v)

	return &Validator{validate: v}
}

// Validate validates a struct
func (v *Validator) Validate(i interface{}) error {
	if err := v.validate.Struct(i); err != nil {
		return FormatValidationErrors(err)
	}
	return nil
}

// ValidateVar validates a single variable
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	if err := v.validate.Var(field, tag); err != nil {
		return FormatValidationErrors(err)
	}
	return nil
}

// ValidateMap validates a map with validation rules
func (v *Validator) ValidateMap(data, rules map[string]interface{}) map[string]interface{} {
	return v.validate.ValidateMap(data, rules)
}

// registerCustomValidators registers all custom validators
func registerCustomValidators(v *validator.Validate) {
	// Phone number validator
	_ = v.RegisterValidation("phone", validatePhone)

	// Slug validator (URL-friendly string)
	_ = v.RegisterValidation("slug", validateSlug)

	// Password strength validator
	_ = v.RegisterValidation("password_strong", validatePasswordStrong)

	// Tenant ID validator
	_ = v.RegisterValidation("tenant_id", validateTenantID)

	// SQL injection prevention
	_ = v.RegisterValidation("safe_string", validateSafeString)

	// Hex color validator
	_ = v.RegisterValidation("hex_color", validateHexColor)

	// Domain name validator
	_ = v.RegisterValidation("domain", validateDomain)

	// JSON string validator
	_ = v.RegisterValidation("json_string", validateJSONString)

	// Username validator
	_ = v.RegisterValidation("username", validateUsername)

	// IP address validators
	_ = v.RegisterValidation("ipv4", validateIPv4)
	_ = v.RegisterValidation("ipv6", validateIPv6)

	// MAC address validator
	_ = v.RegisterValidation("mac_address", validateMACAddress)

	// Geolocation validators
	_ = v.RegisterValidation("latitude", validateLatitude)
	_ = v.RegisterValidation("longitude", validateLongitude)

	// Semantic version validator
	_ = v.RegisterValidation("semver", validateSemver)

	// Credit card validator
	_ = v.RegisterValidation("credit_card", validateCreditCard)

	// Currency code validator
	_ = v.RegisterValidation("currency_code", validateCurrencyCode)

	// Language code validator
	_ = v.RegisterValidation("language_code", validateLanguageCode)

	// File path validator
	_ = v.RegisterValidation("file_path", validateFilePath)
}
