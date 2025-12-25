package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	pkgctx "github.com/vhvcorp/go-shared/context"
	"github.com/vhvcorp/go-shared/jwt"
	"github.com/vhvcorp/go-shared/response"
)

// Auth validates JWT tokens and sets user context
// This middleware extracts the JWT from the Authorization header,
// validates it, and sets user information in the Gin context
func Auth(jwtSecret string) gin.HandlerFunc {
	jwtManager := jwt.NewManager(jwtSecret, 3600, 86400)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}

// TenantScope ensures tenant context exists
// This middleware checks if a tenant ID is present in the context,
// either from JWT claims or from the X-Tenant-ID header
func TenantScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			// Try to get from header
			tenantID = c.GetHeader("X-Tenant-ID")
			if tenantID == "" {
				response.BadRequest(c, "Tenant ID required")
				c.Abort()
				return
			}
			c.Set("tenant_id", tenantID)
		}

		c.Next()
	}
}

// RequireRoles checks if user has any of the required roles
// This middleware ensures the authenticated user has at least one of the specified roles
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !pkgctx.HasAnyRoleFromGin(c, roles...) {
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAllRoles checks if user has all of the required roles
// This middleware ensures the authenticated user has all of the specified roles
func RequireAllRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles := pkgctx.GetRolesFromGin(c)
		
		for _, required := range roles {
			hasRole := false
			for _, userRole := range userRoles {
				if userRole == required {
					hasRole = true
					break
				}
			}
			if !hasRole {
				response.Forbidden(c, "Insufficient permissions")
				c.Abort()
				return
			}
		}
		
		c.Next()
	}
}

// OptionalAuth is similar to Auth but doesn't require authentication
// If a valid token is provided, it sets the user context, otherwise continues without it
func OptionalAuth(jwtSecret string) gin.HandlerFunc {
	jwtManager := jwt.NewManager(jwtSecret, 3600, 86400)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token := parts[1]

		// Validate token
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			c.Next()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}
