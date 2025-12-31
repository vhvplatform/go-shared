package middleware

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/vhvplatform/go-shared/response"
)

// BruteForceProtectionConfig holds configuration for brute force protection
type BruteForceProtectionConfig struct {
	// RedisClient is the Redis client for tracking attempts
	RedisClient *redis.Client

	// MaxAttempts is the maximum number of attempts before locking (default: 5)
	MaxAttempts int

	// LockoutDuration is how long to lock the account/IP after max attempts (default: 15 minutes)
	LockoutDuration time.Duration

	// AttemptWindow is the time window for counting attempts (default: 1 hour)
	AttemptWindow time.Duration

	// KeyPrefix is the Redis key prefix (default: "bruteforce:")
	KeyPrefix string

	// IdentifierFunc extracts the identifier to track (e.g., IP, username, email)
	// Default: uses client IP
	IdentifierFunc func(*gin.Context) string

	// UseExponentialBackoff enables exponential backoff for lockout duration
	UseExponentialBackoff bool

	// BackoffMultiplier is the multiplier for exponential backoff (default: 2)
	BackoffMultiplier float64
}

// BruteForceProtection creates middleware to prevent brute force attacks
func BruteForceProtection(config BruteForceProtectionConfig) gin.HandlerFunc {
	// Set defaults
	if config.MaxAttempts == 0 {
		config.MaxAttempts = 5
	}
	if config.LockoutDuration == 0 {
		config.LockoutDuration = 15 * time.Minute
	}
	if config.AttemptWindow == 0 {
		config.AttemptWindow = 1 * time.Hour
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = "bruteforce:"
	}
	if config.IdentifierFunc == nil {
		config.IdentifierFunc = func(c *gin.Context) string {
			return c.ClientIP()
		}
	}
	if config.BackoffMultiplier == 0 {
		config.BackoffMultiplier = 2
	}
	if config.RedisClient == nil {
		// Panic is intentional here - this is a configuration error that should be
		// caught at application startup, not during request handling
		panic("BruteForceProtection: RedisClient is required")
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Get identifier (IP, username, email, etc.)
		identifier := config.IdentifierFunc(c)
		if identifier == "" {
			identifier = c.ClientIP()
		}

		// Check if currently locked out
		lockKey := config.KeyPrefix + "lock:" + identifier
		locked, err := config.RedisClient.Get(ctx, lockKey).Result()
		if err != nil && err != redis.Nil {
			response.Error(c, http.StatusInternalServerError, "BRUTEFORCE_CHECK_FAILED", "Failed to check brute force protection")
			c.Abort()
			return
		}

		if locked != "" {
			// Get remaining lockout time
			ttl, _ := config.RedisClient.TTL(ctx, lockKey).Result()
			response.ErrorWithDetails(c, http.StatusTooManyRequests, "ACCOUNT_LOCKED",
				"Too many failed attempts. Please try again later.",
				map[string]interface{}{
					"retry_after_seconds": int(ttl.Seconds()),
				})
			c.Abort()
			return
		}

		// Get current attempt count
		attemptKey := config.KeyPrefix + "attempts:" + identifier
		attemptCount, err := config.RedisClient.Get(ctx, attemptKey).Result()
		if err != nil && err != redis.Nil {
			response.Error(c, http.StatusInternalServerError, "BRUTEFORCE_CHECK_FAILED", "Failed to check brute force protection")
			c.Abort()
			return
		}

		attempts := 0
		if attemptCount != "" {
			attempts, _ = strconv.Atoi(attemptCount)
		}

		// Check if max attempts exceeded
		if attempts >= config.MaxAttempts {
			// Calculate lockout duration
			lockoutDuration := config.LockoutDuration
			if config.UseExponentialBackoff {
				lockoutDuration = calculateExponentialBackoff(
					config.LockoutDuration,
					attempts,
					config.MaxAttempts,
					config.BackoffMultiplier,
				)
			}

			// Lock the account/IP
			err = config.RedisClient.Set(ctx, lockKey, "1", lockoutDuration).Err()
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "BRUTEFORCE_LOCK_FAILED", "Failed to apply brute force protection")
				c.Abort()
				return
			}

			// Reset attempt counter
			config.RedisClient.Del(ctx, attemptKey)

			response.ErrorWithDetails(c, http.StatusTooManyRequests, "ACCOUNT_LOCKED",
				"Too many failed attempts. Account temporarily locked.",
				map[string]interface{}{
					"retry_after_seconds": int(lockoutDuration.Seconds()),
				})
			c.Abort()
			return
		}

		// Store attempt count in context for post-processing
		c.Set("bruteforce_attempts", attempts)
		c.Set("bruteforce_identifier", identifier)
		c.Set("bruteforce_config", &config)

		c.Next()
	}
}

// RecordFailedAttempt should be called after failed authentication
// It increments the attempt counter
func RecordFailedAttempt(c *gin.Context) error {
	config, exists := c.Get("bruteforce_config")
	if !exists {
		return fmt.Errorf("brute force config not found in context")
	}

	bfConfig, ok := config.(*BruteForceProtectionConfig)
	if !ok {
		return fmt.Errorf("invalid brute force config in context")
	}

	identifier := c.GetString("bruteforce_identifier")
	if identifier == "" {
		return fmt.Errorf("identifier not found in context")
	}

	ctx := c.Request.Context()
	attemptKey := bfConfig.KeyPrefix + "attempts:" + identifier

	// Increment attempt counter
	_, err := bfConfig.RedisClient.Incr(ctx, attemptKey).Result()
	if err != nil {
		return fmt.Errorf("failed to increment attempt counter: %w", err)
	}

	// Set expiration on first attempt
	bfConfig.RedisClient.Expire(ctx, attemptKey, bfConfig.AttemptWindow)

	return nil
}

// RecordSuccessfulAttempt should be called after successful authentication
// It resets the attempt counter
func RecordSuccessfulAttempt(c *gin.Context) error {
	config, exists := c.Get("bruteforce_config")
	if !exists {
		return nil // Not an error if config not found
	}

	bfConfig, ok := config.(*BruteForceProtectionConfig)
	if !ok {
		return fmt.Errorf("invalid brute force config in context")
	}

	identifier := c.GetString("bruteforce_identifier")
	if identifier == "" {
		return nil
	}

	ctx := c.Request.Context()
	attemptKey := bfConfig.KeyPrefix + "attempts:" + identifier

	// Reset attempt counter
	return bfConfig.RedisClient.Del(ctx, attemptKey).Err()
}

// GetRemainingAttempts returns the number of remaining attempts before lockout
func GetRemainingAttempts(c *gin.Context) (int, error) {
	config, exists := c.Get("bruteforce_config")
	if !exists {
		return 0, fmt.Errorf("brute force config not found in context")
	}

	bfConfig, ok := config.(*BruteForceProtectionConfig)
	if !ok {
		return 0, fmt.Errorf("invalid brute force config in context")
	}

	attempts, exists := c.Get("bruteforce_attempts")
	if !exists {
		return bfConfig.MaxAttempts, nil
	}

	currentAttempts, ok := attempts.(int)
	if !ok {
		return bfConfig.MaxAttempts, nil
	}

	remaining := bfConfig.MaxAttempts - currentAttempts
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// BruteForcePerUser creates brute force protection that tracks by user ID/username
func BruteForcePerUser(config BruteForceProtectionConfig, usernameField string) gin.HandlerFunc {
	if usernameField == "" {
		usernameField = "username"
	}

	config.IdentifierFunc = func(c *gin.Context) string {
		// Try to get username from form values first (doesn't consume body)
		if username := c.PostForm(usernameField); username != "" {
			return "user:" + username
		}

		// Try query parameter
		if username := c.Query(usernameField); username != "" {
			return "user:" + username
		}

		// Fallback to IP if no username found
		return "ip:" + c.ClientIP()
	}

	return BruteForceProtection(config)
}

// BruteForcePerEmail creates brute force protection that tracks by email
func BruteForcePerEmail(config BruteForceProtectionConfig) gin.HandlerFunc {
	config.IdentifierFunc = func(c *gin.Context) string {
		// Try to get email from form values first (doesn't consume body)
		if email := c.PostForm("email"); email != "" {
			return "email:" + email
		}

		// Try query parameter
		if email := c.Query("email"); email != "" {
			return "email:" + email
		}

		// Fallback to IP if no email found
		return "ip:" + c.ClientIP()
	}

	return BruteForceProtection(config)
}

// calculateExponentialBackoff calculates lockout duration with exponential backoff
func calculateExponentialBackoff(baseDuration time.Duration, attempts, maxAttempts int, multiplier float64) time.Duration {
	// Calculate how many times the max was exceeded
	overageCount := attempts - maxAttempts + 1

	// Apply exponential backoff: baseDuration * (multiplier ^ overageCount)
	backoffSeconds := float64(baseDuration.Seconds()) * math.Pow(multiplier, float64(overageCount))

	// Cap at a maximum (e.g., 24 hours)
	maxBackoff := 24 * time.Hour
	if backoffSeconds > maxBackoff.Seconds() {
		return maxBackoff
	}

	return time.Duration(backoffSeconds) * time.Second
}

// ResetBruteForceProtection manually resets brute force protection for an identifier
// Useful for admin operations
func ResetBruteForceProtection(ctx context.Context, redisClient *redis.Client, keyPrefix, identifier string) error {
	if keyPrefix == "" {
		keyPrefix = "bruteforce:"
	}

	// Remove both lock and attempts
	lockKey := keyPrefix + "lock:" + identifier
	attemptKey := keyPrefix + "attempts:" + identifier

	pipe := redisClient.Pipeline()
	pipe.Del(ctx, lockKey)
	pipe.Del(ctx, attemptKey)

	_, err := pipe.Exec(ctx)
	return err
}

// GetBruteForceStatus returns the current status for an identifier
func GetBruteForceStatus(ctx context.Context, redisClient *redis.Client, keyPrefix, identifier string) (locked bool, attempts int, ttl time.Duration, err error) {
	if keyPrefix == "" {
		keyPrefix = "bruteforce:"
	}

	lockKey := keyPrefix + "lock:" + identifier
	attemptKey := keyPrefix + "attempts:" + identifier

	// Check if locked
	lockedVal, err := redisClient.Get(ctx, lockKey).Result()
	if err != nil && err != redis.Nil {
		return false, 0, 0, err
	}
	locked = lockedVal != ""

	if locked {
		// Get TTL of lock
		ttl, err = redisClient.TTL(ctx, lockKey).Result()
		if err != nil {
			return locked, 0, 0, err
		}
	}

	// Get attempt count
	attemptCount, err := redisClient.Get(ctx, attemptKey).Result()
	if err != nil && err != redis.Nil {
		return locked, 0, ttl, err
	}

	if attemptCount != "" {
		attempts, _ = strconv.Atoi(attemptCount)
	}

	return locked, attempts, ttl, nil
}
