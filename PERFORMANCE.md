# Performance Optimization Guide

This document outlines the performance optimizations implemented in the go-shared library and best practices for using the library efficiently.

## Optimization Summary

### 1. Context Management Optimization

**Optimization**: Added caching for full `RequestContext` to avoid repeated field lookups.

**Implementation**:
- Added `RequestCtxKey` constant for storing the complete RequestContext
- Modified `WithRequestContext` to cache the full context
- Optimized `GetRequestContext` to return cached context when available

**Performance Impact**:
- `GetRequestContext`: **9.3 ns/op with 0 allocations** (significant improvement!)
- Eliminates multiple context value lookups when retrieving full request context

**Usage**:
```go
// Store request context
rc := &context.RequestContext{
    UserID:   "user123",
    TenantID: "tenant456",
    // ... other fields
}
ctx := context.WithRequestContext(ctx, rc)

// Retrieve (now extremely fast with caching)
rc = context.GetRequestContext(ctx)
```

### 2. Permission Checking Optimization

**Optimization**: Improved string operations and early returns in permission checking.

**Implementation**:
- Added fast path for wildcard admin permission (`*`)
- Optimized exact match checking with early returns
- Improved wildcard permission matching to avoid unnecessary string allocations

**Performance Impact**:
- `HasPermission`: **11.6 ns/op with 0 allocations**
- `HasPermissionWildcard`: **29.0 ns/op with 0 allocations**

**Usage**:
```go
// Permission checking is now faster
if auth.HasPermission(ctx, "users.read") {
    // Allow access
}

// Wildcard matching is optimized
if auth.HasPermission(ctx, "posts.create") { // matches "posts.*"
    // Allow access
}
```

### 3. Rate Limiter Lock Optimization

**Optimization**: Reduced lock contention in rate limiter with optimized locking strategy.

**Implementation**:
- Implemented fast path with read-lock only for existing limiters
- Minimized time spent holding write locks
- Added triple-check pattern to handle race conditions efficiently

**Performance Impact**:
- Reduced lock contention in high-concurrency scenarios
- Improved throughput for rate limiting operations

**Usage**:
```go
// Rate limiting now handles high concurrency better
rateLimiter := middleware.PerIP(100, 10) // 100 rps, 10 burst
router.Use(rateLimiter)
```

### 4. HTTP Client Request Retry Optimization

**Optimization**: Reduced memory allocations in request retry logic.

**Implementation**:
- Read request body once before retry loop instead of on every attempt
- Reuse body bytes across retry attempts
- Eliminated unnecessary allocations in request cloning

**Performance Impact**:
- Reduced memory allocations per retry attempt
- Faster retry operations

**Usage**:
```go
// HTTP client retries are now more efficient
client := httpclient.NewClient(
    httpclient.WithRetry(3, time.Second),
)
err := client.Post(ctx, "/api/resource", body, &result)
```

### 5. String Utility Optimization

**Optimization**: Improved performance of common string operations.

**Implementation**:
- **ToSnakeCase**: Pre-allocated string builder capacity and direct character conversion
- **IsValidEmail**: Cached regex compilation using `sync.Once`

**Performance Impact**:
- `ToSnakeCase`: **93.8 ns/op with 1 allocation**
- `IsValidEmail`: **389.3 ns/op with 0 allocations**

**Usage**:
```go
// String operations are now faster
snakeCase := utils.ToSnakeCase("MyVariableName") // "my_variable_name"
valid := utils.IsValidEmail("user@example.com")  // true
```

### 6. Gin Context Optimization

**Optimization**: Cache RequestContext in Gin context to avoid rebuilding.

**Implementation**:
- Store full RequestContext object in Gin context
- Return cached context from `FromGinContext` when available
- Eliminate 8 individual field lookups on every retrieval

**Performance Impact**:
- `FromGinContext` (cached): **24.9 ns/op with 0 allocations** (10x faster!)
- `FromGinContext` (uncached): **268.7 ns/op with 1 allocation**

**Usage**:
```go
// Store in Gin context (automatically caches)
rc := &context.RequestContext{
    UserID:   "user123",
    TenantID: "tenant456",
    // ... other fields
}
context.ToGinContext(c, rc)

// Retrieve (extremely fast with caching)
rc = context.FromGinContext(c) // Only 24.9 ns!
```

### 7. Response Package Optimization

**Optimization**: Inline helper to avoid repeated correlation_id lookups.

**Implementation**:
- Added `getCorrelationID` helper function
- Reduced repeated Gin context lookups in response functions

**Performance Impact**:
- Faster response generation in high-throughput APIs

### 8. MongoDB Query Builder Optimization

**Optimization**: Pre-allocate maps and slices with reasonable capacities.

**Implementation**:
- Pre-allocate `QueryBuilder` map with capacity 8
- Pre-allocate `AggregationBuilder` slice with capacity 8
- Reuse underlying storage in `Reset` method

**Performance Impact**:
- `NewQueryBuilder`: **8.3 ns/op with 0 allocations**
- Reduced allocations during query building

**Usage**:
```go
// Query builder now pre-allocates storage
qb := mongodb.NewQueryBuilder()
qb.Where("status", "active").
   WhereGreaterThan("created_at", time.Now().Add(-24*time.Hour))
```

### 9. MongoDB Pagination Optimization

**Optimization**: Run count and find operations concurrently, plus fast pagination mode.

**Implementation**:
- Execute count and find in parallel goroutines
- Added `PaginateFast` that skips counting for better performance

**Performance Impact**:
- Faster pagination queries, especially for large collections
- `PaginateFast` avoids expensive count operation entirely

**Usage**:
```go
// Standard pagination (with total count)
result, err := mongodb.Paginate(ctx, collection, filter, params, &users)

// Fast pagination (without total count, faster)
result, err := mongodb.PaginateFast(ctx, collection, filter, params, &users)
```

## Best Practices for Performance

### 1. Use Context Caching

When you need to access multiple fields from the request context, use `GetRequestContext` once instead of calling individual getters:

```go
// ❌ Inefficient - multiple context lookups
userID, _ := context.GetUserID(ctx)
tenantID, _ := context.GetTenantID(ctx)
email := context.GetEmail(ctx)

// ✅ Efficient - single cached lookup
rc := context.GetRequestContext(ctx)
userID := rc.UserID
tenantID := rc.TenantID
email := rc.Email
```

### 2. Use Gin Context Caching

In Gin handlers, always use `FromGinContext` for fastest access:

```go
// ✅ Efficient - uses cached RequestContext
rc := context.FromGinContext(c)
userID := rc.UserID
tenantID := rc.TenantID
```

### 3. Leverage Wildcard Permissions

Use wildcard permissions to reduce permission checks:

```go
// Instead of checking multiple individual permissions
permissions := []string{
    "users.create",
    "users.read",
    "users.update",
    "users.delete",
}

// Use a wildcard permission
permissions := []string{
    "users.*",
}
```

### 5. Optimize Permission Checks

Check for the most likely permissions first to benefit from early returns:

```go
// ✅ Check most common permission first
if auth.HasPermission(ctx, "users.read") {
    // Most common case handled first
}
```

### 6. Reuse Query Builders

Reuse query builders with `Clone` and `Reset`:

```go
// Create a base query builder
baseQB := mongodb.NewQueryBuilder().Where("tenant_id", tenantID)

// Clone for different queries
activeUsers := baseQB.Clone().Where("status", "active")
inactiveUsers := baseQB.Clone().Where("status", "inactive")

// Or reset and reuse
qb := mongodb.NewQueryBuilder()
result1 := qb.Where("status", "active").Build()
qb.Reset() // Reuses underlying storage
result2 := qb.Where("status", "inactive").Build()
```

### 7. Use Fast Build Targets

For faster development iteration:

```bash
# Faster builds with caching
make build-fast

# Faster tests without race detector (use for quick checks)
make test-fast

# Use race detector for thorough testing
make test
```

### 5. Run Benchmarks Regularly

Use benchmarks to validate performance:

```bash
# Run all benchmarks
make bench

# Run specific package benchmarks
go test -bench=. -benchmem ./context/
go test -bench=. -benchmem ./auth/
```

## Benchmark Results

### Context Package
```
BenchmarkWithRequestContext-4     2,236,758     548.0 ns/op     576 B/op    17 allocs/op
BenchmarkGetRequestContext-4    128,431,989       9.3 ns/op       0 B/op     0 allocs/op
BenchmarkGetUserID-4            124,458,459       9.7 ns/op       0 B/op     0 allocs/op
BenchmarkGetPermissions-4       159,208,786       7.7 ns/op       0 B/op     0 allocs/op
```

### Auth Package
```
BenchmarkHasPermission-4            97,646,530    11.6 ns/op    0 B/op    0 allocs/op
BenchmarkHasPermissionWildcard-4    41,385,736    29.0 ns/op    0 B/op    0 allocs/op
BenchmarkHasPermissionMiss-4        35,253,784    34.0 ns/op    0 B/op    0 allocs/op
BenchmarkHasAnyPermission-4         25,458,678    47.1 ns/op    0 B/op    0 allocs/op
BenchmarkHasRole-4                 100,000,000    11.5 ns/op    0 B/op    0 allocs/op
BenchmarkIsSuperAdmin-4             88,958,823    12.5 ns/op    0 B/op    0 allocs/op
```

### Utils Package
```
BenchmarkToSnakeCase-4           12,729,718    93.8 ns/op    24 B/op    1 allocs/op
BenchmarkToSnakeCaseLong-4        5,672,502   215.2 ns/op    48 B/op    1 allocs/op
BenchmarkIsValidEmail-4           3,099,139   389.3 ns/op     0 B/op    0 allocs/op
BenchmarkContains-4             146,407,848     8.6 ns/op     0 B/op    0 allocs/op
```

## Future Optimization Opportunities

1. **Connection Pooling**: Consider implementing connection pooling for database and HTTP clients
2. **Caching Layer**: Add optional caching layer for frequently accessed data
3. **Batch Operations**: Implement batch operations for database queries
4. **Compression**: Add optional compression for HTTP requests/responses
5. **Profiling**: Use pprof to identify additional bottlenecks in production

## Conclusion

These optimizations provide significant performance improvements while maintaining backward compatibility. The most impactful change is the context caching, which reduces `GetRequestContext` from multiple lookups to a single cached retrieval with zero allocations.

For questions or suggestions, please open an issue on GitHub.
