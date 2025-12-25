package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vhvcorp/go-shared/auth"
)

// RequirePermission creates middleware that checks for specific permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		if !auth.HasPermission(ctx, permission) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":               "Insufficient permissions",
				"required_permission": permission,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission creates middleware that checks for any of the permissions
func RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	checker := auth.NewPermissionChecker()

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		if !checker.HasAnyPermission(ctx, permissions...) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":                "Insufficient permissions",
				"required_permissions": permissions,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole creates middleware that checks for specific role
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		if !auth.HasRole(ctx, role) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Insufficient role",
				"required_role": role,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSuperAdmin creates middleware that requires super admin role
func RequireSuperAdmin() gin.HandlerFunc {
	return RequireRole("super_admin")
}

// RequireTenantAdmin creates middleware that requires tenant admin role
func RequireTenantAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		if !auth.IsTenantAdmin(ctx) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Tenant admin role required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
