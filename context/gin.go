package context

import (
	"context"

	"github.com/gin-gonic/gin"
)

// GinContextKey is the key for storing context in gin.Context
const GinContextKey = "request_context"

// ToGinContext stores request context in gin.Context
// Performance: Also cache the full RequestContext object for efficient retrieval
func ToGinContext(c *gin.Context, rc *RequestContext) {
	c.Set("user_id", rc.UserID)
	c.Set("tenant_id", rc.TenantID)
	c.Set("app_id", rc.AppID)
	c.Set("email", rc.Email)
	c.Set("roles", rc.Roles)
	c.Set("permissions", rc.Permissions)
	c.Set("correlation_id", rc.CorrelationID)
	c.Set("tenant_domain", rc.TenantDomain)
	// Cache the full RequestContext to avoid rebuilding it
	c.Set(GinContextKey, rc)
}

// FromGinContext retrieves request context from gin.Context
// Performance: Return cached RequestContext if available
func FromGinContext(c *gin.Context) *RequestContext {
	// Try to get cached context first (performance optimization)
	if cached, exists := c.Get(GinContextKey); exists {
		if rc, ok := cached.(*RequestContext); ok && rc != nil {
			return rc
		}
	}
	
	// Fallback to building from individual values
	return &RequestContext{
		UserID:        c.GetString("user_id"),
		TenantID:      c.GetString("tenant_id"),
		AppID:         c.GetString("app_id"),
		Email:         c.GetString("email"),
		Roles:         getStringSlice(c, "roles"),
		Permissions:   getStringSlice(c, "permissions"),
		CorrelationID: c.GetString("correlation_id"),
		TenantDomain:  c.GetString("tenant_domain"),
	}
}

// GinToStdContext converts gin.Context to standard context.Context with request context
func GinToStdContext(c *gin.Context) context.Context {
	rc := FromGinContext(c)
	return WithRequestContext(c.Request.Context(), rc)
}

func getStringSlice(c *gin.Context, key string) []string {
	value, exists := c.Get(key)
	if !exists {
		return []string{}
	}
	slice, ok := value.([]string)
	if !ok {
		return []string{}
	}
	return slice
}

// GetUserIDFromGin retrieves user ID from gin context
func GetUserIDFromGin(c *gin.Context) string {
	return c.GetString("user_id")
}

// GetTenantIDFromGin retrieves tenant ID from gin context
func GetTenantIDFromGin(c *gin.Context) string {
	return c.GetString("tenant_id")
}

// GetAppIDFromGin retrieves app ID from gin context
func GetAppIDFromGin(c *gin.Context) string {
	return c.GetString("app_id")
}

// GetRolesFromGin retrieves roles from gin context
func GetRolesFromGin(c *gin.Context) []string {
	return getStringSlice(c, "roles")
}

// GetPermissionsFromGin retrieves permissions from gin context
func GetPermissionsFromGin(c *gin.Context) []string {
	return getStringSlice(c, "permissions")
}

// GetEmailFromGin retrieves email from gin context
func GetEmailFromGin(c *gin.Context) string {
	return c.GetString("email")
}

// GetCorrelationIDFromGin retrieves correlation ID from gin context
func GetCorrelationIDFromGin(c *gin.Context) string {
	return c.GetString("correlation_id")
}

// HasRoleFromGin checks if user has a specific role in gin context
func HasRoleFromGin(c *gin.Context, role string) bool {
	roles := GetRolesFromGin(c)
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRoleFromGin checks if user has any of the specified roles in gin context
func HasAnyRoleFromGin(c *gin.Context, roles ...string) bool {
	userRoles := GetRolesFromGin(c)
	for _, role := range roles {
		for _, userRole := range userRoles {
			if userRole == role {
				return true
			}
		}
	}
	return false
}

// HasAllRolesFromGin checks if user has all of the specified roles in gin context
func HasAllRolesFromGin(c *gin.Context, roles ...string) bool {
	userRoles := GetRolesFromGin(c)
	for _, role := range roles {
		found := false
		for _, userRole := range userRoles {
			if userRole == role {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// UserContext is an alias for RequestContext for backward compatibility
type UserContext = RequestContext

// GetUserContext extracts user context from Gin context
func GetUserContext(c *gin.Context) (*UserContext, error) {
	rc := FromGinContext(c)
	if rc.UserID == "" {
		return nil, ErrUserNotFound
	}
	return rc, nil
}

// SetUserContext sets user context in Gin context
func SetUserContext(c *gin.Context, user *UserContext) {
	ToGinContext(c, user)
}

// HasRole checks if user has specific role in Gin context
func HasRole(c *gin.Context, role string) bool {
	return HasRoleFromGin(c, role)
}

// HasAnyRole checks if user has any of the roles in Gin context
func HasAnyRole(c *gin.Context, roles ...string) bool {
	return HasAnyRoleFromGin(c, roles...)
}
