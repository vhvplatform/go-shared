# MongoDB Package

A comprehensive MongoDB package for Go with advanced features including transaction support, pagination, query builder, and tenant isolation capabilities.

## Features

- **Transaction Support**: Execute operations within MongoDB transactions with proper session management
- **Pagination**: Both offset-based and cursor-based pagination with validation
- **Query Builder**: Fluent API for building complex MongoDB queries
- **Tenant Isolation**: Multi-tenancy support with automatic tenant filtering
- **Aggregation Builder**: Fluent interface for building aggregation pipelines
- **Connection Management**: Enhanced configuration with timeouts and connection pooling

## Installation

```bash
go get github.com/vhvcorp/go-shared/mongodb
```

## Basic Setup

### Creating a Client

```go
import (
    "context"
    "time"
    "github.com/vhvcorp/go-shared/mongodb"
)

func main() {
    cfg := mongodb.Config{
        URI:              "mongodb://localhost:27017",
        Database:         "myapp",
        MaxPoolSize:      100,
        MinPoolSize:      10,
        ConnectTimeout:   10 * time.Second,
        MaxConnIdleTime:  5 * time.Minute,
    }

    client, err := mongodb.NewClient(context.Background(), cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close(context.Background())

    // Get a collection
    collection := client.Collection("users")
}
```

### Backward Compatibility

The package maintains backward compatibility with existing code:

```go
// Old way still works
client, err := mongodb.NewMongoClient("mongodb://localhost:27017", "myapp")
```

## Query Builder

The query builder provides a fluent API for building MongoDB queries.

### Basic Operations

```go
import "github.com/vhvcorp/go-shared/mongodb"

// Simple equality
qb := mongodb.NewQueryBuilder()
filter := qb.Where("status", "active").Build()

// Multiple conditions
filter := qb.
    Where("status", "active").
    WhereGreaterThan("age", 18).
    WhereLessThanOrEqual("score", 100).
    Build()

// IN queries
filter := qb.WhereIn("status", []string{"active", "pending", "approved"}).Build()

// NOT IN queries
filter := qb.WhereNotIn("role", []string{"admin", "superuser"}).Build()
```

### Range Queries

```go
// Between
filter := qb.WhereBetween("age", 18, 65).Build()

// Greater than / Less than
filter := qb.
    WhereGreaterThanOrEqual("price", 100).
    WhereLessThan("price", 1000).
    Build()
```

### Pattern Matching

```go
// Regex search (case-insensitive)
filter := qb.WhereRegex("name", "^john", "i").Build()

// Text search
filter := qb.WhereTextSearch("mongodb query").Build()
```

### Field Existence and Null Checks

```go
// Check if field exists
filter := qb.WhereExists("email", true).Build()

// Null checks
filter := qb.WhereNull("deleted_at").Build()
filter := qb.WhereNotNull("confirmed_at").Build()
```

### Array Operations

```go
// Array contains
filter := qb.WhereArrayContains("tags", "mongodb").Build()

// Array size
filter := qb.WhereArraySize("items", 5).Build()
```

### Logical Operators

```go
// OR conditions
filter := qb.Or(
    bson.M{"status": "active"},
    bson.M{"status": "pending"},
).Build()

// AND conditions
filter := qb.And(
    bson.M{"status": "active"},
    bson.M{"verified": true},
).Build()
```

### Date Operations

```go
import "time"

// Exact date (day precision)
date := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
filter := qb.WhereDate("created_at", date).Build()

// Date ranges
filter := qb.WhereDateAfter("created_at", date).Build()
filter := qb.WhereDateBefore("expires_at", date).Build()
```

### ObjectID Handling

```go
// Query by ObjectID
filter := qb.WhereObjectID("_id", "507f1f77bcf86cd799439011").Build()
```

### Tenant Isolation

```go
// Build query with automatic tenant filtering
filter := qb.
    Where("status", "active").
    BuildWithTenant("tenant-123")
// Result: {status: "active", tenant_id: "tenant-123"}
```

### Query Builder Utilities

```go
// Clone for reusability
baseQuery := mongodb.NewQueryBuilder().Where("status", "active")
activeUsers := baseQuery.Clone().Where("type", "user").Build()
activePosts := baseQuery.Clone().Where("type", "post").Build()

// Reset to clear all conditions
qb.Reset()
```

## Pagination

### Offset-Based Pagination

Best for small to medium datasets with random access needs.

```go
import "github.com/vhvcorp/go-shared/mongodb"

// Create pagination params
params, err := mongodb.NewPaginationParams(1, 20) // page 1, 20 items per page
if err != nil {
    log.Fatal(err)
}

// Add sorting
params.WithSort(bson.D{{"created_at", -1}})

// Execute pagination
var users []User
result, err := mongodb.Paginate(ctx, collection, bson.M{"status": "active"}, params, &users)
if err != nil {
    log.Fatal(err)
}

// Access metadata
fmt.Printf("Page %d of %d\n", result.Page, result.TotalPages)
fmt.Printf("Total records: %d\n", result.Total)
fmt.Printf("Has next: %v\n", result.HasNext)
fmt.Printf("Has previous: %v\n", result.HasPrevious)
```

### Cursor-Based Pagination

Better performance for large datasets and real-time data.

```go
// Create cursor pagination
cp, err := mongodb.NewCursorPagination(20) // 20 items per request
if err != nil {
    log.Fatal(err)
}

// Optional: set cursor from previous request
cp.WithCursor("previous-cursor-token")

// Optional: set sort order
cp.WithSort(bson.D{{"created_at", -1}})

// Execute cursor pagination
var posts []Post
result, err := mongodb.PaginateWithCursor(ctx, collection, bson.M{}, cp, &posts)
if err != nil {
    log.Fatal(err)
}

// Use next cursor for subsequent requests
if result.HasNext {
    nextPage, _ := mongodb.NewCursorPagination(20)
    nextPage.WithCursor(result.NextCursor)
    // Fetch next page...
}
```

### Validation

Both pagination methods enforce limits:
- Page must be greater than 0
- Page size/limit must be between 1 and 100
- Maximum page size is 100 documents

## Transactions

Execute multiple operations atomically within a transaction.

### Basic Transaction

```go
import "go.mongodb.org/mongo-driver/mongo"

err := client.Transaction(ctx, func(sessCtx mongo.SessionContext) error {
    // All operations use sessCtx instead of ctx
    _, err := collection.InsertOne(sessCtx, user)
    if err != nil {
        return err // Transaction will rollback
    }

    _, err = collection.UpdateOne(sessCtx, filter, update)
    if err != nil {
        return err // Transaction will rollback
    }

    return nil // Transaction will commit
})

if err != nil {
    log.Fatal("Transaction failed:", err)
}
```

### Transaction with Options

```go
import "go.mongodb.org/mongo-driver/mongo/options"

opts := options.Transaction().
    SetReadPreference(readpref.Primary()).
    SetWriteConcern(writeconcern.Majority())

err := client.TransactionWithOptions(ctx, func(sessCtx mongo.SessionContext) error {
    // Your transactional operations
    return nil
}, opts)
```

## Tenant Isolation

Multi-tenancy support with automatic tenant filtering.

### TenantAware Interface

Implement this interface on your models for automatic tenant_id injection:

```go
type User struct {
    ID       primitive.ObjectID `bson:"_id,omitempty"`
    Name     string            `bson:"name"`
    Email    string            `bson:"email"`
    TenantID string            `bson:"tenant_id"`
}

func (u *User) GetTenantID() string {
    return u.TenantID
}

func (u *User) SetTenantID(tenantID string) {
    u.TenantID = tenantID
}
```

### TenantRepository

All operations are automatically scoped to the tenant:

```go
import "github.com/vhvcorp/go-shared/mongodb"

// Create tenant repository
collection := client.Collection("users")
tenantRepo := mongodb.NewTenantRepository(collection, "tenant-123")

// All operations are automatically filtered by tenant_id
var user User
err := tenantRepo.FindOne(ctx, bson.M{"email": "user@example.com"}, &user)

// Insert automatically adds tenant_id
user := &User{Name: "John", Email: "john@example.com"}
result, err := tenantRepo.InsertOne(ctx, user)
// user.TenantID is automatically set to "tenant-123"

// Update only affects documents with matching tenant_id
update := bson.M{"$set": bson.M{"status": "active"}}
result, err := tenantRepo.UpdateMany(ctx, bson.M{}, update)

// Delete only affects tenant's documents
result, err := tenantRepo.DeleteOne(ctx, bson.M{"_id": userID})

// Count tenant's documents
count, err := tenantRepo.CountDocuments(ctx, bson.M{"status": "active"})

// Tenant-aware pagination
params, _ := mongodb.NewPaginationParams(1, 20)
var users []User
result, err := tenantRepo.Paginate(ctx, bson.M{}, params, &users)
```

### Index Management

Create indexes for optimal tenant query performance:

```go
// Create index on tenant_id
err := tenantRepo.EnsureTenantIndex(ctx)

// Create compound index with tenant_id
err := tenantRepo.EnsureCompoundIndex(ctx, "email", "status")
// Creates index: {tenant_id: 1, email: 1, status: 1}
```

### Aggregation with Tenant Filtering

```go
pipeline := []bson.M{
    {"$group": bson.M{
        "_id": "$status",
        "count": bson.M{"$sum": 1},
    }},
}

var results []StatusCount
err := tenantRepo.Aggregate(ctx, pipeline, &results)
// Automatically prepends: {$match: {tenant_id: "tenant-123"}}
```

## Aggregation Builder

Build complex aggregation pipelines with a fluent API.

### Basic Pipeline

```go
ab := mongodb.NewAggregationBuilder()
pipeline := ab.
    Match(bson.M{"status": "active"}).
    Sort(bson.D{{"created_at", -1}}).
    Limit(10).
    Build()

var results []Result
cursor, err := collection.Aggregate(ctx, pipeline)
```

### Complex Aggregation

```go
ab := mongodb.NewAggregationBuilder()
pipeline := ab.
    Match(bson.M{"status": "active"}).
    Group(bson.M{
        "_id": "$category",
        "total": bson.M{"$sum": "$amount"},
        "count": bson.M{"$sum": 1},
        "avg": bson.M{"$avg": "$amount"},
    }).
    Sort(bson.D{{"total", -1}}).
    Limit(5).
    Build()
```

### Join Collections (Lookup)

```go
ab := mongodb.NewAggregationBuilder()
pipeline := ab.
    Match(bson.M{"status": "active"}).
    Lookup("users", "user_id", "_id", "user"). // Join with users collection
    Unwind("$user"). // Unwind the user array
    Project(bson.M{
        "name": 1,
        "email": 1,
        "user_name": "$user.name",
    }).
    Build()
```

### Advanced Operations

```go
ab := mongodb.NewAggregationBuilder()
pipeline := ab.
    Match(bson.M{"tenant_id": "tenant-123"}).
    Group(bson.M{
        "_id": bson.M{
            "year": bson.M{"$year": "$created_at"},
            "month": bson.M{"$month": "$created_at"},
        },
        "revenue": bson.M{"$sum": "$amount"},
    }).
    Sort(bson.D{{"_id.year", -1}, {"_id.month", -1}}).
    Skip(0).
    Limit(12).
    Build()
```

### Reusability

```go
// Create base pipeline
basePipeline := mongodb.NewAggregationBuilder().
    Match(bson.M{"status": "active"})

// Clone for different use cases
userStats := basePipeline.Clone().
    Group(bson.M{"_id": "$user_id", "count": bson.M{"$sum": 1}}).
    Build()

categoryStats := basePipeline.Clone().
    Group(bson.M{"_id": "$category", "total": bson.M{"$sum": "$amount"}}).
    Build()
```

## Populating Foreign Keys

The MongoDB package provides powerful helpers to populate (join) foreign key relationships in your queries, similar to SQL joins or Mongoose's `.populate()`.

### Simple Population

Use `PopulateField` to join a single foreign key:

```go
ab := mongodb.NewAggregationBuilder()
pipeline := ab.
    Match(bson.M{"status": "published"}).
    PopulateField("users", "author_id", "_id", "author", false).
    Build()

// This will:
// 1. Match documents with status "published"
// 2. Lookup user from "users" collection where user._id = document.author_id
// 3. Unwind the author array to a single object
```

### Population with Field Selection

Select only specific fields from the related collection:

```go
ab := mongodb.NewAggregationBuilder()
pipeline := ab.
    Match(bson.M{"status": "published"}).
    PopulateMultiple([]mongodb.PopulateConfig{
        {
            From:         "users",
            LocalField:   "author_id",
            ForeignField: "_id",
            As:           "author",
            Fields:       []string{"name", "email", "avatar"}, // Only include these fields
            PreserveNull: false,
        },
    }).
    Build()
```

### Multiple Foreign Keys

Populate multiple relationships in one query:

```go
ab := mongodb.NewAggregationBuilder()
pipeline := ab.
    Match(bson.M{"status": "published"}).
    PopulateMultiple([]mongodb.PopulateConfig{
        {
            From:         "users",
            LocalField:   "author_id",
            ForeignField: "_id",
            As:           "author",
            Fields:       []string{"name", "avatar"},
            PreserveNull: false,
        },
        {
            From:         "categories",
            LocalField:   "category_id",
            ForeignField: "_id",
            As:           "category",
            Fields:       []string{"name", "slug"},
            PreserveNull: true, // Keep posts even if category is missing
        },
        {
            From:         "tags",
            LocalField:   "tag_ids",
            ForeignField: "_id",
            As:           "tags",
            PreserveNull: true,
        },
    }).
    Build()
```

### Using PopulateHelper

For standalone population pipelines without the aggregation builder:

```go
ph := mongodb.NewPopulateHelper()

// Simple populate
pipeline := ph.PopulateSingle("users", "author_id", "_id", "author", false)

// Populate with field selection
pipeline = ph.PopulateWithFields(
    "users",
    "author_id",
    "_id",
    "author",
    []string{"name", "email", "avatar"},
    false,
)

// Populate with field renaming
pipeline = ph.PopulateWithRename(
    "users",
    "author_id",
    "_id",
    "author",
    map[string]string{
        "name":   "authorName",
        "email":  "authorEmail",
        "avatar": "authorAvatar",
    },
    false,
)

// Combine with your own pipeline
fullPipeline := []bson.M{
    {"$match": bson.M{"status": "published"}},
}
fullPipeline = append(fullPipeline, pipeline...)

cursor, _ := collection.Aggregate(ctx, fullPipeline)
```

### Complex Population with Field Renaming

```go
ph := mongodb.NewPopulateHelper()
pipeline := ph.BuildPopulatePipeline([]mongodb.LookupConfig{
    {
        From:         "users",
        LocalField:   "author_id",
        ForeignField: "_id",
        As:           "author",
        SelectFields: []string{"_id", "name", "email", "avatar", "bio"},
        RenameFields: map[string]string{
            "name":   "authorName",
            "email":  "authorEmail",
            "avatar": "authorAvatar",
            "bio":    "authorBio",
        },
        PreserveNull: false,
    },
    {
        From:         "organizations",
        LocalField:   "org_id",
        ForeignField: "_id",
        As:           "organization",
        SelectFields: []string{"name", "logo"},
        RenameFields: map[string]string{
            "name": "orgName",
            "logo": "orgLogo",
        },
        PreserveNull: true,
    },
})

// Apply to your collection
var results []bson.M
cursor, _ := collection.Aggregate(ctx, pipeline)
cursor.All(ctx, &results)
```

### Practical Example: Blog Posts with Author and Comments

```go
ab := mongodb.NewAggregationBuilder()
pipeline := ab.
    Match(bson.M{"published": true}).
    PopulateMultiple([]mongodb.PopulateConfig{
        {
            From:         "users",
            LocalField:   "author_id",
            ForeignField: "_id",
            As:           "author",
            Fields:       []string{"name", "avatar", "bio"},
            PreserveNull: false,
        },
        {
            From:         "categories",
            LocalField:   "category_id",
            ForeignField: "_id",
            As:           "category",
            Fields:       []string{"name", "slug"},
            PreserveNull: true,
        },
    }).
    // Add comment count using lookup without unwind
    LookupWithPipeline(
        "comments",
        "comments",
        bson.M{"postId": "$_id"},
        []bson.M{
            {"$match": bson.M{"$expr": bson.M{"$eq": []interface{}{"$post_id", "$$postId"}}}},
            {"$count": "count"},
        },
    ).
    AddFields(bson.M{
        "commentCount": bson.M{
            "$ifNull": []interface{}{
                bson.M{"$arrayElemAt": []interface{}{"$comments.count", 0}},
                0,
            },
        },
    }).
    Project(bson.M{
        "title":        1,
        "content":      1,
        "author":       1,
        "category":     1,
        "commentCount": 1,
        "created_at":   1,
    }).
    Sort(bson.D{{Key: "created_at", Value: -1}}).
    Limit(20).
    Build()

var posts []BlogPost
cursor, _ := postsCollection.Aggregate(ctx, pipeline)
cursor.All(ctx, &posts)
```

## Advanced Aggregation Features

### Facet Operations

Facets allow you to run multiple aggregation pipelines on the same set of input documents:

```go
ab := mongodb.NewAggregationBuilder()
pipeline := ab.
    Match(bson.M{"status": "active"}).
    Facet(map[string][]bson.M{
        "categorizedByStatus": {
            {"$group": bson.M{"_id": "$status", "count": bson.M{"$sum": 1}}},
            {"$sort": bson.D{{Key: "count", Value: -1}}},
        },
        "categorizedByType": {
            {"$group": bson.M{"_id": "$type", "count": bson.M{"$sum": 1}}},
            {"$sort": bson.D{{Key: "count", Value: -1}}},
        },
        "priceStats": {
            {"$group": bson.M{
                "_id": nil,
                "avgPrice": bson.M{"$avg": "$price"},
                "minPrice": bson.M{"$min": "$price"},
                "maxPrice": bson.M{"$max": "$price"},
            }},
        },
    }).
    Build()

var result struct {
    CategorizedByStatus []struct {
        ID    string `bson:"_id"`
        Count int    `bson:"count"`
    } `bson:"categorizedByStatus"`
    CategorizedByType []struct {
        ID    string `bson:"_id"`
        Count int    `bson:"count"`
    } `bson:"categorizedByType"`
    PriceStats []struct {
        AvgPrice float64 `bson:"avgPrice"`
        MinPrice float64 `bson:"minPrice"`
        MaxPrice float64 `bson:"maxPrice"`
    } `bson:"priceStats"`
}
cursor, _ := collection.Aggregate(ctx, pipeline)
cursor.Decode(&result)
```

### Bucketing Data

Group documents into buckets based on field values:

```go
// Manual bucket boundaries
ab := mongodb.NewAggregationBuilder()
pipeline := ab.
    Bucket(
        "$price",                           // Field to bucket by
        []interface{}{0, 50, 100, 200, 500}, // Bucket boundaries
        "expensive",                         // Default bucket name for values outside boundaries
        bson.M{"count": bson.M{"$sum": 1}, "avgPrice": bson.M{"$avg": "$price"}},
    ).
    Build()

// Automatic bucketing
pipeline = ab.
    BucketAuto(
        "$age",    // Field to bucket by
        5,         // Number of buckets
        bson.M{"count": bson.M{"$sum": 1}},
        "R5",      // Granularity (R5, R10, R20, R40, R80, 1-2-5, E6, E12, E24, E48, E96, E192, POWERSOF2)
    ).
    Build()
```

### Additional Aggregation Utilities

```go
ab := mongodb.NewAggregationBuilder()

// Add computed fields
ab.AddFields(bson.M{
    "fullName": bson.M{"$concat": []string{"$firstName", " ", "$lastName"}},
    "totalPrice": bson.M{"$multiply": []interface{}{"$price", "$quantity"}},
})

// Replace document root
ab.ReplaceRoot("$embeddedDoc")

// Random sampling
ab.Sample(100) // Get 100 random documents

// Count documents
ab.Count("totalDocuments")

// Group and sort by count
ab.SortByCount("$category") // Groups by category and sorts by count

// Unwind with options (preserve null/empty arrays)
ab.UnwindWithOptions("$items", true)
```

## Quick Statistics Helpers

The `StatsHelper` provides pre-built pipelines for common statistical operations:

### Count by Field

```go
sh := mongodb.NewStatsHelper()

// Count documents grouped by a field
pipeline := sh.CountByField("category")
// Result: [{_id: "electronics", count: 150}, {_id: "books", count: 89}, ...]

var results []struct {
    Category string `bson:"_id"`
    Count    int    `bson:"count"`
}
cursor, _ := collection.Aggregate(ctx, pipeline)
cursor.All(ctx, &results)
```

### Sum and Average

```go
sh := mongodb.NewStatsHelper()

// Sum a field grouped by another field
pipeline := sh.SumByField("category", "amount")
// Result: [{_id: "electronics", total: 15000}, {_id: "books", total: 8900}, ...]

// Average a field grouped by another field
pipeline = sh.AverageByField("category", "price")
// Result: [{_id: "electronics", average: 499.99, count: 30}, ...]
```

### Min/Max Values

```go
sh := mongodb.NewStatsHelper()

// Get min and max values
pipeline := sh.MinMaxByField("category", "price")
// Result: [{_id: "electronics", min: 9.99, max: 1999.99}, ...]
```

### Comprehensive Field Statistics

```go
sh := mongodb.NewStatsHelper()

// Get all statistics for a numeric field
pipeline := sh.StatsForField("price")
// Result: {_id: null, count: 500, sum: 250000, avg: 500, min: 9.99, max: 1999.99}

var stats struct {
    Count int     `bson:"count"`
    Sum   float64 `bson:"sum"`
    Avg   float64 `bson:"avg"`
    Min   float64 `bson:"min"`
    Max   float64 `bson:"max"`
}
cursor, _ := collection.Aggregate(ctx, pipeline)
cursor.Decode(&stats)
```

### Top/Bottom N Documents

```go
sh := mongodb.NewStatsHelper()

// Get top 10 documents by score
pipeline := sh.TopN("score", 10)

// Get bottom 5 documents by rating
pipeline = sh.BottomN("rating", 5)
```

### Date Range Statistics

```go
sh := mongodb.NewStatsHelper()

// Group by year
pipeline := sh.DateRangeStats("createdAt", "year")
// Result: [{_id: 2023, count: 450}, {_id: 2024, count: 892}, ...]

// Group by month
pipeline = sh.DateRangeStats("createdAt", "month")
// Result: [{_id: {year: 2024, month: 1}, count: 75}, ...]

// Group by day
pipeline = sh.DateRangeStats("createdAt", "day")
// Result: [{_id: {year: 2024, month: 12, day: 23}, count: 12}, ...]

var monthlyStats []struct {
    Date  struct {
        Year  int `bson:"year"`
        Month int `bson:"month"`
    } `bson:"_id"`
    Count int `bson:"count"`
}
cursor, _ := collection.Aggregate(ctx, pipeline)
cursor.All(ctx, &monthlyStats)
```

### Percentile Statistics

```go
sh := mongodb.NewStatsHelper()

// Calculate specific percentiles (25th, 50th, 75th, 90th)
pipeline := sh.PercentileStats("price", []float64{25, 50, 75, 90})

var result struct {
    Count       int `bson:"count"`
    Percentiles []struct {
        Percentile float64 `bson:"percentile"`
        Value      float64 `bson:"value"`
    } `bson:"percentiles"`
}
cursor, _ := collection.Aggregate(ctx, pipeline)
cursor.Decode(&result)
// Result: {count: 500, percentiles: [
//   {percentile: 25, value: 100}, 
//   {percentile: 50, value: 250},
//   {percentile: 75, value: 500},
//   {percentile: 90, value: 800}
// ]}
```

### Combining Statistics with Filters

```go
sh := mongodb.NewStatsHelper()
ab := mongodb.NewAggregationBuilder()

// Get statistics for active users only
pipeline := ab.
    Match(bson.M{"status": "active"}).
    Build()

// Append stats pipeline
statsForActive := sh.StatsForField("score")
pipeline = append(pipeline, statsForActive...)

// Or use facets for multiple filtered statistics
pipeline = ab.
    Facet(map[string][]bson.M{
        "activeStats": append(
            []bson.M{{"$match": bson.M{"status": "active"}}},
            sh.StatsForField("score")...,
        ),
        "inactiveStats": append(
            []bson.M{{"$match": bson.M{"status": "inactive"}}},
            sh.StatsForField("score")...,
        ),
    }).
    Build()
```

## Base Repository Pattern

The base repository pattern provides generic CRUD operations with built-in support for soft delete, transaction handling, query builders, and index management.

### Creating a Repository

```go
import "github.com/vhvcorp/go-shared/mongodb"

// Create a repository without soft delete
userRepo := mongodb.NewBaseRepository(mongodb.RepositoryConfig{
    Collection: client.Collection("users"),
    Client:     client, // Required for transaction support
    SoftDelete: false,
})

// Create a repository with soft delete enabled
productRepo := mongodb.NewBaseRepository(mongodb.RepositoryConfig{
    Collection: client.Collection("products"),
    Client:     client,
    SoftDelete: true, // Enable soft delete
})
```

### Base Model

Use `BaseModel` in your structs to get automatic timestamp management:

```go
type User struct {
    mongodb.BaseModel // Includes ID, CreatedAt, UpdatedAt, DeletedAt
    Name     string `bson:"name" json:"name"`
    Email    string `bson:"email" json:"email"`
    Password string `bson:"password" json:"-"`
}
```

### CRUD Operations

**Create:**
```go
// Create single document (auto-adds CreatedAt and UpdatedAt)
user := &User{
    Name:  "John Doe",
    Email: "john@example.com",
}
result, err := userRepo.Create(ctx, user)

// Create multiple documents
users := []interface{}{user1, user2, user3}
result, err := userRepo.CreateMany(ctx, users)
```

**Read:**
```go
// Find by ID
var user User
err := userRepo.FindByID(ctx, userID, &user)

// Find one matching filter
err := userRepo.FindOne(ctx, bson.M{"email": "john@example.com"}, &user)

// Find multiple
var users []User
err := userRepo.Find(ctx, bson.M{"status": "active"}, &users)

// Find all
err := userRepo.FindAll(ctx, &users)

// Check if exists
exists, err := userRepo.Exists(ctx, bson.M{"email": "john@example.com"})

// Count documents
count, err := userRepo.Count(ctx, bson.M{"status": "active"})
```

**Update:**
```go
// Update by ID (auto-updates UpdatedAt)
update := bson.M{"$set": bson.M{"name": "Jane Doe"}}
result, err := userRepo.UpdateByID(ctx, userID, update)

// Update single matching filter
result, err := userRepo.Update(ctx, bson.M{"email": "old@example.com"}, update)

// Update multiple
result, err := userRepo.UpdateMany(ctx, bson.M{"status": "pending"}, update)
```

**Delete:**
```go
// Delete by ID (soft delete if enabled)
result, err := userRepo.DeleteByID(ctx, userID)

// Delete matching filter
result, err := userRepo.Delete(ctx, bson.M{"status": "inactive"})

// Delete multiple
result, err := userRepo.DeleteMany(ctx, bson.M{"created_at": bson.M{"$lt": oldDate}})
```

### Soft Delete Support

When soft delete is enabled, documents are marked as deleted instead of being removed:

```go
// Soft delete (sets deleted_at timestamp)
result, err := productRepo.DeleteByID(ctx, productID)

// Hard delete (permanently removes document)
result, err := productRepo.HardDeleteByID(ctx, productID)

// Restore soft-deleted document
result, err := productRepo.RestoreByID(ctx, productID)

// Restore with filter
result, err := productRepo.Restore(ctx, bson.M{"sku": "PROD-123"})
```

Soft-deleted documents are automatically excluded from all query operations.

### Transaction Support

Execute multiple operations atomically:

```go
err := userRepo.Transaction(ctx, func(sessCtx mongo.SessionContext) error {
    // Create user
    _, err := userRepo.Create(sessCtx, newUser)
    if err != nil {
        return err // Rolls back transaction
    }
    
    // Update related document
    _, err = accountRepo.UpdateByID(sessCtx, accountID, update)
    if err != nil {
        return err // Rolls back transaction
    }
    
    return nil // Commits transaction
})

// With custom options
opts := options.Transaction().
    SetReadConcern(readconcern.Majority()).
    SetWriteConcern(writeconcern.Majority())

err := userRepo.TransactionWithOptions(ctx, func(sessCtx mongo.SessionContext) error {
    // Your transactional operations
    return nil
}, opts)
```

### Query Builder Integration

Use the built-in query builder for complex queries:

```go
qb := userRepo.GetQueryBuilder()

filter := qb.
    Where("status", "active").
    WhereGreaterThan("age", 18).
    WhereBetween("score", 50, 100).
    Build()

var users []User
err := userRepo.Find(ctx, filter, &users)
```

### Pagination

```go
params, _ := mongodb.NewPaginationParams(1, 20)
params.WithSort(bson.D{{Key: "created_at", Value: -1}})

var users []User
result, err := userRepo.Paginate(ctx, bson.M{"status": "active"}, params, &users)

fmt.Printf("Page %d of %d, Total: %d\n", 
    result.Page, result.TotalPages, result.Total)
```

### Aggregation with Soft Delete

Aggregation automatically excludes soft-deleted documents:

```go
pipeline := []bson.M{
    {"$group": bson.M{
        "_id": "$category",
        "count": bson.M{"$sum": 1},
    }},
    {"$sort": bson.D{{Key: "count", Value: -1}}},
}

var results []CategoryCount
err := productRepo.Aggregate(ctx, pipeline, &results)
// Automatically prepends: {$match: {deleted_at: {$exists: false}}}
```

### Index Management

**Create Single Index:**
```go
// Simple index
indexName, err := userRepo.CreateIndex(ctx, bson.D{{Key: "email", Value: 1}})

// Unique index
indexName, err := userRepo.CreateUniqueIndex(ctx, 
    bson.D{{Key: "email", Value: 1}}, 
    "email_unique",
)

// Text index for full-text search
indexName, err := userRepo.CreateTextIndex(ctx, "title", "description")

// TTL index for automatic expiration
indexName, err := userRepo.CreateTTLIndex(ctx, "expires_at", 3600) // 1 hour

// Compound index
indexName, err := userRepo.CreateCompoundIndex(ctx, 
    map[string]int{
        "email":  1,
        "status": 1,
    }, 
    true, // unique
)
```

**Create Multiple Indexes:**
```go
indexes := []mongodb.IndexConfig{
    {
        Keys: bson.D{{Key: "email", Value: 1}},
        Options: options.Index().SetUnique(true),
    },
    {
        Keys: bson.D{{Key: "created_at", Value: -1}},
    },
    {
        Keys: bson.D{
            {Key: "status", Value: 1},
            {Key: "priority", Value: -1},
        },
    },
}

indexNames, err := userRepo.CreateIndexes(ctx, indexes)
```

**Ensure Indexes:**
```go
// Create indexes if they don't exist
err := userRepo.EnsureIndexes(ctx, indexes)
```

**Manage Indexes:**
```go
// List all indexes
indexes, err := userRepo.ListIndexes(ctx)

// Drop specific index
err := userRepo.DropIndex(ctx, "email_unique")

// Drop all indexes (except _id)
err := userRepo.DropAllIndexes(ctx)
```

### Complete Example: User Repository

```go
package repository

import (
    "context"
    "github.com/vhvcorp/go-shared/mongodb"
    "go.mongodb.org/mongo-driver/bson"
)

type User struct {
    mongodb.BaseModel
    Email    string `bson:"email" json:"email"`
    Name     string `bson:"name" json:"name"`
    Role     string `bson:"role" json:"role"`
    Status   string `bson:"status" json:"status"`
}

type UserRepository struct {
    *mongodb.BaseRepository
}

func NewUserRepository(client *mongodb.Client) *UserRepository {
    collection := client.Collection("users")
    
    repo := &UserRepository{
        BaseRepository: mongodb.NewBaseRepository(mongodb.RepositoryConfig{
            Collection: collection,
            Client:     client,
            SoftDelete: true, // Enable soft delete for users
        }),
    }
    
    // Ensure indexes
    ctx := context.Background()
    repo.ensureIndexes(ctx)
    
    return repo
}

func (r *UserRepository) ensureIndexes(ctx context.Context) error {
    indexes := []mongodb.IndexConfig{
        {
            Keys: bson.D{{Key: "email", Value: 1}},
            Options: options.Index().SetUnique(true),
        },
        {
            Keys: bson.D{
                {Key: "status", Value: 1},
                {Key: "role", Value: 1},
            },
        },
    }
    return r.EnsureIndexes(ctx, indexes)
}

// Custom methods
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
    var user User
    err := r.FindOne(ctx, bson.M{"email": email}, &user)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) FindActiveUsers(ctx context.Context) ([]User, error) {
    var users []User
    err := r.Find(ctx, bson.M{"status": "active"}, &users)
    return users, err
}

func (r *UserRepository) DeactivateUser(ctx context.Context, userID primitive.ObjectID) error {
    update := bson.M{"$set": bson.M{"status": "inactive"}}
    _, err := r.UpdateByID(ctx, userID, update)
    return err
}
```

**Usage:**
```go
// Initialize
userRepo := repository.NewUserRepository(mongoClient)

// Create user
user := &User{
    Email:  "user@example.com",
    Name:   "John Doe",
    Role:   "user",
    Status: "active",
}
result, err := userRepo.Create(ctx, user)

// Find by email (custom method)
foundUser, err := userRepo.FindByEmail(ctx, "user@example.com")

// Soft delete
_, err = userRepo.DeleteByID(ctx, user.ID)

// User is now hidden from queries
exists, _ := userRepo.Exists(ctx, bson.M{"email": "user@example.com"}) // false

// Restore
_, err = userRepo.RestoreByID(ctx, user.ID)

// Now visible again
exists, _ = userRepo.Exists(ctx, bson.M{"email": "user@example.com"}) // true
```

## Best Practices

### Context Usage

Always use context with timeouts for database operations:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result, err := collection.FindOne(ctx, filter).Decode(&doc)
```

### Index Creation for Tenant ID

Always create indexes on tenant_id for multi-tenant applications:

```go
tenantRepo.EnsureTenantIndex(ctx)
tenantRepo.EnsureCompoundIndex(ctx, "email") // {tenant_id: 1, email: 1}
```

### Transaction Best Practices

- Keep transactions short and focused
- Avoid long-running operations within transactions
- Always handle transaction errors appropriately
- Use appropriate read/write concerns

```go
err := client.Transaction(ctx, func(sessCtx mongo.SessionContext) error {
    // Keep operations minimal and fast
    // Avoid external API calls
    // Return errors to rollback
    return nil
})
```

### Cursor vs Offset Pagination

**Use Offset Pagination when:**
- Dataset is small to medium (< 10,000 records)
- Random page access is needed (jump to page 5)
- Total count and page numbers are required
- Data doesn't change frequently

**Use Cursor Pagination when:**
- Dataset is large (> 10,000 records)
- Sequential access pattern (next/previous)
- Better performance is critical
- Real-time data with frequent updates

### Query Builder Cloning

Clone query builders when you need to reuse base queries:

```go
baseQuery := mongodb.NewQueryBuilder().
    Where("status", "active").
    Where("tenant_id", tenantID)

// Reuse for different queries
users := baseQuery.Clone().Where("type", "user").Build()
admins := baseQuery.Clone().Where("role", "admin").Build()
```

### Connection Pooling

Configure appropriate pool sizes based on your workload:

```go
cfg := mongodb.Config{
    MaxPoolSize: 100, // Maximum connections
    MinPoolSize: 10,  // Minimum connections to maintain
    ConnectTimeout: 10 * time.Second,
    MaxConnIdleTime: 5 * time.Minute,
}
```

## Error Handling

The package uses wrapped errors for better context:

```go
_, err := mongodb.NewPaginationParams(0, 10)
if err != nil {
    // Error: "page must be greater than 0"
}

err := client.Transaction(ctx, fn)
if err != nil {
    // Error: "transaction failed: <original error>"
}
```

## Testing

The package includes comprehensive tests. Run them with:

```bash
make test-pkg
```

## Thread Safety

All components are thread-safe and can be used concurrently:
- QueryBuilder and AggregationBuilder are safe to clone and use across goroutines
- TenantRepository is safe for concurrent use
- Client connection pool is managed by the MongoDB driver

## License

This package is part of the saas-framework-go project.
