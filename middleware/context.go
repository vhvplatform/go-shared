package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	pkgctx "github.com/vhvcorp/go-shared/context"
)

// ContextMiddleware enriches request with context information
func ContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate correlation ID if not present
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		c.Set("correlation_id", correlationID)
		c.Header("X-Correlation-ID", correlationID)

		c.Next()
	}
}

// AppContextMiddleware extracts application context from request
func AppContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract app ID from header or query param
		appID := c.GetHeader("X-App-ID")
		if appID == "" {
			appID = c.Query("app_id")
		}

		if appID != "" {
			c.Set("app_id", appID)
		}

		c.Next()
	}
}

// RequestContextMiddleware builds full request context after auth
func RequestContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This should be used after auth middleware
		rc := pkgctx.FromGinContext(c)

		// Store in gin context
		pkgctx.ToGinContext(c, rc)

		// Also create standard context
		ctx := pkgctx.WithRequestContext(c.Request.Context(), rc)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
