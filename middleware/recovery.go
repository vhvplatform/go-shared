package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vhvcorp/go-shared/logger"
	"github.com/vhvcorp/go-shared/response"
	"go.uber.org/zap"
)

// Recovery provides panic recovery with proper logging
// It catches panics, logs them, and returns a standardized error response
func Recovery(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("correlation_id", c.GetString("correlation_id")),
				)

				response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal server error occurred")
				c.Abort()
			}
		}()
		c.Next()
	}
}

// RecoveryWithStack provides panic recovery with stack trace
func RecoveryWithStack(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("correlation_id", c.GetString("correlation_id")),
					zap.Stack("stack"),
				)

				response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal server error occurred")
				c.Abort()
			}
		}()
		c.Next()
	}
}
