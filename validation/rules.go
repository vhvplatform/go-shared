package validation

// Common validation tags that can be reused across the application

const (
	// Email validation
	RuleEmail = "required,email,max=255"

	// Password validation
	RulePassword       = "required,min=8,max=128"                 //nolint:gosec // This is a validation rule, not a credential
	RulePasswordStrong = "required,min=8,max=128,password_strong" //nolint:gosec // This is a validation rule, not a credential

	// Name validation
	RuleName      = "required,min=2,max=100"
	RuleNameOpt   = "omitempty,min=2,max=100"
	RuleFirstName = "required,min=2,max=50,alpha_unicode"
	RuleLastName  = "required,min=2,max=50,alpha_unicode"

	// ID validation
	RuleID       = "required,uuid4"
	RuleIDOpt    = "omitempty,uuid4"
	RuleTenantID = "required,tenant_id"

	// Phone validation
	RulePhone    = "required,phone"
	RulePhoneOpt = "omitempty,phone"

	// URL validation
	RuleURL    = "required,url,max=2048"
	RuleURLOpt = "omitempty,url,max=2048"

	// Slug validation
	RuleSlug = "required,slug,min=3,max=100"

	// Domain validation
	RuleDomain = "required,domain"

	// Text validation
	RuleTextShort    = "required,min=1,max=255"
	RuleTextMedium   = "required,min=1,max=1000"
	RuleTextLong     = "required,min=1,max=5000"
	RuleTextOptional = "omitempty,max=1000"

	// Number validation
	RulePositiveInt    = "required,min=1"
	RuleNonNegativeInt = "required,min=0"
	RulePercentage     = "required,min=0,max=100"

	// Boolean validation
	RuleBool = "required,boolean"

	// Array validation
	RuleArrayNotEmpty = "required,min=1,dive"
	RuleArrayMax10    = "required,max=10,dive"

	// Date validation
	RuleDate     = "required,datetime=2006-01-02"
	RuleDatetime = "required,datetime=2006-01-02T15:04:05Z07:00"

	// Color validation
	RuleHexColor = "required,hex_color"

	// JSON validation
	RuleJSON = "required,json_string"

	// Safe string (prevent injection)
	RuleSafeString = "required,safe_string,max=1000"

	// Username validation
	RuleUsername = "required,username"

	// IP address validation
	RuleIPv4 = "required,ipv4"
	RuleIPv6 = "required,ipv6"

	// MAC address validation
	RuleMACAddress = "required,mac_address"

	// Geolocation validation
	RuleLatitude  = "required,latitude"
	RuleLongitude = "required,longitude"

	// Version validation
	RuleSemver = "required,semver"

	// Payment validation
	RuleCreditCard = "required,credit_card"

	// Internationalization validation
	RuleCurrencyCode = "required,currency_code"
	RuleLanguageCode = "required,language_code"

	// File path validation
	RuleFilePath = "required,file_path,max=500"

	// Advanced validation with parameters
	RuleStringLength3to100 = "required,string_length=3:100"
	RuleStringLength1to50  = "required,string_length=1:50"
	RuleArrayLength1to10   = "required,array_length=1:10"
	RuleArrayLength1to100  = "required,array_length=1:100"
	RuleNumericRange0to100 = "required,numeric_range=0:100"
	RuleNumericRange1to999 = "required,numeric_range=1:999"

	// Conditional validation examples
	// Use these as templates - replace field names and values as needed
	RuleRequiredIfActive   = "required_if=Status active"
	RuleRequiredUnlessTest = "required_unless=Environment test"

	// String validation with parameters
	RuleEmailDomain    = "required,email,ends_with=@company.com"
	RuleSecureURL      = "required,url,starts_with=https://"
	RuleAlphaNumSpaces = "required,alpha_numeric_spaces=3:50"
)

// ValidationRules defines common validation rule sets for structs
var ValidationRules = map[string]string{
	"user.email":      RuleEmail,
	"user.password":   RulePasswordStrong,
	"user.first_name": RuleFirstName,
	"user.last_name":  RuleLastName,
	"user.phone":      RulePhoneOpt,

	"tenant.name":   RuleName,
	"tenant.slug":   RuleSlug,
	"tenant.domain": RuleDomain,

	"auth.email":    RuleEmail,
	"auth.password": RulePassword,

	"pagination.page":      RulePositiveInt,
	"pagination.page_size": "required,min=1,max=100",

	"search.query": "required,min=1,max=255,safe_string",
}

// GetRule returns a validation rule by key
func GetRule(key string) string {
	if rule, exists := ValidationRules[key]; exists {
		return rule
	}
	return ""
}
