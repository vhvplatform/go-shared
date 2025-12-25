# Validation Package

Comprehensive validation package for the SaaS Framework using [go-playground/validator](https://github.com/go-playground/validator).

## Features

- ✅ Built on industry-standard `go-playground/validator`
- ✅ 18+ custom validators for common SaaS scenarios
- ✅ **Advanced validators with parameters** (string_length, numeric_range, etc.)
- ✅ **Conditional validation** (required_if, required_unless)
- ✅ **Cross-field validation** (field_equals)
- ✅ **Function-based validation** with custom parameters
- ✅ User-friendly error messages
- ✅ Reusable validation rules
- ✅ SQL injection and XSS prevention
- ✅ International format support (phone, currency, language)
- ✅ Network validation (IP, MAC, domain)
- ✅ Security validation (file path, safe string)
- ✅ Payment validation (credit card with Luhn)
- ✅ Thread-safe and performant

## Installation

```go
import "github.com/vhvcorp/go-shared/validation"
```

## Quick Start

### Basic Usage

```go
validator := validation.New()

type User struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,password_strong"`
    Phone    string `json:"phone" validate:"omitempty,phone"`
}

user := User{
    Email:    "user@example.com",
    Password: "Weak123",
    Phone:    "+1234567890",
}

if err := validator.Validate(user); err != nil {
    // Handle validation errors
    fmt.Println(err)
}
```

### Using Predefined Rules

```go
type LoginRequest struct {
    Email    string `json:"email" validate:"required,email,max=255"`
    Password string `json:"password" validate:"required,min=8,max=128"`
}

// Or use constants
type User struct {
    Email string `json:"email" validate:"required,email"`
}
```

## Custom Validators

### Phone Number

Validates international phone numbers.

```go
type Contact struct {
    Phone string `json:"phone" validate:"phone"`
}
```

### Slug

Validates URL-friendly slugs (lowercase, alphanumeric, hyphens).

```go
type Tenant struct {
    Slug string `json:"slug" validate:"slug"`
}
// Valid: "my-company", "acme-corp-123"
// Invalid: "My-Company", "acme_corp", "ACME"
```

### Strong Password

Requires:
- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one number
- At least one special character

```go
type User struct {
    Password string `json:"password" validate:"password_strong"`
}
// Valid: "MyP@ssw0rd"
// Invalid: "password", "Password1", "PASSWORD!"
```

### Tenant ID

Validates tenant identifiers (3-50 alphanumeric with hyphens).

```go
type Request struct {
    TenantID string `json:"tenant_id" validate:"tenant_id"`
}
// Valid: "tenant-123", "acme", "my-tenant-id"
// Invalid: "ab", "tenant_id", "-tenant-"
```

### Safe String

Prevents SQL injection and XSS attacks.

```go
type SearchRequest struct {
    Query string `json:"query" validate:"safe_string"`
}
// Blocks: SQL keywords, script tags, special characters
```

### Hex Color

Validates hex color codes.

```go
type Theme struct {
    PrimaryColor string `json:"primary_color" validate:"hex_color"`
}
// Valid: "#FF5733", "#f09"
// Invalid: "red", "FF5733", "#GGGGGG"
```

### Domain Name

Validates domain names.

```go
type Tenant struct {
    Domain string `json:"domain" validate:"domain"`
}
// Valid: "example.com", "sub.example.co.uk"
// Invalid: "-example.com", "example..com"
```

### JSON String

Validates JSON string format.

```go
type Config struct {
    Settings string `json:"settings" validate:"json_string"`
}
```

### Username

Validates usernames (3-20 characters, alphanumeric, underscore, hyphen).

```go
type User struct {
    Username string `json:"username" validate:"username"`
}
// Valid: "john_doe", "user-123", "alice2024"
// Invalid: "ab", "user@name", "a-very-long-username-here"
```

### IPv4 Address

Validates IPv4 addresses.

```go
type Server struct {
    IPAddress string `json:"ip_address" validate:"ipv4"`
}
// Valid: "192.168.1.1", "10.0.0.1"
// Invalid: "256.1.1.1", "192.168.1"
```

### IPv6 Address

Validates IPv6 addresses.

```go
type Server struct {
    IPAddress string `json:"ip_address" validate:"ipv6"`
}
// Valid: "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
// Invalid: "192.168.1.1", "gggg::1"
```

### MAC Address

Validates MAC addresses.

```go
type Device struct {
    MACAddress string `json:"mac_address" validate:"mac_address"`
}
// Valid: "00:1B:44:11:3A:B7", "00-1B-44-11-3A-B7"
// Invalid: "00:1B:44:11:3A", "GG:1B:44:11:3A:B7"
```

### Latitude/Longitude

Validates geographic coordinates.

```go
type Location struct {
    Latitude  float64 `json:"latitude" validate:"latitude"`
    Longitude float64 `json:"longitude" validate:"longitude"`
}
// Latitude: -90 to 90
// Longitude: -180 to 180
```

### Semantic Version

Validates semantic version strings.

```go
type Release struct {
    Version string `json:"version" validate:"semver"`
}
// Valid: "1.0.0", "2.1.3-alpha", "1.0.0-beta.1+build.123"
// Invalid: "1.0", "v1.0.0", "1.0.0.0"
```

### Credit Card

Validates credit card numbers using Luhn algorithm.

```go
type Payment struct {
    CardNumber string `json:"card_number" validate:"credit_card"`
}
// Validates format and checksum
// Supports spaces and hyphens: "4111-1111-1111-1111"
```

### Currency Code

Validates ISO 4217 currency codes.

```go
type Price struct {
    Currency string `json:"currency" validate:"currency_code"`
}
// Valid: "USD", "EUR", "GBP", "JPY"
// Invalid: "US", "DOLLAR", "usd" (case-insensitive)
```

### Language Code

Validates ISO 639-1 language codes.

```go
type Content struct {
    Language string `json:"language" validate:"language_code"`
}
// Valid: "en", "es", "fr", "de"
// Invalid: "eng", "EN", "english"
```

### File Path

Validates file paths and prevents path traversal attacks.

```go
type Upload struct {
    FilePath string `json:"file_path" validate:"file_path"`
}
// Valid: "documents/file.pdf", "images/photo.jpg"
// Invalid: "../etc/passwd", "/absolute/path", "C:\Windows"
```

## Predefined Rules

Use constants from `rules.go`:

```go
const (
    RuleEmail           = "required,email,max=255"
    RulePasswordStrong  = "required,min=8,max=128,password_strong"
    RuleName            = "required,min=2,max=100"
    RulePhone           = "required,phone"
    RuleSlug            = "required,slug,min=3,max=100"
    // ... and many more
)
```

## Error Handling

### Formatted Errors

```go
err := validator.Validate(user)
if err != nil {
    if validationErrs, ok := err.(validation.ValidationErrors); ok {
        for _, fieldErr := range validationErrs.Errors {
            fmt.Printf("Field: %s, Error: %s\n", fieldErr.Field, fieldErr.Message)
        }
    }
}
```

### JSON Error Response

```go
{
  "errors": [
    {
      "field": "email",
      "tag": "email",
      "value": "invalid-email",
      "message": "email must be a valid email address"
    },
    {
      "field": "password",
      "tag": "password_strong",
      "message": "password must contain at least 8 characters with uppercase, lowercase, number, and special character"
    }
  ]
}
```

## Integration with HTTP Handlers

### Gin Framework

```go
func CreateUser(c *gin.Context) {
    var req CreateUserRequest
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    validator := validation.New()
    if err := validator.Validate(req); err != nil {
        c.JSON(400, err)
        return
    }
    
    // Process valid request...
}
```

## Best Practices

1. **Struct Tags**: Always use JSON tags that match your API field names
2. **Reuse Rules**: Use predefined rule constants for consistency
3. **Custom Validators**: Create custom validators for domain-specific logic
4. **Error Messages**: Provide clear, actionable error messages
5. **Performance**: Reuse validator instances (they're thread-safe)

## Testing

```bash
go test ./pkg/validation/...
go test -cover ./pkg/validation/...
```

## Examples

See `tests/unit/pkg/validation/` for comprehensive examples.

## Advanced Validators with Parameters

### String Length

Validates string length with min and max parameters.

```go
type Product struct {
    Name string `json:"name" validate:"string_length=3:100"`
}
// Valid: strings between 3 and 100 characters
// Invalid: "ab" (too short), strings > 100 chars
```

### Array Length

Validates array/slice length with min and max parameters.

```go
type Cart struct {
    Items []string `json:"items" validate:"array_length=1:10"`
}
// Valid: arrays with 1 to 10 items
// Invalid: empty array, array with >10 items
```

### Numeric Range

Validates numeric values within a specified range.

```go
type Score struct {
    Value int `json:"value" validate:"numeric_range=0:100"`
}
// Valid: numbers from 0 to 100 (inclusive)
// Invalid: -1, 101
```

### Alpha Numeric with Spaces

Validates alphanumeric characters with spaces and optional length.

```go
type Title struct {
    Name string `json:"name" validate:"alpha_numeric_spaces=3:50"`
}
// Valid: "Product Name 123"
// Invalid: "Product@Name", "AB" (too short)
```

## Conditional Validators

### Required If

Field is required when another field has a specific value.

```go
type Account struct {
    Status      string `json:"status"`
    ActivatedAt string `json:"activated_at" validate:"required_if=Status active"`
}
// ActivatedAt is required only when Status is "active"
```

### Required Unless

Field is required unless another field has a specific value.

```go
type User struct {
    Environment string `json:"environment"`
    APIKey      string `json:"api_key" validate:"required_unless=Environment test"`
}
// APIKey is required unless Environment is "test"
```

## Cross-Field Validators

### Field Equals

Validates that a field equals another field's value.

```go
type Registration struct {
    Password        string `json:"password"`
    PasswordConfirm string `json:"password_confirm" validate:"field_equals=Password"`
}
// PasswordConfirm must equal Password
```

## Value Validators with Parameters

### One Of

Validates that field value is one of the specified values.

```go
type Task struct {
    Priority string `json:"priority" validate:"one_of=low medium high"`
}
// Valid: "low", "medium", "high"
// Invalid: "urgent", ""
```

### Excluded

Validates that field value is not one of the specified values.

```go
type Username struct {
    Name string `json:"name" validate:"excluded=admin root system"`
}
// Invalid: "admin", "root", "system"
// Valid: any other value
```

## String Validators with Parameters

### Contains

Validates that string contains a specific substring.

```go
type Email struct {
    Address string `json:"address" validate:"contains=@company.com"`
}
// Valid: "user@company.com"
// Invalid: "user@other.com"
```

### Starts With

Validates that string starts with a specific prefix.

```go
type SecureURL struct {
    URL string `json:"url" validate:"starts_with=https://"`
}
// Valid: "https://example.com"
// Invalid: "http://example.com"
```

### Ends With

Validates that string ends with a specific suffix.

```go
type Filename struct {
    Name string `json:"name" validate:"ends_with=.pdf"`
}
// Valid: "document.pdf"
// Invalid: "document.doc"
```

## Complex Validation Examples

### Multiple Conditions

```go
type Order struct {
    PaymentMethod string `json:"payment_method" validate:"one_of=card bank cash"`
    CardNumber    string `json:"card_number" validate:"required_if=PaymentMethod card,credit_card"`
    BankAccount   string `json:"bank_account" validate:"required_if=PaymentMethod bank"`
}
```

### Combined Validators

```go
type User struct {
    Username string `json:"username" validate:"required,username,excluded=admin root"`
    Email    string `json:"email" validate:"required,email,ends_with=@company.com"`
    Age      int    `json:"age" validate:"required,numeric_range=18:120"`
    Bio      string `json:"bio" validate:"omitempty,string_length=10:500"`
}
```

### Conditional with Parameters

```go
type Product struct {
    Type        string   `json:"type" validate:"one_of=digital physical"`
    Weight      float64  `json:"weight" validate:"required_if=Type physical,numeric_range=0.1:100"`
    DownloadURL string   `json:"download_url" validate:"required_if=Type digital,url,starts_with=https://"`
    Tags        []string `json:"tags" validate:"required,array_length=1:10,dive,string_length=2:30"`
}
```
