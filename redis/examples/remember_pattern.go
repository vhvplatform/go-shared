//go:build ignore
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vhvcorp/go-shared/redis"
)

// Simulated database functions
func fetchUserFromDB(userID string) (map[string]string, error) {
	fmt.Printf("Fetching user %s from database...\n", userID)
	time.Sleep(100 * time.Millisecond) // Simulate slow DB query
	return map[string]string{
		"id":    userID,
		"name":  "John Doe",
		"email": "john@example.com",
	}, nil
}

func fetchConfigFromDB() (map[string]interface{}, error) {
	fmt.Println("Loading configuration from database...")
	time.Sleep(200 * time.Millisecond) // Simulate slow DB query
	return map[string]interface{}{
		"max_connections": 100,
		"timeout":         30,
		"debug":           false,
	}, nil
}

func calculateExpensiveValue(input int) (int, error) {
	fmt.Printf("Computing expensive calculation for input %d...\n", input)
	time.Sleep(150 * time.Millisecond) // Simulate expensive computation
	return input * input * input, nil
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
		KeyPrefix:  "remember-example",
		Serializer: redis.NewJSONSerializer(),
	})

	ctx := context.Background()

	// Clean up from previous runs
	cache.FlushPrefix(ctx)

	// Example 1: Basic Remember Pattern
	fmt.Println("=== Example 1: Basic Remember Pattern ===")
	
	userID := "123"
	key := fmt.Sprintf("user:%s", userID)

	// First call: Cache miss, executes function
	fmt.Println("First call (cache miss):")
	result, err := cache.Remember(ctx, key, 5*time.Minute, func() (interface{}, error) {
		return fetchUserFromDB(userID)
	})
	if err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("Result: %v\n\n", result)

	// Second call: Cache hit, returns cached value
	fmt.Println("Second call (cache hit):")
	result, err = cache.Remember(ctx, key, 5*time.Minute, func() (interface{}, error) {
		return fetchUserFromDB(userID) // This won't be called
	})
	if err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("Result: %v\n\n", result)

	// Example 2: RememberForever for Configuration
	fmt.Println("=== Example 2: RememberForever for Config ===")
	
	// First call: Loads from database
	fmt.Println("First call (cache miss):")
	config, err := cache.RememberForever(ctx, "app:config", func() (interface{}, error) {
		return fetchConfigFromDB()
	})
	if err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("Config: %v\n\n", config)

	// Second call: Returns cached value
	fmt.Println("Second call (cache hit):")
	config, err = cache.RememberForever(ctx, "app:config", func() (interface{}, error) {
		return fetchConfigFromDB() // This won't be called
	})
	if err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("Config: %v\n\n", config)

	// Example 3: Remember with Computation
	fmt.Println("=== Example 3: Remember with Expensive Computation ===")
	
	input := 42
	computeKey := fmt.Sprintf("compute:%d", input)

	// First call: Performs computation
	fmt.Println("First call (cache miss):")
	start := time.Now()
	result, err = cache.Remember(ctx, computeKey, 10*time.Minute, func() (interface{}, error) {
		return calculateExpensiveValue(input)
	})
	if err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("Result: %v (took %v)\n\n", result, time.Since(start))

	// Second call: Returns cached result instantly
	fmt.Println("Second call (cache hit):")
	start = time.Now()
	result, err = cache.Remember(ctx, computeKey, 10*time.Minute, func() (interface{}, error) {
		return calculateExpensiveValue(input) // This won't be called
	})
	if err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("Result: %v (took %v)\n\n", result, time.Since(start))

	// Example 4: Remember with Error Handling
	fmt.Println("=== Example 4: Error Handling ===")
	
	result, err = cache.Remember(ctx, "error:key", 5*time.Minute, func() (interface{}, error) {
		return nil, fmt.Errorf("simulated error")
	})
	if err != nil {
		fmt.Printf("Expected error: %v\n\n", err)
	}

	// Example 5: Real-world GetUser Function
	fmt.Println("=== Example 5: Real-World GetUser Function ===")
	
	GetUser := func(ctx context.Context, cache *redis.Cache, userID string) (map[string]string, error) {
		key := fmt.Sprintf("user:%s", userID)
		
		result, err := cache.Remember(ctx, key, 5*time.Minute, func() (interface{}, error) {
			return fetchUserFromDB(userID)
		})
		
		if err != nil {
			return nil, err
		}
		
		// Safe type assertion
		user, ok := result.(map[string]string)
		if !ok {
			return nil, fmt.Errorf("unexpected type from cache: %T", result)
		}
		
		return user, nil
	}

	// Use the function multiple times
	for i := 0; i < 3; i++ {
		fmt.Printf("Call %d:\n", i+1)
		user, err := GetUser(ctx, cache, "456")
		if err != nil {
			log.Printf("Error: %v", err)
		}
		fmt.Printf("User: %v\n\n", user)
		time.Sleep(100 * time.Millisecond)
	}

	// Example 6: Cache Invalidation
	fmt.Println("=== Example 6: Cache Invalidation ===")
	
	invalidateKey := "invalidate:test"
	
	// Set cached value
	cache.Remember(ctx, invalidateKey, 5*time.Minute, func() (interface{}, error) {
		return "original value", nil
	})
	
	// Retrieve cached value
	result, _ = cache.Remember(ctx, invalidateKey, 5*time.Minute, func() (interface{}, error) {
		return "new value", nil
	})
	fmt.Printf("Before invalidation: %v\n", result)
	
	// Invalidate cache
	cache.Delete(ctx, invalidateKey)
	
	// Value is recomputed
	result, _ = cache.Remember(ctx, invalidateKey, 5*time.Minute, func() (interface{}, error) {
		return "new value", nil
	})
	fmt.Printf("After invalidation: %v\n\n", result)

	// Clean up
	fmt.Println("=== Cleanup ===")
	cache.FlushPrefix(ctx)
	fmt.Println("Flushed all cache keys")
}
