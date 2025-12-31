package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/vhvplatform/go-shared/response"
)

// ReplayProtectionConfig holds configuration for replay attack protection
type ReplayProtectionConfig struct {
	// RedisClient is the Redis client for storing nonces
	RedisClient *redis.Client

	// NonceHeader is the header name for the nonce (default: "X-Request-Nonce")
	NonceHeader string

	// TimestampHeader is the header name for the timestamp (default: "X-Request-Timestamp")
	TimestampHeader string

	// TimeWindow is the maximum age of a request in seconds (default: 300 = 5 minutes)
	TimeWindow time.Duration

	// KeyPrefix is the Redis key prefix for nonces (default: "replay:")
	KeyPrefix string
}

// ReplayProtection creates middleware to prevent replay attacks
// It requires clients to send a unique nonce and timestamp with each request
func ReplayProtection(config ReplayProtectionConfig) gin.HandlerFunc {
	// Set defaults
	if config.NonceHeader == "" {
		config.NonceHeader = "X-Request-Nonce"
	}
	if config.TimestampHeader == "" {
		config.TimestampHeader = "X-Request-Timestamp"
	}
	if config.TimeWindow == 0 {
		config.TimeWindow = 5 * time.Minute
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = "replay:"
	}
	if config.RedisClient == nil {
		panic("ReplayProtection: RedisClient is required")
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Extract nonce from header
		nonce := c.GetHeader(config.NonceHeader)
		if nonce == "" {
			response.Error(c, http.StatusBadRequest, "MISSING_NONCE", "Request nonce is required")
			c.Abort()
			return
		}

		// Extract timestamp from header
		timestampStr := c.GetHeader(config.TimestampHeader)
		if timestampStr == "" {
			response.Error(c, http.StatusBadRequest, "MISSING_TIMESTAMP", "Request timestamp is required")
			c.Abort()
			return
		}

		// Parse timestamp
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_TIMESTAMP", "Invalid timestamp format")
			c.Abort()
			return
		}

		requestTime := time.Unix(timestamp, 0)
		now := time.Now()

		// Check if request is within time window
		age := now.Sub(requestTime)
		if age < 0 {
			response.Error(c, http.StatusBadRequest, "FUTURE_TIMESTAMP", "Request timestamp is in the future")
			c.Abort()
			return
		}
		if age > config.TimeWindow {
			response.Error(c, http.StatusBadRequest, "EXPIRED_REQUEST", "Request has expired")
			c.Abort()
			return
		}

		// Create a unique key combining nonce, timestamp, and optionally user ID
		userID := c.GetString("user_id")
		nonceKey := buildNonceKey(config.KeyPrefix, nonce, timestampStr, userID)

		// Check if nonce has been used before using Redis SET NX (set if not exists)
		result, err := config.RedisClient.SetNX(ctx, nonceKey, "1", config.TimeWindow).Result()
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "REPLAY_CHECK_FAILED", "Failed to verify request uniqueness")
			c.Abort()
			return
		}

		// If SetNX returns false, the key already exists = replay attack
		if !result {
			response.Error(c, http.StatusBadRequest, "REPLAY_DETECTED", "Request replay detected")
			c.Abort()
			return
		}

		c.Next()
	}
}

// ReplayProtectionWithHash creates middleware that uses request hash for replay detection
// This is useful when you want to detect replays based on request body/headers
func ReplayProtectionWithHash(config ReplayProtectionConfig, includeBody bool) gin.HandlerFunc {
	// Set defaults
	if config.TimeWindow == 0 {
		config.TimeWindow = 5 * time.Minute
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = "replay:hash:"
	}
	if config.RedisClient == nil {
		panic("ReplayProtectionWithHash: RedisClient is required")
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Build request signature
		signature := buildRequestSignature(c, includeBody)

		// Create Redis key
		nonceKey := config.KeyPrefix + signature

		// Check if this exact request has been seen before
		result, err := config.RedisClient.SetNX(ctx, nonceKey, "1", config.TimeWindow).Result()
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "REPLAY_CHECK_FAILED", "Failed to verify request uniqueness")
			c.Abort()
			return
		}

		// If SetNX returns false, the key already exists = replay attack
		if !result {
			response.Error(c, http.StatusBadRequest, "REPLAY_DETECTED", "Duplicate request detected")
			c.Abort()
			return
		}

		c.Next()
	}
}

// buildNonceKey creates a unique key for the nonce
func buildNonceKey(prefix, nonce, timestamp, userID string) string {
	if userID != "" {
		return fmt.Sprintf("%s%s:%s:%s", prefix, userID, timestamp, nonce)
	}
	return fmt.Sprintf("%s%s:%s", prefix, timestamp, nonce)
}

// buildRequestSignature creates a hash of the request for replay detection
func buildRequestSignature(c *gin.Context, includeBody bool) string {
	hasher := sha256.New()

	// Include method and path
	hasher.Write([]byte(c.Request.Method))
	hasher.Write([]byte(c.Request.URL.Path))

	// Include user ID if available
	if userID := c.GetString("user_id"); userID != "" {
		hasher.Write([]byte(userID))
	}

	// Include body if requested
	if includeBody {
		if body, exists := c.Get(gin.BodyBytesKey); exists {
			if bodyBytes, ok := body.([]byte); ok {
				hasher.Write(bodyBytes)
			}
		}
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

// CleanupExpiredNonces removes expired nonces from Redis (maintenance function)
// This is typically not needed as Redis TTL handles cleanup automatically
func CleanupExpiredNonces(ctx context.Context, redisClient *redis.Client, keyPattern string) (int, error) {
	if keyPattern == "" {
		keyPattern = "replay:*"
	}

	var cursor uint64
	var deletedCount int

	for {
		var keys []string
		var err error
		keys, cursor, err = redisClient.Scan(ctx, cursor, keyPattern, 100).Result()
		if err != nil {
			return deletedCount, fmt.Errorf("failed to scan keys: %w", err)
		}

		if len(keys) > 0 {
			// Delete keys in batch
			deleted, err := redisClient.Del(ctx, keys...).Result()
			if err != nil {
				return deletedCount, fmt.Errorf("failed to delete keys: %w", err)
			}
			deletedCount += int(deleted)
		}

		if cursor == 0 {
			break
		}
	}

	return deletedCount, nil
}
