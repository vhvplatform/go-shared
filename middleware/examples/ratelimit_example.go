//go:build examples

package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/vhvplatform/go-shared/middleware"
)

func main() {
	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	router := gin.Default()

	// Example 1: Basic Rate Limiting
	rateLimitedGroup := router.Group("/api")
	rateLimitedGroup.Use(middleware.PerIP(10, 20)) // 10 req/sec with burst of 20
	rateLimitedGroup.GET("/data", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Success"})
	})

	// Example 2: Rate Limiting by Tenant
	tenantGroup := router.Group("/tenant")
	tenantGroup.Use(middleware.PerTenant(100, 200))
	tenantGroup.GET("/info", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Tenant info"})
	})

	// Example 3: Custom Rate Limiting
	apiKeyGroup := router.Group("/premium")
	apiKeyGroup.Use(middleware.RateLimit(1000, 2000, func(c *gin.Context) string {
		// Rate limit by API key
		return c.GetHeader("X-API-Key")
	}))
	apiKeyGroup.GET("/service", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Premium service"})
	})

	log.Println("Server starting on :8080")
	router.Run(":8080")
}
