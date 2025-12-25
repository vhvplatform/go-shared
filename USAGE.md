# Shared Packages Usage Guide

This document demonstrates how to use the new shared packages in your microservices.

## HTTP Client Package (`pkg/httpclient`)

The HTTP client package provides a standardized HTTP client with retry logic and circuit breaker support.

### Basic Usage

```go
import "github.com/longvhv/saas-framework-go/pkg/httpclient"

// Create a simple client
client := httpclient.NewClient(
    httpclient.WithBaseURL("http://user-service:8082"),
    httpclient.WithTimeout(5*time.Second),
)

// Make a GET request
var user User
err := client.Get(ctx, "/api/v1/users/123", &user)
if err != nil {
    log.Error("Failed to get user", zap.Error(err))
}

// Make a POST request
newUser := User{Name: "John Doe", Email: "john@example.com"}
var result User
err = client.Post(ctx, "/api/v1/users", newUser, &result)
```

### With Retry and Circuit Breaker

```go
client := httpclient.NewClient(
    httpclient.WithBaseURL("http://user-service:8082"),
    httpclient.WithTimeout(10*time.Second),
    httpclient.WithRetry(3, time.Second),
    httpclient.WithCircuitBreaker(),
    httpclient.WithHeader("X-Service-Name", "api-gateway"),
)
```

## Middleware Package (`pkg/middleware`)

The middleware package provides common middleware for Gin applications.

### Authentication Middleware

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/longvhv/saas-framework-go/pkg/middleware"
)

r := gin.Default()

// Add correlation ID tracking
r.Use(middleware.CorrelationID())

// Add request logging
r.Use(middleware.Logger(log))

// Add panic recovery
r.Use(middleware.Recovery(log))

// Protected routes
protected := r.Group("/api/v1")
protected.Use(middleware.Auth(jwtSecret))
protected.Use(middleware.TenantScope())

// Admin only routes
admin := protected.Group("/admin")
admin.Use(middleware.RequireRoles("admin"))
admin.GET("/users", handlers.ListAllUsers)
```

### Rate Limiting

```go
// Rate limit by IP address (100 requests/second, burst of 200)
r.Use(middleware.PerIP(100, 200))

// Rate limit by tenant ID (50 requests/second per tenant)
protected.Use(middleware.PerTenant(50, 100))

// Rate limit by user ID
protected.Use(middleware.PerUser(10, 20))

// Custom rate limiting
r.Use(middleware.RateLimit(100, 200, func(c *gin.Context) string {
    // Use API key as rate limit key
    return c.GetHeader("X-API-Key")
}))
```

### Request Timeout

```go
// Add 30 second timeout to all requests
r.Use(middleware.Timeout(30 * time.Second))

// Custom timeout with message
r.Use(middleware.TimeoutWithCustomMessage(
    10*time.Second,
    "Request took too long to process",
))
```

### Metrics Collection

```go
// Create metrics collector
collector := middleware.NewMetricsCollector("api_gateway")
collector.Register()

// Add metrics middleware
r.Use(middleware.Metrics(collector))

// Or use default metrics
r.Use(middleware.DefaultMetrics("api_gateway"))
```

### Request Validation

```go
// Ensure Content-Type header is present for POST/PUT
r.Use(middleware.RequestValidation())

// Require JSON content type
api := r.Group("/api")
api.Use(middleware.RequireJSON())

// Limit request body size (10MB)
r.Use(middleware.RequestSizeLimit(10 * 1024 * 1024))
```

## Response Package (`pkg/response`)

The response package provides standardized API response formats.

### Success Responses

```go
import (
    "github.com/longvhv/saas-framework-go/pkg/response"
)

func GetUser(c *gin.Context) {
    user := // fetch user
    response.Success(c, user)
}

func ListUsers(c *gin.Context) {
    users := // fetch users
    meta := response.NewMeta(1, 20, 100) // page 1, 20 per page, 100 total
    response.SuccessWithMeta(c, users, meta)
}

func CreateUser(c *gin.Context) {
    user := // create user
    response.Created(c, user)
}
```

### Error Responses

```go
func HandleError(c *gin.Context) {
    // Standard error responses
    response.BadRequest(c, "Invalid input")
    response.Unauthorized(c, "Authentication required")
    response.Forbidden(c, "Access denied")
    response.NotFound(c, "User not found")
    response.Conflict(c, "Email already exists")
    response.InternalServerError(c, "Something went wrong")
}

func HandleCustomError(c *gin.Context) {
    // Custom error with details
    response.ErrorWithDetails(c, 422, "VALIDATION_ERROR", 
        "Validation failed", 
        map[string]string{
            "email": "Invalid email format",
            "age": "Must be at least 18",
        })
}
```

## Context Package (`pkg/context`)

The context package provides helpers for managing user context in Gin applications.

### Using User Context

```go
import (
    "github.com/longvhv/saas-framework-go/pkg/context"
)

func SomeHandler(c *gin.Context) {
    // Get full user context
    userCtx, err := context.GetUserContext(c)
    if err != nil {
        response.Unauthorized(c, "User not authenticated")
        return
    }

    log.Info("Request from user", 
        zap.String("user_id", userCtx.UserID),
        zap.String("tenant_id", userCtx.TenantID),
        zap.Strings("roles", userCtx.Roles),
    )

    // Check user role
    if context.HasRole(c, "admin") {
        // Admin logic
    }

    if context.HasAnyRole(c, "admin", "moderator") {
        // Admin or moderator logic
    }

    // Get individual values
    userID := context.GetUserIDFromGin(c)
    tenantID := context.GetTenantIDFromGin(c)
    email := context.GetEmailFromGin(c)
}
```

### Setting User Context

```go
func AuthMiddleware(c *gin.Context) {
    // After validating JWT
    userCtx := &context.UserContext{
        UserID:        claims.UserID,
        TenantID:      claims.TenantID,
        Email:         claims.Email,
        Roles:         claims.Roles,
        CorrelationID: c.GetString("correlation_id"),
    }
    
    context.SetUserContext(c, userCtx)
    c.Next()
}
```

## Complete Example

Here's a complete example combining all packages:

```go
package main

import (
    "context"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/longvhv/saas-framework-go/pkg/httpclient"
    "github.com/longvhv/saas-framework-go/pkg/middleware"
    "github.com/longvhv/saas-framework-go/pkg/response"
    pkgctx "github.com/longvhv/saas-framework-go/pkg/context"
    "github.com/longvhv/saas-framework-go/pkg/logger"
)

func main() {
    // Initialize logger
    log, _ := logger.New(logger.Config{
        Level:      "info",
        Encoding:   "json",
        OutputPath: "stdout",
    })

    // Create HTTP client for user service
    userClient := httpclient.NewClient(
        httpclient.WithBaseURL("http://user-service:8082"),
        httpclient.WithTimeout(5*time.Second),
        httpclient.WithRetry(3, time.Second),
        httpclient.WithCircuitBreaker(),
    )

    // Setup Gin router
    r := gin.New()

    // Global middleware
    r.Use(middleware.CorrelationID())
    r.Use(middleware.Logger(log))
    r.Use(middleware.Recovery(log))
    r.Use(middleware.PerIP(100, 200))
    r.Use(middleware.DefaultMetrics("api_gateway"))

    // Public routes
    r.GET("/health", func(c *gin.Context) {
        response.Success(c, gin.H{"status": "healthy"})
    })

    // Protected API routes
    api := r.Group("/api/v1")
    api.Use(middleware.Auth("jwt-secret"))
    api.Use(middleware.TenantScope())
    api.Use(middleware.PerTenant(50, 100))

    // User routes
    api.GET("/users/:id", func(c *gin.Context) {
        userID := c.Param("id")
        
        // Get user from user service
        var user map[string]interface{}
        err := userClient.Get(c.Request.Context(), "/api/v1/users/"+userID, &user)
        if err != nil {
            response.InternalServerError(c, "Failed to fetch user")
            return
        }

        response.Success(c, user)
    })

    // Admin routes
    admin := api.Group("/admin")
    admin.Use(middleware.RequireRoles("admin"))
    
    admin.GET("/stats", func(c *gin.Context) {
        // Only admins can access this
        stats := map[string]interface{}{
            "total_users": 100,
            "active_users": 75,
        }
        response.Success(c, stats)
    })

    // Start server
    r.Run(":8080")
}
```

## Testing with New Packages

### Testing HTTP Client

```go
func TestHTTPClient(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    }))
    defer server.Close()

    client := httpclient.NewClient(httpclient.WithBaseURL(server.URL))
    
    var result map[string]string
    err := client.Get(context.Background(), "/test", &result)
    
    assert.NoError(t, err)
    assert.Equal(t, "ok", result["status"])
}
```

### Testing Middleware

```go
func TestAuthMiddleware(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    r := gin.New()
    r.Use(middleware.Auth("test-secret"))
    r.GET("/protected", func(c *gin.Context) {
        c.String(200, "success")
    })

    // Test without token
    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/protected", nil)
    r.ServeHTTP(w, req)
    assert.Equal(t, http.StatusUnauthorized, w.Code)

    // Test with valid token
    w = httptest.NewRecorder()
    req = httptest.NewRequest("GET", "/protected", nil)
    req.Header.Set("Authorization", "Bearer "+validToken)
    r.ServeHTTP(w, req)
    assert.Equal(t, http.StatusOK, w.Code)
}
```

## Best Practices

1. **Always use correlation IDs** - Add `middleware.CorrelationID()` as the first middleware
2. **Add proper logging** - Use `middleware.Logger()` to track all requests
3. **Implement panic recovery** - Use `middleware.Recovery()` to handle panics gracefully
4. **Set appropriate timeouts** - Use `middleware.Timeout()` to prevent long-running requests
5. **Rate limit appropriately** - Use per-IP or per-tenant rate limiting based on your needs
6. **Use circuit breakers** - Enable circuit breakers in HTTP clients for external service calls
7. **Standardize responses** - Always use the response package for consistent API responses
8. **Check permissions** - Use `context.HasRole()` to verify user permissions
9. **Add metrics** - Use `middleware.Metrics()` to collect performance data
10. **Validate inputs** - Use `middleware.RequestValidation()` to ensure proper request format
