package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// GenerateRandomToken generates a cryptographically secure random token
func GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateOpaqueToken generates an opaque access token
func GenerateOpaqueToken() (string, error) {
	return GenerateRandomToken(32)
}

// ValidateEmail validates an email address
func ValidateEmail(email string) bool {
	if email == "" {
		return false
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidatePhone validates a phone number (basic validation)
func ValidatePhone(phone string) bool {
	if phone == "" {
		return false
	}
	// Remove common phone number characters
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	cleaned = strings.ReplaceAll(cleaned, "+", "")

	// Check if remaining characters are digits and length is reasonable
	phoneRegex := regexp.MustCompile(`^\d{7,15}$`)
	return phoneRegex.MatchString(cleaned)
}

// ValidateUsername validates a username
func ValidateUsername(username string) bool {
	if username == "" {
		return false
	}
	// Username should be 3-30 characters, alphanumeric, can contain underscore and dash
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)
	return usernameRegex.MatchString(username)
}

// ValidateDocumentNumber validates a document number (basic validation)
func ValidateDocumentNumber(docNumber string) bool {
	if docNumber == "" {
		return false
	}
	// Document numbers are typically alphanumeric, 5-20 characters
	docRegex := regexp.MustCompile(`^[a-zA-Z0-9-]{5,20}$`)
	return docRegex.MatchString(docNumber)
}

// PasswordStrength calculates password strength (0-4)
// 0 = very weak, 1 = weak, 2 = medium, 3 = strong, 4 = very strong
func PasswordStrength(password string) int {
	strength := 0

	// Length check
	if len(password) >= 8 {
		strength++
	}
	if len(password) >= 12 {
		strength++
	}

	// Character variety checks
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	charVariety := 0
	if hasLower {
		charVariety++
	}
	if hasUpper {
		charVariety++
	}
	if hasDigit {
		charVariety++
	}
	if hasSpecial {
		charVariety++
	}

	// Add points based on character variety
	if charVariety >= 3 {
		strength++
	}
	if charVariety == 4 {
		strength++
	}

	return strength
}

// ContainsUppercase checks if string contains uppercase letters
func ContainsUppercase(s string) bool {
	for _, char := range s {
		if unicode.IsUpper(char) {
			return true
		}
	}
	return false
}

// ContainsLowercase checks if string contains lowercase letters
func ContainsLowercase(s string) bool {
	for _, char := range s {
		if unicode.IsLower(char) {
			return true
		}
	}
	return false
}

// ContainsDigit checks if string contains digits
func ContainsDigit(s string) bool {
	for _, char := range s {
		if unicode.IsDigit(char) {
			return true
		}
	}
	return false
}

// ContainsSpecialChar checks if string contains special characters
func ContainsSpecialChar(s string) bool {
	for _, char := range s {
		if unicode.IsPunct(char) || unicode.IsSymbol(char) {
			return true
		}
	}
	return false
}

// SanitizeIdentifier removes potentially dangerous characters from identifiers
func SanitizeIdentifier(identifier string) string {
	// Remove leading/trailing whitespace
	identifier = strings.TrimSpace(identifier)
	// Convert to lowercase for email
	if strings.Contains(identifier, "@") {
		identifier = strings.ToLower(identifier)
	}
	return identifier
}

// DetectIdentifierType attempts to detect the type of identifier
func DetectIdentifierType(identifier string) string {
	identifier = strings.TrimSpace(identifier)

	if ValidateEmail(identifier) {
		return "email"
	}
	if ValidatePhone(identifier) {
		return "phone"
	}
	if ValidateUsername(identifier) {
		return "username"
	}
	if ValidateDocumentNumber(identifier) {
		return "document_number"
	}

	return "unknown"
}

// NormalizeIdentifier normalizes an identifier for storage
func NormalizeIdentifier(identifier, identifierType string) string {
	identifier = strings.TrimSpace(identifier)

	switch identifierType {
	case "email":
		return strings.ToLower(identifier)
	case "phone":
		// Remove all non-digit characters except leading +
		normalized := ""
		for i, char := range identifier {
			if unicode.IsDigit(char) || (i == 0 && char == '+') {
				normalized += string(char)
			}
		}
		return normalized
	case "username":
		return strings.ToLower(identifier)
	case "document_number":
		return strings.ToUpper(identifier)
	default:
		return identifier
	}
}
