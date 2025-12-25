# Redis Cache Package

A comprehensive, production-ready Redis caching package for Go with advanced features including distributed locking, batch operations, and the Remember pattern.

## Features

✅ **Enhanced Cache Client** - Full-featured caching with sensible defaults  
✅ **Multi-Operations** - Efficient batch get/set/delete operations  
✅ **Remember Pattern** - Cache-or-compute functionality  
✅ **Distributed Locks** - Safe distributed locking with Lua scripts  
✅ **Counter Operations** - Atomic increment/decrement operations  
✅ **Pattern-Based Deletion** - Safe deletion using SCAN  
✅ **TTL Management** - Comprehensive expiration control  
✅ **Key Management** - Advanced key operations  
✅ **Namespace Support** - Hierarchical key prefixing  
✅ **Pluggable Serialization** - JSON and string serializers  
✅ **Thread-Safe** - Safe for concurrent use  
✅ **Context Support** - Full context cancellation support  

## Installation

```bash
go get github.com/vhvcorp/go-shared/redis
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
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
        panic(err)
    }
    defer client.Close()

    // Create cache with configuration
    cache := redis.NewCache(client, redis.CacheConfig{
        DefaultTTL: 5 * time.Minute,
        KeyPrefix:  "myapp",
        Serializer: redis.NewJSONSerializer(),
    })

    ctx := context.Background()

    // Simple set and get
    cache.Set(ctx, "user:1", map[string]string{"name": "John"}, 0)
    
    var user map[string]string
    cache.Get(ctx, "user:1", &user)
    
    fmt.Printf("User: %+v\n", user)
}
```

## Configuration

### CacheConfig

```go
type CacheConfig struct {
    DefaultTTL time.Duration // Default expiration time (default: 1 hour)
    KeyPrefix  string        // Namespace prefix for all keys
    Serializer Serializer    // Serializer implementation (default: JSONSerializer)
}
```

### Creating a Cache

```go
// With defaults
cache := redis.NewCache(client, redis.CacheConfig{})

// With custom configuration
cache := redis.NewCache(client, redis.CacheConfig{
    DefaultTTL: 10 * time.Minute,
    KeyPrefix:  "myapp",
    Serializer: redis.NewJSONSerializer(),
})
```

## Basic Operations

### Set and Get

```go
ctx := context.Background()

// Set with default TTL
err := cache.Set(ctx, "key", "value", 0)

// Set with custom TTL
err := cache.Set(ctx, "key", "value", 5*time.Minute)

// Get value
var value string
err := cache.Get(ctx, "key", &value)
```

### Delete

```go
// Delete single key
err := cache.Delete(ctx, "key")
```

### Key Existence

```go
// Check if keys exist (returns count)
count, err := cache.Exists(ctx, "key1", "key2", "key3")
```

## Multi-Operations

Perform batch operations efficiently:

### MGet - Batch Get

```go
// Get multiple keys at once
results, err := cache.MGet(ctx, "user:1", "user:2", "user:3")
for key, value := range results {
    fmt.Printf("%s = %v\n", key, value)
}
```

### MSet - Batch Set

```go
// Set multiple keys at once
items := map[string]interface{}{
    "user:1": "John",
    "user:2": "Jane",
    "user:3": "Bob",
}
err := cache.MSet(ctx, items, 5*time.Minute)
```

### MDelete - Batch Delete

```go
// Delete multiple keys at once
err := cache.MDelete(ctx, "user:1", "user:2", "user:3")
```

## Remember Pattern (Cache-or-Compute)

The Remember pattern automatically handles cache misses by computing and storing values:

```go
// Remember with TTL
value, err := cache.Remember(ctx, "expensive:calculation", 10*time.Minute, func() (interface{}, error) {
    // This function only runs on cache miss
    result := performExpensiveCalculation()
    return result, nil
})

// Remember forever (no expiration)
value, err := cache.RememberForever(ctx, "config:settings", func() (interface{}, error) {
    return loadConfigFromDatabase(), nil
})
```

### Use Cases

1. **Database Query Caching**
```go
func GetUser(ctx context.Context, cache *redis.Cache, userID string) (*User, error) {
    key := fmt.Sprintf("user:%s", userID)
    
    result, err := cache.Remember(ctx, key, 5*time.Minute, func() (interface{}, error) {
        return db.QueryUser(userID)
    })
    
    return result.(*User), err
}
```

2. **API Response Caching**
```go
func GetWeather(ctx context.Context, cache *redis.Cache, city string) (string, error) {
    key := fmt.Sprintf("weather:%s", city)
    
    result, err := cache.Remember(ctx, key, 30*time.Minute, func() (interface{}, error) {
        return weatherAPI.Fetch(city)
    })
    
    return result.(string), err
}
```

## Pattern-Based Deletion

Safely delete keys matching patterns using SCAN (not KEYS):

```go
// Delete all user session keys
count, err := cache.DeleteByPattern(ctx, "session:*")
fmt.Printf("Deleted %d keys\n", count)

// Delete all keys with the cache prefix
err := cache.FlushPrefix(ctx)
```

### Example Patterns

```go
// Delete all user cache entries
cache.DeleteByPattern(ctx, "user:*")

// Delete specific user sessions
cache.DeleteByPattern(ctx, "session:user:123:*")

// Delete all temporary cache entries
cache.DeleteByPattern(ctx, "temp:*")
```

## Counter Operations

Atomic counter operations for rate limiting, statistics, etc:

```go
// Increment by 1
count, err := cache.Increment(ctx, "page:views")

// Increment by N
count, err := cache.IncrementBy(ctx, "api:calls", 10)

// Decrement by 1
count, err := cache.Decrement(ctx, "stock:items")

// Decrement by N
count, err := cache.DecrementBy(ctx, "credits", 5)

// Increment by float
score, err := cache.IncrementFloat(ctx, "user:score", 2.5)
```

### Rate Limiting Example

```go
func CheckRateLimit(ctx context.Context, cache *redis.Cache, userID string) (bool, error) {
    key := fmt.Sprintf("ratelimit:%s", userID)
    
    count, err := cache.Increment(ctx, key)
    if err != nil {
        return false, err
    }
    
    if count == 1 {
        // Set TTL on first request
        cache.Expire(ctx, key, 1*time.Minute)
    }
    
    return count <= 100, nil // 100 requests per minute
}
```

## Distributed Locking

Implement safe distributed locking with automatic expiration:

### Basic Lock Usage

```go
// Create a lock
lock := cache.Lock("resource:1", 10*time.Second)

// Acquire lock with timeout
err := lock.Acquire(ctx, 5*time.Second)
if err != nil {
    if err == redis.ErrLockNotAcquired {
        // Lock is held by another process
    }
    return err
}
defer lock.Release(ctx)

// Protected code here
processResource()
```

### WithLock Helper

```go
// Execute function with automatic lock management
err := cache.WithLock(ctx, "resource:1", 10*time.Second, func() error {
    // This code runs with lock held
    return processResource()
})
```

### Lock Extension

```go
lock := cache.Lock("long:task", 10*time.Second)
err := lock.Acquire(ctx, 1*time.Second)
if err != nil {
    return err
}
defer lock.Release(ctx)

// Extend lock while processing
for item := range items {
    processItem(item)
    
    // Extend lock every iteration
    lock.Extend(ctx, 10*time.Second)
}
```

### Auto-Refresh Lock

```go
lock := cache.Lock("task:1", 10*time.Second)
err := lock.Acquire(ctx, 1*time.Second)
if err != nil {
    return err
}
defer lock.Release(ctx)

// Start auto-refresh
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

errChan := lock.RefreshLoop(ctx, 5*time.Second)

// Do work
performLongTask()

// Check for refresh errors
select {
case err := <-errChan:
    if err != nil {
        log.Printf("Lock refresh error: %v", err)
    }
default:
}
```

## TTL Management

Comprehensive expiration control:

```go
// Get remaining TTL
ttl, err := cache.GetTTL(ctx, "key")

// Set expiration
err := cache.Expire(ctx, "key", 5*time.Minute)

// Expire at specific time
expireTime := time.Now().Add(24 * time.Hour)
err := cache.ExpireAt(ctx, "key", expireTime)

// Remove expiration (make key persistent)
err := cache.Persist(ctx, "key")

// Refresh TTL to default
err := cache.Touch(ctx, "key")
```

## Key Management

### List Keys

```go
// Get all keys matching pattern (uses SCAN internally)
keys, err := cache.Keys(ctx, "user:*")

// Iterate over keys safely
iterator := cache.Scan(ctx, "session:*", 100)
defer iterator.Close()

for iterator.Next(ctx) {
    key := iterator.Key()
    fmt.Println(key)
}

if err := iterator.Err(); err != nil {
    log.Fatal(err)
}
```

### Rename Keys

```go
// Rename a key
err := cache.Rename(ctx, "old:key", "new:key")
```

## Namespace Support

Create isolated cache namespaces:

```go
// Create main cache
cache := redis.NewCache(client, redis.CacheConfig{
    KeyPrefix: "myapp",
})

// Create sub-caches with additional prefixes
userCache := cache.WithPrefix("users")     // Keys: myapp:users:*
sessionCache := cache.WithPrefix("sessions") // Keys: myapp:sessions:*
tempCache := cache.WithPrefix("temp")       // Keys: myapp:temp:*

// Each sub-cache is isolated
userCache.Set(ctx, "1", userData, 0)      // Stored as: myapp:users:1
sessionCache.Set(ctx, "1", sessionData, 0) // Stored as: myapp:sessions:1
```

### Multi-Tenant Example

```go
func GetTenantCache(cache *redis.Cache, tenantID string) *redis.Cache {
    return cache.WithPrefix(fmt.Sprintf("tenant:%s", tenantID))
}

// Usage
tenant1Cache := GetTenantCache(cache, "tenant-1")
tenant2Cache := GetTenantCache(cache, "tenant-2")

// Completely isolated
tenant1Cache.Set(ctx, "setting", "value1", 0)
tenant2Cache.Set(ctx, "setting", "value2", 0)
```

## Serialization

### JSON Serializer (Default)

```go
serializer := redis.NewJSONSerializer()

cache := redis.NewCache(client, redis.CacheConfig{
    Serializer: serializer,
})

// Works with any JSON-serializable type
type User struct {
    Name  string
    Email string
}

user := User{Name: "John", Email: "john@example.com"}
cache.Set(ctx, "user:1", user, 0)

var retrieved User
cache.Get(ctx, "user:1", &retrieved)
```

### String Serializer

```go
serializer := redis.NewStringSerializer()

cache := redis.NewCache(client, redis.CacheConfig{
    Serializer: serializer,
})

// Only works with strings and []byte
cache.Set(ctx, "message", "Hello, World!", 0)
```

## Performance Considerations

### Batch Operations

Use multi-operations for better performance:

```go
// ❌ Slow: Multiple round trips
for id := range userIDs {
    cache.Get(ctx, fmt.Sprintf("user:%s", id), &user)
}

// ✅ Fast: Single round trip
keys := make([]string, len(userIDs))
for i, id := range userIDs {
    keys[i] = fmt.Sprintf("user:%s", id)
}
results, err := cache.MGet(ctx, keys...)
```

### Pattern Deletion

Always use `DeleteByPattern` with SCAN, never use KEYS in production:

```go
// ✅ Safe: Uses SCAN internally
count, err := cache.DeleteByPattern(ctx, "temp:*")

// ❌ Dangerous in production: Blocks Redis
// keys, _ := client.Keys(ctx, "temp:*")
```

### Lock Timeouts

Choose appropriate lock timeouts:

```go
// Short operations
cache.WithLock(ctx, "counter:update", 1*time.Second, updateCounter)

// Long operations with refresh
lock := cache.Lock("batch:process", 30*time.Second)
lock.Acquire(ctx, 5*time.Second)
defer lock.Release(ctx)

ctx, cancel := context.WithCancel(context.Background())
defer cancel()
errChan := lock.RefreshLoop(ctx, 10*time.Second)
```

## Best Practices

### 1. Always Use Context

```go
// ✅ Good: Use context for cancellation
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := cache.Set(ctx, "key", "value", 0)
```

### 2. Handle Cache Misses

```go
// ✅ Good: Use Remember pattern
value, err := cache.Remember(ctx, "key", ttl, compute)

// Or handle explicitly
var value string
err := cache.Get(ctx, "key", &value)
if err == redis.Nil {
    value = computeValue()
    cache.Set(ctx, "key", value, ttl)
}
```

### 3. Use Appropriate TTLs

```go
// Frequently changing data: short TTL
cache.Set(ctx, "stock:price", price, 30*time.Second)

// Stable data: longer TTL
cache.Set(ctx, "user:profile", profile, 1*time.Hour)

// Configuration: very long or no TTL
cache.Set(ctx, "app:config", config, 24*time.Hour)
```

### 4. Namespace Your Keys

```go
// ✅ Good: Use prefixes for organization
userCache := cache.WithPrefix("users")
sessionCache := cache.WithPrefix("sessions")

// ✅ Good: Structured key names
cache.Set(ctx, "user:123:profile", profile, 0)
cache.Set(ctx, "user:123:settings", settings, 0)
```

### 5. Lock Safety

```go
// ✅ Good: Always defer Release
lock := cache.Lock("resource", 10*time.Second)
err := lock.Acquire(ctx, 5*time.Second)
if err != nil {
    return err
}
defer lock.Release(ctx)

// ✅ Good: Use WithLock for simpler code
cache.WithLock(ctx, "resource", 10*time.Second, func() error {
    return processResource()
})
```

## Common Patterns

### Cache-Aside Pattern

```go
func GetData(ctx context.Context, cache *redis.Cache, id string) (*Data, error) {
    key := fmt.Sprintf("data:%s", id)
    
    // Try cache first
    var data Data
    err := cache.Get(ctx, key, &data)
    if err == nil {
        return &data, nil
    }
    
    // Cache miss: load from database
    data, err = db.Load(id)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    cache.Set(ctx, key, data, 5*time.Minute)
    
    return &data, nil
}
```

### Session Management

```go
type SessionManager struct {
    cache *redis.Cache
}

func (sm *SessionManager) Create(ctx context.Context, userID string) (string, error) {
    sessionID := generateSessionID()
    key := fmt.Sprintf("session:%s", sessionID)
    
    session := Session{
        UserID:    userID,
        CreatedAt: time.Now(),
    }
    
    err := sm.cache.Set(ctx, key, session, 24*time.Hour)
    return sessionID, err
}

func (sm *SessionManager) Get(ctx context.Context, sessionID string) (*Session, error) {
    key := fmt.Sprintf("session:%s", sessionID)
    
    var session Session
    err := sm.cache.Get(ctx, key, &session)
    if err != nil {
        return nil, err
    }
    
    // Refresh session TTL
    sm.cache.Touch(ctx, key)
    
    return &session, nil
}
```

### Distributed Task Coordination

```go
func ProcessTask(ctx context.Context, cache *redis.Cache, taskID string) error {
    lockKey := fmt.Sprintf("task:lock:%s", taskID)
    
    return cache.WithLock(ctx, lockKey, 5*time.Minute, func() error {
        // Only one worker processes this task
        task := loadTask(taskID)
        
        if task.Status == "completed" {
            return nil // Already processed
        }
        
        // Process task
        result := processTask(task)
        
        // Mark as completed
        task.Status = "completed"
        task.Result = result
        
        return saveTask(task)
    })
}
```

## Error Handling

```go
import "github.com/redis/go-redis/v9"

// Check for specific errors
err := cache.Get(ctx, "key", &value)
if err == redis.Nil {
    // Key doesn't exist
} else if err != nil {
    // Other error
}

// Lock errors
err := lock.Acquire(ctx, 1*time.Second)
if err == redis.ErrLockNotAcquired {
    // Couldn't acquire lock
} else if err != nil {
    // Other error
}

err := lock.Release(ctx)
if err == redis.ErrLockNotHeld {
    // Lock wasn't held or expired
}
```

## Thread Safety

All cache operations are thread-safe and can be used concurrently:

```go
cache := redis.NewCache(client, config)

// Safe to use from multiple goroutines
for i := 0; i < 10; i++ {
    go func(id int) {
        cache.Set(ctx, fmt.Sprintf("key:%d", id), id, 0)
    }(i)
}
```

## Testing

The package includes comprehensive tests using miniredis:

```bash
go test -v ./pkg/redis/...
go test -race ./pkg/redis/...
go test -cover ./pkg/redis/...
```

## Examples

See the `examples/` directory for complete working examples:

- `basic_cache.go` - Basic caching operations
- `remember_pattern.go` - Cache-or-compute examples
- `distributed_lock.go` - Distributed locking patterns
- `rate_limiting.go` - Rate limiting implementation
- `multi_tenant.go` - Multi-tenant cache isolation

## License

This package is part of the saas-framework-go project.
