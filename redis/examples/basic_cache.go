//go:build ignore
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vhvcorp/go-shared/redis"
)

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

	// Create cache with configuration
	cache := redis.NewCache(client, redis.CacheConfig{
		DefaultTTL: 5 * time.Minute,
		KeyPrefix:  "example",
		Serializer: redis.NewJSONSerializer(),
	})

	ctx := context.Background()

	// Example 1: Simple string caching
	fmt.Println("=== Example 1: Simple String Caching ===")
	err = cache.Set(ctx, "greeting", "Hello, World!", 0)
	if err != nil {
		log.Printf("Set error: %v", err)
	}

	var greeting string
	err = cache.Get(ctx, "greeting", &greeting)
	if err != nil {
		log.Printf("Get error: %v", err)
	}
	fmt.Printf("Greeting: %s\n\n", greeting)

	// Example 2: Struct caching
	fmt.Println("=== Example 2: Struct Caching ===")
	type User struct {
		ID    int
		Name  string
		Email string
	}

	user := User{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
	}

	err = cache.Set(ctx, "user:1", user, 10*time.Minute)
	if err != nil {
		log.Printf("Set error: %v", err)
	}

	var retrievedUser User
	err = cache.Get(ctx, "user:1", &retrievedUser)
	if err != nil {
		log.Printf("Get error: %v", err)
	}
	fmt.Printf("User: %+v\n\n", retrievedUser)

	// Example 3: Multi-operations
	fmt.Println("=== Example 3: Batch Operations ===")
	
	// Batch set
	items := map[string]interface{}{
		"product:1": "Laptop",
		"product:2": "Mouse",
		"product:3": "Keyboard",
	}
	err = cache.MSet(ctx, items, 5*time.Minute)
	if err != nil {
		log.Printf("MSet error: %v", err)
	}

	// Batch get
	results, err := cache.MGet(ctx, "product:1", "product:2", "product:3")
	if err != nil {
		log.Printf("MGet error: %v", err)
	}
	for key, value := range results {
		fmt.Printf("%s = %v\n", key, value)
	}
	fmt.Println()

	// Example 4: Counter operations
	fmt.Println("=== Example 4: Counter Operations ===")
	
	// Increment page views
	views, err := cache.Increment(ctx, "page:views")
	if err != nil {
		log.Printf("Increment error: %v", err)
	}
	fmt.Printf("Page views: %d\n", views)

	// Increment by 10
	views, err = cache.IncrementBy(ctx, "page:views", 10)
	if err != nil {
		log.Printf("IncrementBy error: %v", err)
	}
	fmt.Printf("Page views after +10: %d\n", views)

	// Float increment
	score, err := cache.IncrementFloat(ctx, "user:score", 2.5)
	if err != nil {
		log.Printf("IncrementFloat error: %v", err)
	}
	fmt.Printf("User score: %.2f\n\n", score)

	// Example 5: TTL management
	fmt.Println("=== Example 5: TTL Management ===")
	
	err = cache.Set(ctx, "temp:data", "temporary", 10*time.Second)
	if err != nil {
		log.Printf("Set error: %v", err)
	}

	ttl, err := cache.GetTTL(ctx, "temp:data")
	if err != nil {
		log.Printf("GetTTL error: %v", err)
	}
	fmt.Printf("TTL: %v\n", ttl)

	// Extend TTL
	err = cache.Expire(ctx, "temp:data", 1*time.Minute)
	if err != nil {
		log.Printf("Expire error: %v", err)
	}
	fmt.Println("Extended TTL to 1 minute")

	// Make persistent
	err = cache.Persist(ctx, "temp:data")
	if err != nil {
		log.Printf("Persist error: %v", err)
	}
	fmt.Println("Made key persistent\n")

	// Example 6: Pattern-based operations
	fmt.Println("=== Example 6: Pattern-Based Operations ===")
	
	// Set multiple keys with pattern
	for i := 1; i <= 5; i++ {
		key := fmt.Sprintf("session:%d", i)
		cache.Set(ctx, key, fmt.Sprintf("session-data-%d", i), 0)
	}

	// List keys matching pattern
	keys, err := cache.Keys(ctx, "session:*")
	if err != nil {
		log.Printf("Keys error: %v", err)
	}
	fmt.Printf("Found %d session keys: %v\n", len(keys), keys)

	// Delete keys by pattern
	count, err := cache.DeleteByPattern(ctx, "session:*")
	if err != nil {
		log.Printf("DeleteByPattern error: %v", err)
	}
	fmt.Printf("Deleted %d session keys\n\n", count)

	// Example 7: Namespaced caches
	fmt.Println("=== Example 7: Namespaced Caches ===")
	
	userCache := cache.WithPrefix("users")
	sessionCache := cache.WithPrefix("sessions")

	userCache.Set(ctx, "1", "User data", 0)
	sessionCache.Set(ctx, "1", "Session data", 0)

	var userData, sessionData string
	userCache.Get(ctx, "1", &userData)
	sessionCache.Get(ctx, "1", &sessionData)

	fmt.Printf("User data: %s\n", userData)
	fmt.Printf("Session data: %s\n\n", sessionData)

	// Clean up
	fmt.Println("=== Cleanup ===")
	cache.FlushPrefix(ctx)
	fmt.Println("Flushed all cache keys")
}
