# ClickHouse Package

A comprehensive ClickHouse client library for Go with support for connection management, batch operations, query building, and transaction-like operations.

## Features

- **Connection Management**: Enhanced configuration with connection pooling and timeouts
- **Batch Operations**: Efficient bulk inserts with batch API
- **Query Builder**: Fluent API for building SQL queries
- **Transaction-like Operations**: Grouped query execution (Note: ClickHouse has limited native transaction support)
- **Health Checks**: Built-in health check functionality
- **Context Support**: Full context cancellation support
- **Error Handling**: Comprehensive error handling with wrapped errors

## Installation

```bash
go get github.com/vhvplatform/go-shared/clickhouse
```

## Quick Start

### Creating a Client

```go
import (
    "context"
    "time"
    "github.com/vhvplatform/go-shared/clickhouse"
)

func main() {
    cfg := clickhouse.Config{
        Addr:            []string{"localhost:9000"},
        Database:        "default",
        Username:        "default",
        Password:        "",
        MaxOpenConns:    10,
        MaxIdleConns:    5,
        ConnMaxLifetime: 1 * time.Hour,
        ConnMaxIdleTime: 10 * time.Minute,
        DialTimeout:     10 * time.Second,
        Debug:           false,
        Compression:     true,
    }

    client, err := clickhouse.NewClient(context.Background(), cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
}
```

## Basic Operations

### Execute Queries

```go
ctx := context.Background()

// Create table
err := client.Exec(ctx, `
    CREATE TABLE IF NOT EXISTS users (
        id UInt64,
        name String,
        email String,
        created_at DateTime
    ) ENGINE = MergeTree()
    ORDER BY id
`)

// Insert single row
err = client.Exec(ctx, 
    "INSERT INTO users (id, name, email, created_at) VALUES (?, ?, ?, ?)",
    1, "John Doe", "john@example.com", time.Now(),
)
```

### Query Data

```go
// Define result struct
type User struct {
    ID        uint64    `ch:"id"`
    Name      string    `ch:"name"`
    Email     string    `ch:"email"`
    CreatedAt time.Time `ch:"created_at"`
}

// Query multiple rows
var users []User
err := client.Query(ctx, &users, "SELECT * FROM users WHERE id > ?", 0)

// Query single row
var user User
err := client.QueryRow(ctx, &user, "SELECT * FROM users WHERE id = ?", 1)
```

## Batch Operations

Batch operations are highly efficient for bulk inserts:

### Using BatchInserter

```go
ctx := context.Background()

// Create batch inserter
batch, err := client.NewBatchInserter(ctx, "INSERT INTO users")
if err != nil {
    log.Fatal(err)
}

// Add rows
for i := 0; i < 1000; i++ {
    err := batch.Append(
        uint64(i),
        fmt.Sprintf("User %d", i),
        fmt.Sprintf("user%d@example.com", i),
        time.Now(),
    )
    if err != nil {
        batch.Abort()
        log.Fatal(err)
    }
}

// Send batch
if err := batch.Send(); err != nil {
    log.Fatal(err)
}
```

### Using BatchInsert Helper

```go
rows := [][]interface{}{
    {1, "John Doe", "john@example.com", time.Now()},
    {2, "Jane Smith", "jane@example.com", time.Now()},
    {3, "Bob Johnson", "bob@example.com", time.Now()},
}

columns := []string{"id", "name", "email", "created_at"}

err := client.BatchInsert(ctx, "users", columns, rows)
```

## Query Builder

Build queries using a fluent API:

### SELECT Queries

```go
qb := clickhouse.NewQueryBuilder()

// Simple select
query, args := qb.
    Table("users").
    Select("id", "name", "email").
    Where("id", 1).
    BuildSelect()

var user User
err := client.Query(ctx, &user, query, args...)

// Complex query
query, args = qb.
    Table("orders").
    Select("*").
    Where("status", "active").
    WhereGreaterThan("amount", 100).
    WhereBetween("created_at", startDate, endDate).
    OrderBy("created_at", "DESC").
    Limit(10).
    Offset(0).
    BuildSelect()

var orders []Order
err := client.Query(ctx, &orders, query, args...)
```

### COUNT Queries

```go
qb := clickhouse.NewQueryBuilder()
query, args := qb.
    Table("users").
    Where("status", "active").
    BuildCount()

var count uint64
err := client.QueryRow(ctx, &count, query, args...)
```

### DELETE Queries

```go
qb := clickhouse.NewQueryBuilder()
query, args := qb.
    Table("users").
    Where("id", 123).
    BuildDelete()

err := client.Exec(ctx, query, args...)
```

### Advanced Query Builder Features

```go
qb := clickhouse.NewQueryBuilder()

// WHERE IN
query, args := qb.
    Table("users").
    WhereIn("id", []interface{}{1, 2, 3, 4, 5}).
    BuildSelect()

// LIKE queries
query, args = qb.
    Table("users").
    WhereLike("email", "%@example.com").
    BuildSelect()

// Comparison operators
query, args = qb.
    Table("orders").
    WhereGreaterThan("amount", 100).
    WhereLessThan("amount", 1000).
    BuildSelect()

// Clone for reusability
baseQuery := clickhouse.NewQueryBuilder().
    Table("users").
    Where("status", "active")

activeAdmins := baseQuery.Clone().Where("role", "admin").BuildSelect()
activeUsers := baseQuery.Clone().Where("role", "user").BuildSelect()
```

## Transaction-like Operations

ClickHouse has limited transaction support, but this library provides a way to group operations:

### Using BeginTx

```go
ctx := context.Background()
tx := client.BeginTx(ctx)

// Queue operations
tx.Exec("INSERT INTO users VALUES (?, ?, ?)", 1, "John", "john@example.com")
tx.Exec("INSERT INTO logs VALUES (?, ?)", 1, "User created")

// Commit all operations
if err := tx.Commit(); err != nil {
    log.Fatal(err)
}

// Or rollback (clears queued operations)
tx.Rollback()
```

### Using WithTransaction

```go
err := client.WithTransaction(ctx, func(tx *clickhouse.Transaction) error {
    // Add operations
    if err := tx.Exec("INSERT INTO users VALUES (?, ?, ?)", 1, "John", "john@example.com"); err != nil {
        return err // Will trigger rollback
    }
    
    if err := tx.Exec("INSERT INTO logs VALUES (?, ?)", 1, "User created"); err != nil {
        return err // Will trigger rollback
    }
    
    return nil // Will commit
})
```

**Important Notes:**
- ClickHouse does not support traditional ACID transactions
- These transaction methods queue operations and execute them sequentially
- If an operation fails, subsequent operations won't execute
- This is useful for maintaining operation order and error handling

## Health Checks

```go
// Check connection health
if err := client.HealthCheck(ctx); err != nil {
    log.Printf("Health check failed: %v", err)
}

// Get connection statistics
stats := client.Stats()
fmt.Printf("Open connections: %d\n", stats.Open)
```

## Advanced Features

### Custom Query Timeout

```go
// Create context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

var users []User
err := client.Query(ctx, &users, "SELECT * FROM users")
```

### Working with Different Data Types

```go
type Event struct {
    ID        uint64                 `ch:"id"`
    Name      string                 `ch:"name"`
    Tags      []string              `ch:"tags"`      // Array
    Metadata  map[string]string     `ch:"metadata"`  // Map
    Timestamp time.Time             `ch:"timestamp"`
    Count     int64                 `ch:"count"`
}

// Insert with various types
err := client.Exec(ctx, `
    INSERT INTO events (id, name, tags, metadata, timestamp, count)
    VALUES (?, ?, ?, ?, ?, ?)
`,
    1,
    "user.login",
    []string{"auth", "user"},
    map[string]string{"ip": "192.168.1.1", "browser": "Chrome"},
    time.Now(),
    int64(42),
)
```

### Compression

Compression is enabled by default for better network performance:

```go
// Enable compression (default)
cfg := clickhouse.Config{
    Compression: true, // Uses LZ4 compression
}

// Disable compression if needed
cfg := clickhouse.Config{
    Compression: false,
}
```

## Best Practices

### 1. Use Batch Operations for Bulk Inserts

```go
// ❌ Slow: Individual inserts
for _, user := range users {
    client.Exec(ctx, "INSERT INTO users VALUES (?, ?, ?)", user.ID, user.Name, user.Email)
}

// ✅ Fast: Batch insert
batch, _ := client.NewBatchInserter(ctx, "INSERT INTO users")
for _, user := range users {
    batch.Append(user.ID, user.Name, user.Email)
}
batch.Send()
```

### 2. Use Query Builder for Complex Queries

```go
// ✅ Good: Use query builder
qb := clickhouse.NewQueryBuilder()
query, args := qb.
    Table("users").
    Where("status", "active").
    WhereGreaterThan("created_at", startDate).
    BuildSelect()

client.Query(ctx, &users, query, args...)
```

### 3. Always Use Context

```go
// ✅ Good: Use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := client.Query(ctx, &users, "SELECT * FROM users")
```

### 4. Handle Errors Properly

```go
// ✅ Good: Check and handle errors
if err := client.Exec(ctx, query, args...); err != nil {
    log.Printf("Query failed: %v", err)
    return err
}
```

### 5. Close Resources

```go
// ✅ Good: Always defer Close
client, err := clickhouse.NewClient(ctx, cfg)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

## Error Handling

The package uses wrapped errors for better context:

```go
client, err := clickhouse.NewClient(ctx, cfg)
if err != nil {
    // Error: "failed to connect to ClickHouse: <original error>"
    log.Fatal(err)
}

err = client.Exec(ctx, "INVALID SQL")
if err != nil {
    // Contains detailed error from ClickHouse
    log.Printf("Exec error: %v", err)
}
```

## Performance Considerations

### Connection Pooling

Configure appropriate connection pool sizes:

```go
cfg := clickhouse.Config{
    MaxOpenConns:    20,  // Maximum concurrent connections
    MaxIdleConns:    10,  // Idle connections to maintain
    ConnMaxLifetime: 1 * time.Hour,
    ConnMaxIdleTime: 10 * time.Minute,
}
```

### Batch Size

For optimal performance, use appropriate batch sizes:

```go
// Good batch sizes: 1,000 - 10,000 rows
// Adjust based on your data and network
const batchSize = 5000

for i := 0; i < len(data); i += batchSize {
    end := i + batchSize
    if end > len(data) {
        end = len(data)
    }
    
    batch, _ := client.NewBatchInserter(ctx, "INSERT INTO table")
    for _, row := range data[i:end] {
        batch.Append(row...)
    }
    batch.Send()
}
```

### Use Compression

Compression can significantly reduce network traffic:

```go
cfg := clickhouse.Config{
    Compression: true, // Enabled by default
}
```

## Common Patterns

### Pagination

```go
func GetUsersPaginated(client *clickhouse.Client, page, pageSize int) ([]User, error) {
    offset := (page - 1) * pageSize
    
    qb := clickhouse.NewQueryBuilder()
    query, args := qb.
        Table("users").
        OrderBy("id", "ASC").
        Limit(pageSize).
        Offset(offset).
        BuildSelect()
    
    var users []User
    err := client.Query(context.Background(), &users, query, args...)
    return users, err
}
```

### Counting Total Records

```go
func GetUsersWithCount(client *clickhouse.Client, page, pageSize int) ([]User, int64, error) {
    ctx := context.Background()
    
    // Get count
    qb := clickhouse.NewQueryBuilder()
    countQuery, countArgs := qb.Table("users").BuildCount()
    
    var total uint64
    if err := client.QueryRow(ctx, &total, countQuery, countArgs...); err != nil {
        return nil, 0, err
    }
    
    // Get data
    offset := (page - 1) * pageSize
    qb.Reset()
    query, args := qb.
        Table("users").
        Limit(pageSize).
        Offset(offset).
        BuildSelect()
    
    var users []User
    err := client.Query(ctx, &users, query, args...)
    
    return users, int64(total), err
}
```

### Time Series Data

```go
func GetEventsByTimeRange(client *clickhouse.Client, start, end time.Time) ([]Event, error) {
    qb := clickhouse.NewQueryBuilder()
    query, args := qb.
        Table("events").
        WhereBetween("timestamp", start, end).
        OrderBy("timestamp", "DESC").
        BuildSelect()
    
    var events []Event
    err := client.Query(context.Background(), &events, query, args...)
    return events, err
}
```

## Testing

Run tests with:

```bash
# Test the package
go test ./clickhouse/...

# Test with coverage
go test -cover ./clickhouse/...

# Test with race detector
go test -race ./clickhouse/...
```

## Thread Safety

All client operations are thread-safe and can be used concurrently:

```go
client, _ := clickhouse.NewClient(ctx, cfg)

// Safe to use from multiple goroutines
for i := 0; i < 10; i++ {
    go func(id int) {
        client.Exec(ctx, "INSERT INTO logs VALUES (?, ?)", id, "message")
    }(i)
}
```

## License

This package is part of the vhvplatform/go-shared project.
