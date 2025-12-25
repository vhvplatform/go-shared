//go:build ignore
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vhvcorp/go-shared/redis"
)

// RateLimiter implements token bucket rate limiting using Redis
type RateLimiter struct {
	cache *redis.Cache
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cache *redis.Cache) *RateLimiter {
	return &RateLimiter{cache: cache}
}

// Allow checks if a request is allowed for the given key
// Returns true if allowed, false if rate limit exceeded
func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	count, err := rl.cache.Increment(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to increment counter: %w", err)
	}
	
	// Set expiration on first request
	if count == 1 {
		if err := rl.cache.Expire(ctx, key, window); err != nil {
			return false, fmt.Errorf("failed to set expiration: %w", err)
		}
	}
	
	return count <= limit, nil
}

// Remaining returns the number of remaining requests allowed
func (rl *RateLimiter) Remaining(ctx context.Context, key string, limit int64) (int64, error) {
	count, err := rl.cache.Exists(ctx, key)
	if err != nil {
		return 0, err
	}
	
	if count == 0 {
		return limit, nil
	}
	
	// Get current count
	var currentCount int64
	err = rl.cache.Get(ctx, key, &currentCount)
	if err != nil {
		return limit, nil // Key doesn't exist yet
	}
	
	remaining := limit - currentCount
	if remaining < 0 {
		return 0, nil
	}
	return remaining, nil
}

// Reset resets the rate limit for a key
func (rl *RateLimiter) Reset(ctx context.Context, key string) error {
	return rl.cache.Delete(ctx, key)
}

func main() {
	// Connect to Redis
	client, err := redis.NewClient(redis.Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer client.Close()

	// Create cache
	cache := redis.NewCache(client, redis.CacheConfig{
		DefaultTTL: 5 * time.Minute,
		KeyPrefix:  "ratelimit-example",
	})

	ctx := context.Background()

	// Clean up from previous runs
	cache.FlushPrefix(ctx)

	// Example 1: Basic Rate Limiting
	fmt.Println("=== Example 1: Basic Rate Limiting ===")
	
	limiter := NewRateLimiter(cache)
	userID := "user:123"
	
	// Allow 5 requests per minute
	limit := int64(5)
	window := 1 * time.Minute
	
	for i := 1; i <= 7; i++ {
		allowed, err := limiter.Allow(ctx, userID, limit, window)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		
		if allowed {
			fmt.Printf("Request %d: Allowed\n", i)
		} else {
			fmt.Printf("Request %d: Rate limit exceeded\n", i)
		}
	}
	fmt.Println()

	// Example 2: API Rate Limiting
	fmt.Println("=== Example 2: API Rate Limiting ===")
	
	apiKey := "api:key:abc123"
	apiLimit := int64(10)
	apiWindow := 1 * time.Minute
	
	checkAPIRateLimit := func(ctx context.Context, apiKey string) error {
		allowed, err := limiter.Allow(ctx, apiKey, apiLimit, apiWindow)
		if err != nil {
			return fmt.Errorf("rate limit check failed: %w", err)
		}
		
		if !allowed {
			remaining, _ := limiter.Remaining(ctx, apiKey, apiLimit)
			return fmt.Errorf("rate limit exceeded, %d requests remaining", remaining)
		}
		
		return nil
	}
	
	// Simulate API calls
	for i := 1; i <= 12; i++ {
		err := checkAPIRateLimit(ctx, apiKey)
		if err != nil {
			fmt.Printf("API call %d: %v\n", i, err)
		} else {
			fmt.Printf("API call %d: Success\n", i)
		}
	}
	fmt.Println()

	// Example 3: Per-IP Rate Limiting
	fmt.Println("=== Example 3: Per-IP Rate Limiting ===")
	
	ipLimit := int64(3)
	ipWindow := 10 * time.Second
	
	checkIPRateLimit := func(ctx context.Context, ip string) (bool, error) {
		key := fmt.Sprintf("ip:%s", ip)
		return limiter.Allow(ctx, key, ipLimit, ipWindow)
	}
	
	// Simulate requests from different IPs
	ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.1"}
	
	for _, ip := range ips {
		allowed, err := checkIPRateLimit(ctx, ip)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		
		if allowed {
			fmt.Printf("Request from %s: Allowed\n", ip)
		} else {
			fmt.Printf("Request from %s: Blocked\n", ip)
		}
	}
	fmt.Println()

	// Example 4: Tiered Rate Limiting
	fmt.Println("=== Example 4: Tiered Rate Limiting ===")
	
	type Tier struct {
		Name   string
		Limit  int64
		Window time.Duration
	}
	
	tiers := map[string]Tier{
		"free":    {Name: "Free", Limit: 10, Window: 1 * time.Minute},
		"premium": {Name: "Premium", Limit: 100, Window: 1 * time.Minute},
		"enterprise": {Name: "Enterprise", Limit: 1000, Window: 1 * time.Minute},
	}
	
	checkTieredRateLimit := func(ctx context.Context, userID, tierName string) (bool, error) {
		tier, exists := tiers[tierName]
		if !exists {
			tier = tiers["free"]
		}
		
		key := fmt.Sprintf("user:%s:tier:%s", userID, tierName)
		allowed, err := limiter.Allow(ctx, key, tier.Limit, tier.Window)
		if err != nil {
			return false, err
		}
		
		if !allowed {
			fmt.Printf("User %s (%s tier) exceeded limit of %d requests\n", 
				userID, tier.Name, tier.Limit)
		}
		
		return allowed, nil
	}
	
	// Test different tiers
	users := []struct {
		ID   string
		Tier string
	}{
		{"user1", "free"},
		{"user2", "premium"},
		{"user3", "enterprise"},
	}
	
	for _, user := range users {
		tier := tiers[user.Tier]
		fmt.Printf("Testing %s tier (limit: %d):\n", tier.Name, tier.Limit)
		
		successCount := 0
		for i := 0; i < 15; i++ {
			allowed, _ := checkTieredRateLimit(ctx, user.ID, user.Tier)
			if allowed {
				successCount++
			}
		}
		fmt.Printf("  Successful requests: %d out of 15\n\n", successCount)
	}

	// Example 5: Rate Limiting with Reset
	fmt.Println("=== Example 5: Rate Limiting with Reset ===")
	
	resetKey := "user:reset:test"
	
	// Make some requests
	for i := 1; i <= 3; i++ {
		limiter.Allow(ctx, resetKey, 5, 1*time.Minute)
		fmt.Printf("Request %d made\n", i)
	}
	
	// Reset the limit
	fmt.Println("Resetting rate limit...")
	limiter.Reset(ctx, resetKey)
	
	// Make more requests
	for i := 1; i <= 3; i++ {
		allowed, _ := limiter.Allow(ctx, resetKey, 5, 1*time.Minute)
		if allowed {
			fmt.Printf("Request %d after reset: Allowed\n", i)
		}
	}
	fmt.Println()

	// Example 6: Sliding Window Rate Limiting (Simple)
	fmt.Println("=== Example 6: Usage Statistics ===")
	
	// Track API usage
	endpoints := []string{"/api/users", "/api/posts", "/api/comments"}
	
	for _, endpoint := range endpoints {
		key := fmt.Sprintf("endpoint:%s", endpoint)
		count := int64((len(endpoint) % 5) + 3) // Simulate different counts
		
		for i := int64(0); i < count; i++ {
			cache.Increment(ctx, key)
		}
	}
	
	// Display statistics
	fmt.Println("Endpoint usage statistics:")
	for _, endpoint := range endpoints {
		key := fmt.Sprintf("endpoint:%s", endpoint)
		var count int64
		cache.Get(ctx, key, &count)
		fmt.Printf("  %s: %d requests\n", endpoint, count)
	}
	fmt.Println()

	// Example 7: Burst Handling
	fmt.Println("=== Example 7: Burst Handling ===")
	
	burstKey := "burst:test"
	normalLimit := int64(5)
	burstLimit := int64(10)
	
	// Allow burst initially
	fmt.Println("Allowing burst traffic:")
	for i := 1; i <= 12; i++ {
		var allowed bool
		if i <= 10 {
			allowed, _ = limiter.Allow(ctx, burstKey, burstLimit, 1*time.Minute)
		} else {
			allowed, _ = limiter.Allow(ctx, burstKey, normalLimit, 1*time.Minute)
		}
		
		if allowed {
			fmt.Printf("Burst request %d: Allowed\n", i)
		} else {
			fmt.Printf("Burst request %d: Blocked (burst quota exceeded)\n", i)
		}
	}

	// Clean up
	fmt.Println("\n=== Cleanup ===")
	cache.FlushPrefix(ctx)
	fmt.Println("Cleanup complete")
}
