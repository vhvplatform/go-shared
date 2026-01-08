package auth

// MultiTenantContext holds authentication context for multi-tenant systems
type MultiTenantContext struct {
	UserID      string            `json:"user_id"`
	TenantID    string            `json:"tenant_id"`
	Email       string            `json:"email"`
	Roles       []string          `json:"roles"`
	Permissions []string          `json:"permissions"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// HasRole checks if user has a specific role
func (c *MultiTenantContext) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if user has any of the specified roles
func (c *MultiTenantContext) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}

// HasAllRoles checks if user has all of the specified roles
func (c *MultiTenantContext) HasAllRoles(roles ...string) bool {
	for _, role := range roles {
		if !c.HasRole(role) {
			return false
		}
	}
	return true
}

// HasPermission checks if user has a specific permission
func (c *MultiTenantContext) HasPermission(permission string) bool {
	for _, p := range c.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if user has any of the specified permissions
func (c *MultiTenantContext) HasAnyPermission(permissions ...string) bool {
	for _, permission := range permissions {
		if c.HasPermission(permission) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if user has all of the specified permissions
func (c *MultiTenantContext) HasAllPermissions(permissions ...string) bool {
	for _, permission := range permissions {
		if !c.HasPermission(permission) {
			return false
		}
	}
	return true
}

// IsAdmin checks if user has admin role
func (c *MultiTenantContext) IsAdmin() bool {
	return c.HasAnyRole("admin", "administrator", "super_admin")
}

// IsSuperAdmin checks if user has super admin role
func (c *MultiTenantContext) IsSuperAdmin() bool {
	return c.HasRole("super_admin")
}

// TenantLoginConfig represents tenant-specific login configuration
type TenantLoginConfig struct {
	TenantID             string            `json:"tenant_id"`
	AllowedIdentifiers   []string          `json:"allowed_identifiers"`
	Require2FA           bool              `json:"require_2fa"`
	AllowRegistration    bool              `json:"allow_registration"`
	CustomLogoURL        string            `json:"custom_logo_url,omitempty"`
	CustomBackgroundURL  string            `json:"custom_background_url,omitempty"`
	CustomFields         map[string]string `json:"custom_fields,omitempty"`
	PasswordMinLength    int               `json:"password_min_length"`
	PasswordRequireUpper bool              `json:"password_require_upper"`
	PasswordRequireLower bool              `json:"password_require_lower"`
	PasswordRequireDigit bool              `json:"password_require_digit"`
	PasswordRequireSpec  bool              `json:"password_require_spec"`
	SessionTimeout       int               `json:"session_timeout"`
	MaxLoginAttempts     int               `json:"max_login_attempts"`
	LockoutDuration      int               `json:"lockout_duration"`
}

// IsIdentifierAllowed checks if an identifier type is allowed for login
func (c *TenantLoginConfig) IsIdentifierAllowed(identifierType string) bool {
	for _, allowed := range c.AllowedIdentifiers {
		if allowed == identifierType {
			return true
		}
	}
	return false
}

// GetSessionTimeoutDuration returns session timeout as duration
func (c *TenantLoginConfig) GetSessionTimeoutDuration() int {
	if c.SessionTimeout > 0 {
		return c.SessionTimeout
	}
	return 1440 // Default 24 hours
}

// GetLockoutDurationMinutes returns lockout duration in minutes
func (c *TenantLoginConfig) GetLockoutDurationMinutes() int {
	if c.LockoutDuration > 0 {
		return c.LockoutDuration
	}
	return 30 // Default 30 minutes
}

// UserTenantRelation represents a user's relationship with a tenant
type UserTenantRelation struct {
	UserID   string   `json:"user_id"`
	TenantID string   `json:"tenant_id"`
	Roles    []string `json:"roles"`
	IsActive bool     `json:"is_active"`
}
