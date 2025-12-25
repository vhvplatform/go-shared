package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vhvcorp/go-shared/logger"
	"go.uber.org/zap"
)

// Logger logs HTTP requests with detailed information
// It captures method, path, query, status code, latency, IP, and user agent
func Logger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		log.Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", c.Writer.Status()),
			zap.String("latency", latency.String()),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("correlation_id", c.GetString("correlation_id")),
			zap.String("user_id", c.GetString("user_id")),
			zap.String("tenant_id", c.GetString("tenant_id")),
		)
	}
}

// RequestLogger logs incoming requests before processing
func RequestLogger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Info("Incoming Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("ip", c.ClientIP()),
			zap.String("correlation_id", c.GetString("correlation_id")),
		)
		c.Next()
	}
}
