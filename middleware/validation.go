package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vhvcorp/go-shared/response"
)

// RequestSizeLimit limits the size of incoming requests
func RequestSizeLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

// RequestValidation validates incoming requests
// It ensures Content-Type header is present for non-GET/DELETE requests
func RequestValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate required headers for non-GET/DELETE requests
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodDelete {
			contentType := c.GetHeader("Content-Type")
			if contentType == "" {
				response.BadRequest(c, "Content-Type header required")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// RequireContentType validates that the request has the specified content type
func RequireContentType(contentType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodDelete {
			ct := c.GetHeader("Content-Type")
			if ct != contentType {
				response.BadRequest(c, "Invalid Content-Type header")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// RequireJSON ensures the request has JSON content type
func RequireJSON() gin.HandlerFunc {
	return RequireContentType("application/json")
}
