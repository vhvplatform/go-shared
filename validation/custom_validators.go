package validation

import (
	"encoding/json"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var (
	// Phone regex: supports international formats
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

	// Slug regex: lowercase letters, numbers, hyphens
	slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

	// Tenant ID regex: alphanumeric with optional hyphens
	tenantIDRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$`)

	// Hex color regex: #RGB or #RRGGBB
	hexColorRegex = regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)

	// Domain regex: basic domain validation
	domainRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

	// Username regex: alphanumeric, underscore, hyphen, 3-20 chars
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`)

	// IPv4 regex
	ipv4Regex = regexp.MustCompile(`^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}$`)

	// IPv6 regex (simplified)
	//nolint:lll // Complex regex pattern for IPv6 validation
	ipv6Regex = regexp.MustCompile(`^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$`)

	// MAC address regex
	macAddressRegex = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)

	// Semantic version regex
	//nolint:lll // Complex regex pattern for semantic version validation
	semverRegex = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

	// Credit card regex (basic format check)
	creditCardRegex = regexp.MustCompile(`^\d{13,19}$`)

	// SQL injection patterns to block (lowercase for case-insensitive matching)
	sqlInjectionPatterns = []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_",
		"exec", "execute", "select", "insert", "update", "delete",
		"drop", "create", "alter", "union", "script",
	}

	// Valid ISO 4217 currency codes (subset)
	validCurrencyCodes = map[string]bool{
		"USD": true, "EUR": true, "GBP": true, "JPY": true, "CNY": true,
		"AUD": true, "CAD": true, "CHF": true, "INR": true, "KRW": true,
		"BRL": true, "MXN": true, "RUB": true, "ZAR": true, "SGD": true,
		"HKD": true, "NOK": true, "SEK": true, "DKK": true, "PLN": true,
		"THB": true, "IDR": true, "MYR": true, "PHP": true, "VND": true,
	}

	// Valid ISO 639-1 language codes (subset)
	validLanguageCodes = map[string]bool{
		"en": true, "es": true, "fr": true, "de": true, "it": true,
		"pt": true, "ru": true, "ja": true, "ko": true, "zh": true,
		"ar": true, "hi": true, "bn": true, "pa": true, "te": true,
		"vi": true, "tr": true, "pl": true, "uk": true, "nl": true,
		"th": true, "id": true, "ms": true, "fil": true, "sv": true,
	}
)

// validatePhone validates international phone numbers
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")

	return phoneRegex.MatchString(phone)
}

// validateSlug validates URL-friendly slugs
func validateSlug(fl validator.FieldLevel) bool {
	slug := fl.Field().String()
	if len(slug) < 1 || len(slug) > 100 {
		return false
	}
	return slugRegex.MatchString(slug)
}

// validatePasswordStrong validates password strength
// Requirements:
// - At least 8 characters
// - At least one uppercase letter
// - At least one lowercase letter
// - At least one number
// - At least one special character
func validatePasswordStrong(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// validateTenantID validates tenant IDs
func validateTenantID(fl validator.FieldLevel) bool {
	tenantID := fl.Field().String()

	// Length validation
	if len(tenantID) < 3 || len(tenantID) > 50 {
		return false
	}

	// Pattern validation
	return tenantIDRegex.MatchString(tenantID)
}

// validateSafeString prevents SQL injection and XSS
func validateSafeString(fl validator.FieldLevel) bool {
	str := strings.ToLower(fl.Field().String())

	// Check for SQL injection patterns (patterns are already lowercase)
	for _, pattern := range sqlInjectionPatterns {
		if strings.Contains(str, pattern) {
			return false
		}
	}

	// Check for script tags
	if strings.Contains(str, "<script") || strings.Contains(str, "</script>") {
		return false
	}

	return true
}

// validateHexColor validates hex color codes
func validateHexColor(fl validator.FieldLevel) bool {
	color := fl.Field().String()
	return hexColorRegex.MatchString(color)
}

// validateDomain validates domain names
func validateDomain(fl validator.FieldLevel) bool {
	domain := fl.Field().String()

	// Length validation
	if len(domain) < 4 || len(domain) > 253 {
		return false
	}

	return domainRegex.MatchString(domain)
}

// validateJSONString validates if string is valid JSON
func validateJSONString(fl validator.FieldLevel) bool {
	str := fl.Field().String()

	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// validateUsername validates usernames
// Requirements: 3-20 characters, alphanumeric, underscore, hyphen
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	return usernameRegex.MatchString(username)
}

// validateIPv4 validates IPv4 addresses
func validateIPv4(fl validator.FieldLevel) bool {
	ip := fl.Field().String()
	return ipv4Regex.MatchString(ip)
}

// validateIPv6 validates IPv6 addresses
func validateIPv6(fl validator.FieldLevel) bool {
	ip := fl.Field().String()
	return ipv6Regex.MatchString(ip)
}

// validateMACAddress validates MAC addresses
func validateMACAddress(fl validator.FieldLevel) bool {
	mac := fl.Field().String()
	return macAddressRegex.MatchString(mac)
}

// validateLatitude validates latitude coordinates (-90 to 90)
func validateLatitude(fl validator.FieldLevel) bool {
	lat := fl.Field().Float()
	return lat >= -90 && lat <= 90
}

// validateLongitude validates longitude coordinates (-180 to 180)
func validateLongitude(fl validator.FieldLevel) bool {
	lon := fl.Field().Float()
	return lon >= -180 && lon <= 180
}

// validateSemver validates semantic version strings
func validateSemver(fl validator.FieldLevel) bool {
	version := fl.Field().String()
	return semverRegex.MatchString(version)
}

// validateCreditCard validates credit card numbers (basic Luhn check)
func validateCreditCard(fl validator.FieldLevel) bool {
	cardNumber := fl.Field().String()

	// Remove spaces and hyphens
	cardNumber = strings.ReplaceAll(cardNumber, " ", "")
	cardNumber = strings.ReplaceAll(cardNumber, "-", "")

	// Check basic format
	if !creditCardRegex.MatchString(cardNumber) {
		return false
	}

	// Luhn algorithm
	sum := 0
	isEven := false

	for i := len(cardNumber) - 1; i >= 0; i-- {
		digit := int(cardNumber[i] - '0')

		if isEven {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isEven = !isEven
	}

	return sum%10 == 0
}

// validateCurrencyCode validates ISO 4217 currency codes
func validateCurrencyCode(fl validator.FieldLevel) bool {
	code := strings.ToUpper(fl.Field().String())
	return validCurrencyCodes[code]
}

// validateLanguageCode validates ISO 639-1 language codes
func validateLanguageCode(fl validator.FieldLevel) bool {
	code := strings.ToLower(fl.Field().String())
	return validLanguageCodes[code]
}

// validateFilePath validates file paths (prevents path traversal)
func validateFilePath(fl validator.FieldLevel) bool {
	path := fl.Field().String()

	// Block path traversal attempts
	if strings.Contains(path, "..") {
		return false
	}

	// Block absolute paths starting with /
	if strings.HasPrefix(path, "/") {
		return false
	}

	// Block Windows drive letters
	if len(path) >= 2 && path[1] == ':' {
		return false
	}

	// Block null bytes
	if strings.Contains(path, "\x00") {
		return false
	}

	return true
}
