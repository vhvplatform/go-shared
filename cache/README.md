# Cache Package

The cache package provides a unified interface for caching with support for multiple backends and two-tier caching strategies.

## Features

- **Unified Interface**: Single `Cache` interface for all cache backends
- **Two-Tier Caching**: Combine local (L1) and distributed (L2) caches for optimal performance
- **Multiple Backends**: 
  - Ristretto (in-memory, high-performance)
  - Redis (distributed, persistent)
- **Flexible Configuration**: Choose L1-only, L2-only, or tiered caching

## Quick Start

### Ristretto Cache (Local)

```go
import "github.com/vhvplatform/go-shared/cache"

// Create local cache
localCache, err := cache.NewRistrettoCache(cache.RistrettoConfig{
    MaxCost:     10000,   // ~10MB
    NumCounters: 100000,  // ~10x expected keys
})

// Use cache
localCache.Set(ctx, "key", value, 5*time.Minute)
var result MyType
localCache.Get(ctx, "key", &result)
```

### Redis Cache (Distributed)

```go
import (
    "github.com/vhvplatform/go-shared/cache"
    "github.com/vhvplatform/go-shared/redis"
)

// Create Redis client
redisClient := redis.NewClient(redis.Config{
    URL: "redis://localhost:6379",
})

// Create Redis cache
distCache := cache.NewRedisCache(redisClient, redis.CacheConfig{
    KeyPrefix:  "myapp",
    DefaultTTL: 15 * time.Minute,
})

// Use cache
distCache.Set(ctx, "key", value, 15*time.Minute)
```

### Two-Tier Cache (L1 + L2)

```go
// Create L1 (local)
l1, _ := cache.NewRistrettoCache(cache.RistrettoConfig{
    MaxCost:     10000,
    NumCounters: 100000,
})

// Create L2 (distributed)
l2 := cache.NewRedisCache(redisClient, redis.CacheConfig{
    KeyPrefix:  "myapp",
    DefaultTTL: 1 * time.Hour,
})

// Create tiered cache
tieredCache := cache.NewTieredCache(l1, l2, nil)

// Get: tries L1 → L2 → backfills L1
tieredCache.Get(ctx, "key", &result)

// Set: writes to L2 → L1 (async)
tieredCache.Set(ctx, "key", value, 1*time.Hour)

// Access specific tiers
localOnly := tieredCache.LocalOnly()
distOnly := tieredCache.DistributedOnly()
```

## Use Cases

### API Gateway - Token Validation

```go
// High-frequency reads, need fast access
tokenCache := cache.NewTieredCache(l1, l2, nil)

// Cache token validation results
tokenCache.Set(ctx, tokenID, tokenData, 15*time.Minute)
```

### User Service - User Profiles

```go
// Frequently accessed users in L1, all users in L2
userCache := cache.NewTieredCache(l1, l2, nil)

// Get user (L1 hit = nanoseconds, L2 hit = milliseconds)
var user UserProfile
userCache.Get(ctx, userID, &user)
```

### Temporary Data - Local Only

```go
// No need for distribution
tempCache := tieredCache.LocalOnly()
tempCache.Set(ctx, "session_temp", data, 1*time.Minute)
```

### Critical Data - Distributed Only

```go
// Need consistency across all nodes
criticalCache := tieredCache.DistributedOnly()
criticalCache.Set(ctx, "config", data, 10*time.Minute)
```

## Error Handling

```go
var result MyType
err := cache.Get(ctx, "key", &result)
if err == cache.ErrCacheMiss {
    // Key not found, fetch from database
    result = fetchFromDB()
    cache.Set(ctx, "key", result, ttl)
} else if err != nil {
    // Other error
    return err
}
```

## Performance

| Operation | L1 (Ristretto) | L2 (Redis) | Tiered (L1 hit) | Tiered (L2 hit) |
|-----------|----------------|------------|-----------------|-----------------|
| Get       | ~10 ns         | ~1-2 ms    | ~10 ns          | ~1-2 ms + async backfill |
| Set       | ~10 ns         | ~1-2 ms    | ~1-2 ms         | ~1-2 ms + async L1 |

## Best Practices

1. **Use Tiered Cache for Hot Data**: Frequently accessed data benefits from L1 speed
2. **Cap L1 TTL**: Prevent stale data by capping L1 TTL (default: 5 minutes)
3. **L2 as Source of Truth**: Always write to L2 first in tiered setup
4. **Monitor Cache Hit Rates**: Track L1/L2 hit rates to optimize configuration
5. **Handle Cache Misses Gracefully**: Always have fallback to database

## Configuration

```yaml
cache:
  local:
    enabled: true
    max_cost: 10000
    num_counters: 100000
  redis:
    enabled: true
    url: redis://localhost:6379
    default_ttl: 15m
    key_prefix: "app"
```
