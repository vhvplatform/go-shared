package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vhvcorp/go-shared/response"
)

// Timeout adds a timeout to requests
// If a request takes longer than the specified duration, it returns a timeout error
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		// Check if timeout occurred
		if ctx.Err() == context.DeadlineExceeded {
			// Only send error response if nothing was written yet
			if !c.Writer.Written() {
				response.Error(c, http.StatusGatewayTimeout, "TIMEOUT", "Request timeout exceeded")
			}
			c.Abort()
		}
	}
}

// TimeoutWithCustomMessage adds a timeout with a custom error message
func TimeoutWithCustomMessage(timeout time.Duration, message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		// Check if timeout occurred
		if ctx.Err() == context.DeadlineExceeded {
			// Only send error response if nothing was written yet
			if !c.Writer.Written() {
				response.Error(c, http.StatusGatewayTimeout, "TIMEOUT", message)
			}
			c.Abort()
		}
	}
}
