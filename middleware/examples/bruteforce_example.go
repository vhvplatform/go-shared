//go:build examples

package main

import (
	"context"
	"log"
	"net/http"
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

	// Example 1: Basic Brute Force Protection (by IP)
	basicConfig := middleware.BruteForceProtectionConfig{
		RedisClient:     redisClient,
		MaxAttempts:     5,
		LockoutDuration: 15 * time.Minute,
		AttemptWindow:   1 * time.Hour,
		KeyPrefix:       "bf:",
	}

	router.POST("/login-basic", middleware.BruteForceProtection(basicConfig), handleLogin)

	// Example 2: Per-User Brute Force Protection
	userConfig := middleware.BruteForceProtectionConfig{
		RedisClient:     redisClient,
		MaxAttempts:     5,
		LockoutDuration: 15 * time.Minute,
		AttemptWindow:   1 * time.Hour,
		KeyPrefix:       "bf:user:",
	}

	router.POST("/login-user", middleware.BruteForcePerUser(userConfig, "username"), handleLogin)

	// Example 3: Per-Email with Exponential Backoff
	emailConfig := middleware.BruteForceProtectionConfig{
		RedisClient:           redisClient,
		MaxAttempts:           3,
		LockoutDuration:       5 * time.Minute,
		AttemptWindow:         30 * time.Minute,
		UseExponentialBackoff: true,
		BackoffMultiplier:     2,
		KeyPrefix:             "bf:email:",
	}

	router.POST("/login-email", middleware.BruteForcePerEmail(emailConfig), handleLogin)

	// Admin endpoint to reset brute force protection
	router.POST("/admin/reset-bruteforce", func(c *gin.Context) {
		identifier := c.Query("identifier")
		if identifier == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "identifier required"})
			return
		}

		err := middleware.ResetBruteForceProtection(
			context.Background(),
			redisClient,
			"bf:",
			identifier,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Reset successful"})
	})

	// Admin endpoint to check brute force status
	router.GET("/admin/bruteforce-status", func(c *gin.Context) {
		identifier := c.Query("identifier")
		if identifier == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "identifier required"})
			return
		}

		locked, attempts, ttl, err := middleware.GetBruteForceStatus(
			context.Background(),
			redisClient,
			"bf:",
			identifier,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"locked":   locked,
			"attempts": attempts,
			"ttl":      ttl.String(),
		})
	})

	log.Println("Server starting on :8080")
	log.Println("Available endpoints:")
	log.Println("  POST /login-basic - Basic IP-based protection")
	log.Println("  POST /login-user - Username-based protection")
	log.Println("  POST /login-email - Email-based with exponential backoff")
	log.Println("  POST /admin/reset-bruteforce?identifier=xxx")
	log.Println("  GET  /admin/bruteforce-status?identifier=xxx")
	router.Run(":8080")
}

func handleLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	email := c.PostForm("email")

	// Get remaining attempts
	remaining, _ := middleware.GetRemainingAttempts(c)

	// Simulate credential validation
	// In production, validate against your database
	validCredentials := (username == "admin" && password == "secret") ||
		(email == "admin@example.com" && password == "secret")

	if !validCredentials {
		// Record failed attempt
		if err := middleware.RecordFailedAttempt(c); err != nil {
			log.Printf("Failed to record attempt: %v", err)
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"error":               "Invalid credentials",
			"remaining_attempts":  remaining - 1,
		})
		return
	}

	// Record successful attempt (resets counter)
	if err := middleware.RecordSuccessfulAttempt(c); err != nil {
		log.Printf("Failed to record success: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   "dummy-jwt-token",
	})
}
