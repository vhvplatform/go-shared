# SaaS Shared Go Library

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)](https://golang.org/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/vhvcorp/go-shared)](https://goreportcard.com/report/github.com/vhvcorp/go-shared)

A comprehensive collection of reusable Go packages for building multi-tenant SaaS applications. This library provides essential building blocks including authentication, authorization, context management, middleware, database utilities, and more.

## Installation

```bash
go get github.com/vhvcorp/go-shared@latest
```

## Quick Start

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/vhvcorp/go-shared/middleware"
    "github.com/vhvcorp/go-shared/response"
    "github.com/vhvcorp/go-shared/tenant"
)

func main() {
    router := gin.Default()
    
    // Setup tenant resolver
    tenantResolver := tenant.NewResolver(tenant.ResolverConfig{
        Strategies: []tenant.ResolutionStrategy{
            tenant.StrategyHeader,
            tenant.StrategySubdomain,
        },
    })
    
    // Apply middleware
    router.Use(middleware.ContextMiddleware())
    router.Use(tenantResolver.Middleware())
    
    router.GET("/hello", func(c *gin.Context) {
        response.Success(c, map[string]string{"message": "Hello, World!"})
    })
    
    router.Run(":8080")
}
```

## Features

- 🔐 **Authentication & Authorization** - JWT-based auth with permission and role management
- 🏢 **Multi-Tenancy** - Flexible tenant resolution strategies (header, subdomain, domain)
- 📝 **Context Management** - Request-scoped data with correlation ID tracking
- 🔌 **Middleware Collection** - Ready-to-use Gin middleware for common tasks
- 🗄️ **Database Utilities** - MongoDB helpers with tenant isolation and transactions
- 💾 **Caching** - Redis client with multi-tenant support and common patterns
- 📨 **Messaging** - RabbitMQ integration
- 📊 **Structured Logging** - Zap-based logger with context support
- ✅ **Validation** - Request validation utilities
- 🔄 **HTTP Client** - Configurable HTTP client with retry and timeout
- 📦 **Standard Responses** - Consistent JSON API responses

This library contains reusable packages that are shared across all microservices in the SaaS framework.

## Packages

### 1. `context` - Request Context Management

Provides context management for handling request-scoped data across the application.

**Features:**
- Store and retrieve user information (ID, email, tenant ID)
- Manage roles and permissions
- Track correlation IDs for distributed tracing
- Tenant domain resolution
- Integration with Gin framework

**Example Usage:**
```go
import (
    pkgctx "github.com/vhvcorp/go-shared/context"
)

// Create a request context
rc := &pkgctx.RequestContext{
    UserID:   "user123",
    TenantID: "tenant456",
    Email:    "user@example.com",
    Roles:    []string{"admin"},
}

// Store in standard context
ctx := pkgctx.WithRequestContext(context.Background(), rc)

// Store in Gin context
pkgctx.ToGinContext(c, rc)

// Retrieve values
userID, err := pkgctx.GetUserID(ctx)
tenantID, err := pkgctx.GetTenantID(ctx)
```

### 2. `auth` - Authentication & Authorization Helpers

Provides utilities for checking permissions and roles, and retrieving current user information.

**Features:**
- Permission checking with wildcard support (`users.*`)
- Role-based access control
- User information retrieval
- Global permission checker instance

**Example Usage:**
```go
import (
    "github.com/vhvcorp/go-shared/auth"
)

// Check permissions
if auth.HasPermission(ctx, "users.read") {
    // Allow access
}

// Require permission (returns error)
if err := auth.RequirePermission(ctx, "users.write"); err != nil {
    // Permission denied
}

// Check roles
if auth.IsSuperAdmin(ctx) {
    // Super admin access
}

// Get current user
user, err := auth.GetCurrentUser(ctx)
```

### 3. `tenant` - Tenant Resolution

Provides multi-strategy tenant resolution from HTTP requests.

**Supported Strategies:**
- `StrategyHeader` - Extract tenant from HTTP header (default: `X-Tenant-ID`)
- `StrategySubdomain` - Extract tenant from subdomain (e.g., `tenant.example.com`)
- `StrategyDomain` - Extract tenant from custom domain
- `StrategyParam` - Extract tenant from query/URL parameter

**Example Usage:**
```go
import (
    "github.com/vhvcorp/go-shared/tenant"
)

// Create resolver with multiple strategies
resolver := tenant.NewResolver(tenant.ResolverConfig{
    Strategies: []tenant.ResolutionStrategy{
        tenant.StrategyHeader,
        tenant.StrategySubdomain,
    },
    HeaderName: "X-Tenant-ID",
})

// Use as middleware
router.Use(resolver.Middleware())

// Manual resolution
tenantID, domain, err := resolver.Resolve(c)
```

### 4. `middleware` - Gin Middleware

Provides ready-to-use middleware for common functionalities.

**Available Middleware:**
- `ContextMiddleware()` - Generates correlation IDs
- `AppContextMiddleware()` - Extracts application ID
- `RequestContextMiddleware()` - Builds full request context
- `RequirePermission(permission)` - Requires specific permission
- `RequireAnyPermission(permissions...)` - Requires any of the permissions
- `RequireRole(role)` - Requires specific role
- `RequireSuperAdmin()` - Requires super admin role
- `RequireTenantAdmin()` - Requires tenant admin role

**Example Usage:**
```go
import (
    "github.com/vhvcorp/go-shared/middleware"
)

// Setup middleware chain
router.Use(middleware.ContextMiddleware())
router.Use(middleware.AppContextMiddleware())

// Protected routes
protected := router.Group("")
protected.Use(AuthMiddleware()) // Your auth middleware
protected.Use(middleware.RequestContextMiddleware())

// Permission-protected endpoint
protected.POST("/users",
    middleware.RequirePermission("users.create"),
    userHandler.CreateUser,
)

// Admin-only endpoint
admin := protected.Group("/admin")
admin.Use(middleware.RequireTenantAdmin())
admin.GET("/settings", settingsHandler.GetSettings)
```

### 5. `response` - Standard API Responses

Provides consistent JSON response formatting across all services.

**Features:**
- Success responses with optional metadata
- Error responses with error codes and details
- Pagination metadata support
- Automatic correlation ID inclusion

**Example Usage:**
```go
import (
    "github.com/vhvcorp/go-shared/response"
)

// Success response
response.Success(c, userData)

// Success with pagination
meta := response.NewMeta(page, perPage, total)
response.SuccessWithMeta(c, users, meta)

// Created response
response.Created(c, newUser)

// Error responses
response.BadRequest(c, "Invalid input")
response.Unauthorized(c, "Authentication required")
response.Forbidden(c, "Insufficient permissions")
response.NotFound(c, "User not found")
response.InternalServerError(c, "Something went wrong")

// Custom error with details
details := map[string]string{"field": "email"}
response.ErrorWithDetails(c, 400, "VALIDATION_ERROR", "Invalid email", details)
```

### Response Format

All responses follow this structure:

**Success Response:**
```json
{
  "success": true,
  "data": {...},
  "meta": {
    "page": 1,
    "per_page": 10,
    "total": 100,
    "total_pages": 10
  },
  "correlation_id": "uuid"
}
```

**Error Response:**
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Error message",
    "details": {...}
  },
  "correlation_id": "uuid"
}
```

## Integration Example

Here's a complete example integrating all packages:

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/vhvcorp/go-shared/auth"
    pkgctx "github.com/vhvcorp/go-shared/context"
    "github.com/vhvcorp/go-shared/middleware"
    "github.com/vhvcorp/go-shared/response"
    "github.com/vhvcorp/go-shared/tenant"
)

func main() {
    router := gin.Default()

    // Setup tenant resolver
    tenantResolver := tenant.NewResolver(tenant.ResolverConfig{
        Strategies: []tenant.ResolutionStrategy{
            tenant.StrategyHeader,
            tenant.StrategySubdomain,
        },
    })

    // Global middleware
    router.Use(middleware.ContextMiddleware())
    router.Use(middleware.AppContextMiddleware())
    router.Use(tenantResolver.Middleware())

    // API routes
    api := router.Group("/api/v1")

    // Protected routes
    protected := api.Group("")
    protected.Use(AuthMiddleware())
    protected.Use(middleware.RequestContextMiddleware())

    // User management
    users := protected.Group("/users")
    users.GET("", middleware.RequirePermission("users.read"), ListUsers)
    users.POST("", middleware.RequirePermission("users.create"), CreateUser)
    users.PUT("/:id", middleware.RequirePermission("users.update"), UpdateUser)
    users.DELETE("/:id", middleware.RequirePermission("users.delete"), DeleteUser)

    // Admin routes
    admin := protected.Group("/admin")
    admin.Use(middleware.RequireTenantAdmin())
    admin.GET("/settings", GetSettings)

    router.Run(":8080")
}

func ListUsers(c *gin.Context) {
    ctx := c.Request.Context()

    // Check authentication
    if !auth.IsAuthenticated(ctx) {
        response.Unauthorized(c, "Authentication required")
        return
    }

    // Get current user info
    currentUser, err := auth.GetCurrentUser(ctx)
    if err != nil {
        response.Unauthorized(c, "Invalid user")
        return
    }

    // Business logic here...
    users := []User{} // Fetch users from DB

    // Return paginated response
    meta := response.NewMeta(1, 10, 100)
    response.SuccessWithMeta(c, users, meta)
}

func CreateUser(c *gin.Context) {
    ctx := c.Request.Context()

    // Get tenant ID
    tenantID, err := auth.GetCurrentTenantID(ctx)
    if err != nil {
        response.BadRequest(c, "Tenant context required")
        return
    }

    // Create user logic...

    response.Created(c, newUser)
}
```

## Testing

Run tests with:

```bash
# Test all packages
go test ./...

# Test specific package
go test ./context/
go test ./auth/
go test ./tenant/

# Test with coverage
go test -coverprofile=coverage.txt -covermode=atomic ./...
```

## Versioning

This library follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions  
- **PATCH** version for backwards-compatible bug fixes

To use a specific version:
```bash
go get github.com/vhvcorp/go-shared@v1.0.0
```

## Dependencies

Key dependencies:
- `github.com/gin-gonic/gin` - HTTP web framework
- `github.com/golang-jwt/jwt/v5` - JWT implementation
- `go.mongodb.org/mongo-driver` - MongoDB driver
- `github.com/redis/go-redis/v9` - Redis client
- `go.uber.org/zap` - Structured logging
- `github.com/spf13/viper` - Configuration management

See [go.mod](go.mod) for complete dependency list.

## Best Practices

1. **Always use context propagation** - Pass context through all layers of your application
2. **Check permissions early** - Use middleware for route-level checks, helper functions for business logic
3. **Use consistent response format** - Always use the response package for API responses
4. **Include correlation IDs** - Use ContextMiddleware to track requests across services
5. **Validate tenant context** - Always ensure tenant context is available for multi-tenant operations
6. **Test your code** - All packages include test examples you can follow

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For issues and questions:
- Open an issue on [GitHub](https://github.com/vhvcorp/go-shared/issues)
- Refer to the [documentation](README.md)

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and changes.
