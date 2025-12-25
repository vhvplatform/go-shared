package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vhvcorp/go-shared/response"
	"golang.org/x/time/rate"
)

// limiterEntry holds a rate limiter and its last access time
type limiterEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// RateLimiter implements rate limiting with automatic cleanup
type RateLimiter struct {
	limiters     map[string]*limiterEntry
	mu           sync.RWMutex
	rate         rate.Limit
	burst        int
	cleanupOnce  sync.Once
	cleanupDone  chan struct{}
	cleanupTimer *time.Ticker
}

// NewRateLimiter creates a new rate limiter
// rps: requests per second
// burst: maximum burst size
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters:    make(map[string]*limiterEntry),
		rate:        rate.Limit(rps),
		burst:       burst,
		cleanupDone: make(chan struct{}),
	}
}

// GetLimiter returns a limiter for the given key
func (rl *RateLimiter) GetLimiter(key string) *rate.Limiter {
	rl.mu.RLock()
	_, exists := rl.limiters[key]
	rl.mu.RUnlock()

	if exists {
		// Update last access time
		rl.mu.Lock()
		// Check again after acquiring write lock
		if entry, exists := rl.limiters[key]; exists {
			entry.lastAccess = time.Now()
			rl.mu.Unlock()
			return entry.limiter
		}
		rl.mu.Unlock()
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if entry, exists := rl.limiters[key]; exists {
		entry.lastAccess = time.Now()
		return entry.limiter
	}

	limiter := rate.NewLimiter(rl.rate, rl.burst)
	rl.limiters[key] = &limiterEntry{
		limiter:    limiter,
		lastAccess: time.Now(),
	}

	return limiter
}

// CleanupLimiters removes inactive limiters periodically
func (rl *RateLimiter) CleanupLimiters(ctx context.Context) {
	rl.cleanupOnce.Do(func() {
		rl.cleanupTimer = time.NewTicker(5 * time.Minute)
		go func() {
			defer rl.cleanupTimer.Stop()
			defer close(rl.cleanupDone)

			for {
				select {
				case <-rl.cleanupTimer.C:
					rl.mu.Lock()
					now := time.Now()
					for key, entry := range rl.limiters {
						// Delete limiters inactive for more than 10 minutes
						if now.Sub(entry.lastAccess) > 10*time.Minute {
							delete(rl.limiters, key)
						}
					}
					rl.mu.Unlock()
				case <-ctx.Done():
					return
				}
			}
		}()
	})
}

// Stop stops the cleanup goroutine
func (rl *RateLimiter) Stop() {
	if rl.cleanupTimer != nil {
		rl.cleanupTimer.Stop()
		<-rl.cleanupDone
	}
}

// PerIP creates a rate limiting middleware that limits by IP address
func PerIP(rps float64, burst int) gin.HandlerFunc {
	rl := NewRateLimiter(rps, burst)
	// Start cleanup with background context - will run until process ends
	go rl.CleanupLimiters(context.Background())

	return func(c *gin.Context) {
		key := c.ClientIP()
		limiter := rl.GetLimiter(key)

		if !limiter.Allow() {
			response.Error(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}

// PerTenant creates a rate limiting middleware that limits by tenant ID
func PerTenant(rps float64, burst int) gin.HandlerFunc {
	rl := NewRateLimiter(rps, burst)
	// Start cleanup with background context - will run until process ends
	go rl.CleanupLimiters(context.Background())

	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = c.GetHeader("X-Tenant-ID")
		}

		if tenantID == "" {
			// If no tenant ID, fall back to IP
			tenantID = c.ClientIP()
		}

		limiter := rl.GetLimiter(tenantID)

		if !limiter.Allow() {
			response.Error(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}

// PerUser creates a rate limiting middleware that limits by user ID
func PerUser(rps float64, burst int) gin.HandlerFunc {
	rl := NewRateLimiter(rps, burst)
	// Start cleanup with background context - will run until process ends
	go rl.CleanupLimiters(context.Background())

	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			// If no user ID, fall back to IP
			userID = c.ClientIP()
		}

		limiter := rl.GetLimiter(userID)

		if !limiter.Allow() {
			response.Error(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimit creates a generic rate limiting middleware with a custom key extractor
func RateLimit(rps float64, burst int, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	rl := NewRateLimiter(rps, burst)
	// Start cleanup with background context - will run until process ends
	go rl.CleanupLimiters(context.Background())

	return func(c *gin.Context) {
		key := keyFunc(c)
		if key == "" {
			key = c.ClientIP()
		}

		limiter := rl.GetLimiter(key)

		if !limiter.Allow() {
			response.Error(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}
