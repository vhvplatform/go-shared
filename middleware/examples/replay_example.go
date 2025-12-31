// +build examples

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

	// Example 1: Basic Replay Protection
	// Clients must send X-Request-Nonce and X-Request-Timestamp headers
	config := middleware.ReplayProtectionConfig{
		RedisClient:     redisClient,
		NonceHeader:     "X-Request-Nonce",
		TimestampHeader: "X-Request-Timestamp",
		TimeWindow:      5 * time.Minute,
		KeyPrefix:       "replay:",
	}

	protectedAPI := router.Group("/api/v1")
	protectedAPI.Use(middleware.ReplayProtection(config))
	protectedAPI.POST("/transaction", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Transaction processed"})
	})

	// Example 2: Hash-Based Duplicate Detection
	// Prevents exact duplicate requests (useful for idempotency)
	hashConfig := middleware.ReplayProtectionConfig{
		RedisClient: redisClient,
		TimeWindow:  10 * time.Minute,
		KeyPrefix:   "replay:hash:",
	}

	idempotentAPI := router.Group("/api/idempotent")
	idempotentAPI.Use(middleware.ReplayProtectionWithHash(hashConfig, true))
	idempotentAPI.POST("/order", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Order created"})
	})

	log.Println("Server starting on :8080")
	log.Println("For replay protection, clients must send:")
	log.Println("  X-Request-Nonce: unique-value (UUID recommended)")
	log.Println("  X-Request-Timestamp: unix-timestamp")
	router.Run(":8080")
}
