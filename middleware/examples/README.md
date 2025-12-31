# Middleware Examples

This directory contains example applications demonstrating the usage of security middleware provided by the go-shared library.

## Prerequisites

- Go 1.21 or higher
- Redis server running on localhost:6379
- Basic understanding of Gin framework

## Examples

### 1. Rate Limiting (`ratelimit_example.go`)

Demonstrates different rate limiting strategies:
- **Per-IP limiting**: Limits requests based on client IP address
- **Per-Tenant limiting**: Limits requests based on tenant ID
- **Custom limiting**: Limits requests based on custom keys (e.g., API key)

**Running the example:**
```bash
go run ratelimit_example.go
```

**Testing:**
```bash
# Test rate limiting - make multiple requests quickly
for i in {1..15}; do
  curl http://localhost:8080/api/data
  echo ""
done
```

After the burst limit is exceeded, you should see 429 errors.

### 2. Replay Attack Protection (`replay_example.go`)

Demonstrates two types of replay protection:
- **Nonce-based**: Requires clients to send unique nonces with timestamps
- **Hash-based**: Prevents duplicate requests automatically

**Running the example:**
```bash
go run replay_example.go
```

**Testing nonce-based protection:**
```bash
# Generate a nonce (use UUID in production)
NONCE=$(uuidgen)
TIMESTAMP=$(date +%s)

# First request succeeds
curl -X POST http://localhost:8080/api/v1/transaction \
  -H "X-Request-Nonce: $NONCE" \
  -H "X-Request-Timestamp: $TIMESTAMP"

# Second request with same nonce fails (replay detected)
curl -X POST http://localhost:8080/api/v1/transaction \
  -H "X-Request-Nonce: $NONCE" \
  -H "X-Request-Timestamp: $TIMESTAMP"
```

**Testing hash-based protection:**
```bash
# First request succeeds
curl -X POST http://localhost:8080/api/idempotent/order \
  -H "Content-Type: application/json" \
  -d '{"product_id": "123", "quantity": 1}'

# Duplicate request fails
curl -X POST http://localhost:8080/api/idempotent/order \
  -H "Content-Type: application/json" \
  -d '{"product_id": "123", "quantity": 1}'
```

### 3. Brute Force Protection (`bruteforce_example.go`)

Demonstrates three approaches to brute force protection:
- **IP-based**: Tracks failed attempts by IP address
- **Username-based**: Tracks failed attempts by username
- **Email-based with exponential backoff**: Increases lockout duration exponentially

Also includes admin endpoints for managing protection.

**Running the example:**
```bash
go run bruteforce_example.go
```

**Testing basic protection:**
```bash
# Make failed login attempts
for i in {1..6}; do
  curl -X POST http://localhost:8080/login-basic \
    -d "username=admin&password=wrong"
  echo ""
done
```

After 5 failed attempts, you should see a 429 error (locked out).

**Testing username-based protection:**
```bash
# Fail login for user1
for i in {1..6}; do
  curl -X POST http://localhost:8080/login-user \
    -d "username=user1&password=wrong"
  echo ""
done

# Different user can still try
curl -X POST http://localhost:8080/login-user \
  -d "username=user2&password=wrong"
```

**Testing successful login (resets counter):**
```bash
# Make 2 failed attempts
curl -X POST http://localhost:8080/login-user \
  -d "username=admin&password=wrong"

curl -X POST http://localhost:8080/login-user \
  -d "username=admin&password=wrong"

# Successful login resets counter
curl -X POST http://localhost:8080/login-user \
  -d "username=admin&password=secret"

# Can make attempts again
curl -X POST http://localhost:8080/login-user \
  -d "username=admin&password=wrong"
```

**Admin operations:**
```bash
# Check brute force status
curl "http://localhost:8080/admin/bruteforce-status?identifier=192.168.1.1"

# Reset brute force protection
curl -X POST "http://localhost:8080/admin/reset-bruteforce?identifier=192.168.1.1"
```

## Combined Example

Here's how to use all security middleware together:

```go
package main

import (
    "context"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
    "github.com/vhvplatform/go-shared/middleware"
)

func main() {
    router := gin.Default()
    
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // Global rate limiting
    router.Use(middleware.PerIP(100, 200))
    
    // Authentication endpoints with brute force protection
    auth := router.Group("/auth")
    auth.Use(middleware.BruteForcePerUser(middleware.BruteForceProtectionConfig{
        RedisClient:     redisClient,
        MaxAttempts:     5,
        LockoutDuration: 15 * time.Minute,
    }, "username"))
    auth.POST("/login", handleLogin)
    
    // Protected API with replay protection
    api := router.Group("/api")
    api.Use(middleware.ReplayProtection(middleware.ReplayProtectionConfig{
        RedisClient: redisClient,
        TimeWindow:  5 * time.Minute,
    }))
    api.POST("/transaction", handleTransaction)
    
    router.Run(":8080")
}
```

## Best Practices

1. **Rate Limiting:**
   - Set appropriate limits based on your API capacity
   - Use different limits for different endpoints
   - Consider using PerUser for authenticated endpoints

2. **Replay Protection:**
   - Use nonce-based protection for critical operations (payments, transfers)
   - Use hash-based protection for idempotency
   - Set reasonable time windows (5-15 minutes)
   - Ensure clients generate cryptographically secure nonces

3. **Brute Force Protection:**
   - Apply to authentication endpoints only
   - Use exponential backoff for persistent attackers
   - Implement admin tools to unlock legitimate users
   - Monitor lockout events for security analysis
   - Consider combining with CAPTCHA after N attempts

4. **Redis Configuration:**
   - Use separate Redis database for security features
   - Set up Redis persistence for production
   - Monitor Redis memory usage
   - Consider Redis Cluster for high availability

5. **Error Handling:**
   - Return user-friendly error messages
   - Log security events for monitoring
   - Include retry information in lockout responses
   - Implement alerting for suspicious patterns

## Troubleshooting

**Redis Connection Issues:**
```bash
# Check if Redis is running
redis-cli ping

# Start Redis (if not running)
redis-server

# Or use Docker for easy setup
docker run -d -p 6379:6379 redis:alpine
```

**High Memory Usage:**
```bash
# Check Redis memory
redis-cli info memory

# Clear test data
redis-cli FLUSHDB
```

**Testing Without Redis:**
The examples require Redis. For testing without Redis, you can mock the Redis client or use an in-memory implementation (not included in these examples).

## Production Considerations

1. Use environment variables for configuration
2. Implement proper logging and monitoring
3. Set up Redis clustering for high availability
4. Use SSL/TLS for Redis connections
5. Implement rate limit headers (X-RateLimit-Limit, X-RateLimit-Remaining)
6. Add metrics collection (Prometheus, DataDog, etc.)
7. Implement IP whitelisting for trusted sources
8. Consider using Redis Sentinel for automatic failover
