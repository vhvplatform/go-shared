//go:build ignore
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vhvcorp/go-shared/redis"
)

// Tenant represents a tenant in the system
type Tenant struct {
	ID   string
	Name string
	Plan string
}

// TenantCacheManager manages cache isolation for tenants
type TenantCacheManager struct {
	baseCache *redis.Cache
	caches    map[string]*redis.Cache
}

// NewTenantCacheManager creates a new tenant cache manager
func NewTenantCacheManager(cache *redis.Cache) *TenantCacheManager {
	return &TenantCacheManager{
		baseCache: cache,
		caches:    make(map[string]*redis.Cache),
	}
}

// GetTenantCache returns an isolated cache for a specific tenant
func (tcm *TenantCacheManager) GetTenantCache(tenantID string) *redis.Cache {
	if cache, exists := tcm.caches[tenantID]; exists {
		return cache
	}
	
	cache := tcm.baseCache.WithPrefix(fmt.Sprintf("tenant:%s", tenantID))
	tcm.caches[tenantID] = cache
	return cache
}

// FlushTenant removes all cache entries for a specific tenant
func (tcm *TenantCacheManager) FlushTenant(ctx context.Context, tenantID string) error {
	cache := tcm.GetTenantCache(tenantID)
	return cache.FlushPrefix(ctx)
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

	// Create base cache
	baseCache := redis.NewCache(client, redis.CacheConfig{
		DefaultTTL: 10 * time.Minute,
		KeyPrefix:  "multitenant-example",
		Serializer: redis.NewJSONSerializer(),
	})

	ctx := context.Background()

	// Clean up from previous runs
	baseCache.FlushPrefix(ctx)

	// Example 1: Basic Tenant Isolation
	fmt.Println("=== Example 1: Basic Tenant Isolation ===")
	
	manager := NewTenantCacheManager(baseCache)
	
	// Get caches for different tenants
	tenant1Cache := manager.GetTenantCache("tenant-1")
	tenant2Cache := manager.GetTenantCache("tenant-2")
	
	// Store data for tenant 1
	tenant1Cache.Set(ctx, "setting:theme", "dark", 0)
	tenant1Cache.Set(ctx, "setting:language", "en", 0)
	
	// Store data for tenant 2
	tenant2Cache.Set(ctx, "setting:theme", "light", 0)
	tenant2Cache.Set(ctx, "setting:language", "fr", 0)
	
	// Retrieve data
	var tenant1Theme, tenant2Theme string
	tenant1Cache.Get(ctx, "setting:theme", &tenant1Theme)
	tenant2Cache.Get(ctx, "setting:theme", &tenant2Theme)
	
	fmt.Printf("Tenant 1 theme: %s\n", tenant1Theme)
	fmt.Printf("Tenant 2 theme: %s\n\n", tenant2Theme)

	// Example 2: Tenant-Specific User Data
	fmt.Println("=== Example 2: Tenant-Specific User Data ===")
	
	type User struct {
		ID       string
		Name     string
		TenantID string
	}
	
	tenants := []Tenant{
		{ID: "tenant-a", Name: "Company A", Plan: "premium"},
		{ID: "tenant-b", Name: "Company B", Plan: "free"},
	}
	
	// Store users for each tenant
	for _, tenant := range tenants {
		cache := manager.GetTenantCache(tenant.ID)
		
		for j := 1; j <= 3; j++ {
			user := User{
				ID:       fmt.Sprintf("user-%d", j),
				Name:     fmt.Sprintf("User %d", j),
				TenantID: tenant.ID,
			}
			
			key := fmt.Sprintf("user:%s", user.ID)
			cache.Set(ctx, key, user, 0)
		}
		
		fmt.Printf("Stored %d users for %s\n", 3, tenant.Name)
	}
	fmt.Println()

	// Example 3: Per-Tenant Configuration
	fmt.Println("=== Example 3: Per-Tenant Configuration ===")
	
	type TenantConfig struct {
		MaxUsers      int
		StorageQuota  int64
		FeaturesFlags map[string]bool
	}
	
	for _, tenant := range tenants {
		cache := manager.GetTenantCache(tenant.ID)
		
		config := TenantConfig{
			MaxUsers:     100,
			StorageQuota: 1024 * 1024 * 1024, // 1GB
			FeaturesFlags: map[string]bool{
				"advanced_analytics": tenant.Plan == "premium",
				"custom_branding":    tenant.Plan == "premium",
				"api_access":         true,
			},
		}
		
		cache.Set(ctx, "config", config, 24*time.Hour)
		
		// Retrieve and display
		var retrievedConfig TenantConfig
		cache.Get(ctx, "config", &retrievedConfig)
		
		fmt.Printf("%s (%s plan):\n", tenant.Name, tenant.Plan)
		fmt.Printf("  Max Users: %d\n", retrievedConfig.MaxUsers)
		fmt.Printf("  Features: %+v\n\n", retrievedConfig.FeaturesFlags)
	}

	// Example 4: Tenant-Specific Session Management
	fmt.Println("=== Example 4: Tenant-Specific Sessions ===")
	
	type Session struct {
		UserID    string
		TenantID  string
		LoginTime time.Time
		IP        string
	}
	
	// Create sessions for different tenants
	sessions := []Session{
		{UserID: "user-1", TenantID: "tenant-a", LoginTime: time.Now(), IP: "192.168.1.1"},
		{UserID: "user-2", TenantID: "tenant-a", LoginTime: time.Now(), IP: "192.168.1.2"},
		{UserID: "user-1", TenantID: "tenant-b", LoginTime: time.Now(), IP: "192.168.1.3"},
	}
	
	for _, session := range sessions {
		cache := manager.GetTenantCache(session.TenantID)
		sessionKey := fmt.Sprintf("session:%s", session.UserID)
		
		cache.Set(ctx, sessionKey, session, 30*time.Minute)
		fmt.Printf("Created session for %s in %s from %s\n", 
			session.UserID, session.TenantID, session.IP)
	}
	fmt.Println()

	// Example 5: Tenant-Specific Rate Limiting
	fmt.Println("=== Example 5: Tenant-Specific Rate Limiting ===")
	
	type RateLimitConfig struct {
		RequestsPerMinute int64
		BurstSize         int64
	}
	
	limits := map[string]RateLimitConfig{
		"tenant-a": {RequestsPerMinute: 100, BurstSize: 20},
		"tenant-b": {RequestsPerMinute: 10, BurstSize: 5},
	}
	
	for tenantID, limit := range limits {
		cache := manager.GetTenantCache(tenantID)
		
		// Simulate requests
		successCount := 0
		for i := 0; i < 15; i++ {
			count, _ := cache.Increment(ctx, "ratelimit:requests")
			if count == 1 {
				cache.Expire(ctx, "ratelimit:requests", 1*time.Minute)
			}
			
			if count <= limit.RequestsPerMinute {
				successCount++
			}
		}
		
		fmt.Printf("%s: %d/%d requests allowed (limit: %d/min)\n", 
			tenantID, successCount, 15, limit.RequestsPerMinute)
	}
	fmt.Println()

	// Example 6: Tenant Data Analytics
	fmt.Println("=== Example 6: Tenant Analytics ===")
	
	for _, tenant := range tenants {
		cache := manager.GetTenantCache(tenant.ID)
		
		// Track various metrics
		cache.IncrementBy(ctx, "metrics:page_views", 150)
		cache.IncrementBy(ctx, "metrics:api_calls", 75)
		cache.IncrementBy(ctx, "metrics:active_users", 12)
		
		// Retrieve metrics
		var pageViews, apiCalls, activeUsers int64
		cache.Get(ctx, "metrics:page_views", &pageViews)
		cache.Get(ctx, "metrics:api_calls", &apiCalls)
		cache.Get(ctx, "metrics:active_users", &activeUsers)
		
		fmt.Printf("%s analytics:\n", tenant.Name)
		fmt.Printf("  Page Views: %d\n", pageViews)
		fmt.Printf("  API Calls: %d\n", apiCalls)
		fmt.Printf("  Active Users: %d\n\n", activeUsers)
	}

	// Example 7: Tenant Cache Patterns
	fmt.Println("=== Example 7: Tenant-Specific Cache Patterns ===")
	
	// List all keys for a tenant
	tenant1Cache = manager.GetTenantCache("tenant-a")
	keys, err := tenant1Cache.Keys(ctx, "*")
	if err != nil {
		log.Printf("Error listing keys: %v", err)
	}
	fmt.Printf("Tenant A has %d cache keys\n", len(keys))
	
	// Count specific patterns
	sessionCount, _ := tenant1Cache.DeleteByPattern(ctx, "session:*")
	fmt.Printf("Cleared %d session keys for Tenant A\n\n", sessionCount)

	// Example 8: Hierarchical Tenant Caching
	fmt.Println("=== Example 8: Hierarchical Caching ===")
	
	tenant := tenants[0]
	tenantCache := manager.GetTenantCache(tenant.ID)
	
	// Create hierarchical sub-caches
	userCache := tenantCache.WithPrefix("users")
	productCache := tenantCache.WithPrefix("products")
	orderCache := tenantCache.WithPrefix("orders")
	
	// Store data in different namespaces
	userCache.Set(ctx, "1", "User 1 data", 0)
	productCache.Set(ctx, "1", "Product 1 data", 0)
	orderCache.Set(ctx, "1", "Order 1 data", 0)
	
	var userData, productData, orderData string
	userCache.Get(ctx, "1", &userData)
	productCache.Get(ctx, "1", &productData)
	orderCache.Get(ctx, "1", &orderData)
	
	fmt.Printf("Hierarchical cache for %s:\n", tenant.Name)
	fmt.Printf("  User: %s\n", userData)
	fmt.Printf("  Product: %s\n", productData)
	fmt.Printf("  Order: %s\n\n", orderData)

	// Example 9: Flush Specific Tenant
	fmt.Println("=== Example 9: Tenant Cache Management ===")
	
	// Count keys before flush
	keys, _ = manager.GetTenantCache("tenant-a").Keys(ctx, "*")
	fmt.Printf("Tenant A has %d keys before flush\n", len(keys))
	
	// Flush tenant A
	err = manager.FlushTenant(ctx, "tenant-a")
	if err != nil {
		log.Printf("Error flushing tenant: %v", err)
	}
	
	// Count keys after flush
	keys, _ = manager.GetTenantCache("tenant-a").Keys(ctx, "*")
	fmt.Printf("Tenant A has %d keys after flush\n", len(keys))
	
	// Verify tenant B is unaffected
	keys, _ = manager.GetTenantCache("tenant-b").Keys(ctx, "*")
	fmt.Printf("Tenant B still has %d keys\n\n", len(keys))

	// Example 10: Real-world Tenant Service
	fmt.Println("=== Example 10: Real-World Usage ===")
	
	type TenantService struct {
		manager *TenantCacheManager
	}
	
	service := &TenantService{manager: manager}
	
	getUserSettings := func(ctx context.Context, tenantID, userID string) (map[string]interface{}, error) {
		cache := service.manager.GetTenantCache(tenantID)
		key := fmt.Sprintf("user:%s:settings", userID)
		
		// Try cache first
		var settings map[string]interface{}
		err := cache.Get(ctx, key, &settings)
		if err == nil {
			fmt.Println("  (from cache)")
			return settings, nil
		}
		
		// Cache miss - simulate database fetch
		fmt.Println("  (from database)")
		settings = map[string]interface{}{
			"theme":        "dark",
			"notifications": true,
			"language":     "en",
		}
		
		// Store in cache
		cache.Set(ctx, key, settings, 5*time.Minute)
		
		return settings, nil
	}
	
	// First call - cache miss
	fmt.Println("First call for Tenant A, User 1:")
	settings, _ := getUserSettings(ctx, "tenant-a", "user-1")
	fmt.Printf("Settings: %+v\n\n", settings)
	
	// Second call - cache hit
	fmt.Println("Second call for Tenant A, User 1:")
	settings, _ = getUserSettings(ctx, "tenant-a", "user-1")
	fmt.Printf("Settings: %+v\n\n", settings)

	// Clean up
	fmt.Println("=== Cleanup ===")
	baseCache.FlushPrefix(ctx)
	fmt.Println("Cleanup complete")
}
