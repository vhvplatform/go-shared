package mongodb

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PaginationParams holds parameters for offset-based pagination
type PaginationParams struct {
	Page     int64
	PageSize int64
	Sort     bson.D
}

// PaginationResult contains paginated results with metadata
type PaginationResult struct {
	Data        interface{} `json:"data"`
	Total       int64       `json:"total"`
	Page        int64       `json:"page"`
	PageSize    int64       `json:"page_size"`
	TotalPages  int64       `json:"total_pages"`
	HasNext     bool        `json:"has_next"`
	HasPrevious bool        `json:"has_previous"`
}

// CursorPagination holds parameters for cursor-based pagination
type CursorPagination struct {
	Limit  int64
	Cursor string
	Sort   bson.D
}

// CursorResult contains cursor-based pagination results
type CursorResult struct {
	Data       interface{} `json:"data"`
	NextCursor string      `json:"next_cursor,omitempty"`
	PrevCursor string      `json:"prev_cursor,omitempty"`
	HasNext    bool        `json:"has_next"`
	HasPrev    bool        `json:"has_prev"`
}

const (
	// MaxPageSize is the maximum allowed page size
	MaxPageSize = 100
)

// NewPaginationParams creates a new PaginationParams with validation
func NewPaginationParams(page, pageSize int64) (*PaginationParams, error) {
	if page < 1 {
		return nil, errors.New("page must be greater than 0")
	}
	if pageSize < 1 {
		return nil, errors.New("page_size must be greater than 0")
	}
	if pageSize > MaxPageSize {
		return nil, fmt.Errorf("page_size cannot exceed %d", MaxPageSize)
	}

	return &PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Sort:     bson.D{},
	}, nil
}

// Skip calculates the number of documents to skip
func (p *PaginationParams) Skip() int64 {
	return (p.Page - 1) * p.PageSize
}

// WithSort sets the sort order for pagination
func (p *PaginationParams) WithSort(sort bson.D) *PaginationParams {
	p.Sort = sort
	return p
}

// Paginate performs offset-based pagination on a collection
// Performance: Run count and find operations concurrently for better performance
func Paginate(ctx context.Context, collection *mongo.Collection, filter bson.M, params *PaginationParams, results interface{}) (*PaginationResult, error) {
	if params == nil {
		return nil, errors.New("pagination params cannot be nil")
	}

	// Set find options
	findOptions := options.Find().
		SetSkip(params.Skip()).
		SetLimit(params.PageSize)

	if len(params.Sort) > 0 {
		findOptions.SetSort(params.Sort)
	}

	// Run count and find operations concurrently for better performance
	type countResult struct {
		total int64
		err   error
	}
	type findResult struct {
		cursor *mongo.Cursor
		err    error
	}

	countChan := make(chan countResult, 1)
	findChan := make(chan findResult, 1)

	// Execute count in goroutine
	go func() {
		total, err := collection.CountDocuments(ctx, filter)
		countChan <- countResult{total: total, err: err}
	}()

	// Execute find in goroutine
	go func() {
		cursor, err := collection.Find(ctx, filter, findOptions)
		findChan <- findResult{cursor: cursor, err: err}
	}()

	// Wait for both operations
	countRes := <-countChan
	findRes := <-findChan

	if countRes.err != nil {
		return nil, fmt.Errorf("failed to count documents: %w", countRes.err)
	}
	if findRes.err != nil {
		return nil, fmt.Errorf("failed to execute find query: %w", findRes.err)
	}

	defer findRes.cursor.Close(ctx)

	// Decode results
	if err := findRes.cursor.All(ctx, results); err != nil {
		return nil, fmt.Errorf("failed to decode results: %w", err)
	}

	// Calculate total pages
	totalPages := (countRes.total + params.PageSize - 1) / params.PageSize

	return &PaginationResult{
		Data:        results,
		Total:       countRes.total,
		Page:        params.Page,
		PageSize:    params.PageSize,
		TotalPages:  totalPages,
		HasNext:     params.Page < totalPages,
		HasPrevious: params.Page > 1,
	}, nil
}

// NewCursorPagination creates a new CursorPagination with validation
func NewCursorPagination(limit int64) (*CursorPagination, error) {
	if limit < 1 {
		return nil, errors.New("limit must be greater than 0")
	}
	if limit > MaxPageSize {
		return nil, fmt.Errorf("limit cannot exceed %d", MaxPageSize)
	}

	return &CursorPagination{
		Limit:  limit,
		Cursor: "",
		Sort:   bson.D{},
	}, nil
}

// WithCursor sets the cursor for pagination
func (cp *CursorPagination) WithCursor(cursor string) *CursorPagination {
	cp.Cursor = cursor
	return cp
}

// WithSort sets the sort order for cursor pagination
func (cp *CursorPagination) WithSort(sort bson.D) *CursorPagination {
	cp.Sort = sort
	return cp
}

// PaginateWithCursor performs cursor-based pagination on a collection
// Note: This implementation uses _id field as the cursor. For ObjectID types,
// the cursor will be the hex string representation.
func PaginateWithCursor(ctx context.Context, collection *mongo.Collection, filter bson.M, params *CursorPagination, results interface{}) (*CursorResult, error) {
	if params == nil {
		return nil, errors.New("cursor pagination params cannot be nil")
	}

	// Build filter with cursor if provided
	queryFilter := filter
	if params.Cursor != "" {
		// Try to parse as ObjectID first, fall back to string comparison
		var cursorValue interface{} = params.Cursor
		if objectID, err := primitive.ObjectIDFromHex(params.Cursor); err == nil {
			cursorValue = objectID
		}

		cursorFilter := bson.M{"_id": bson.M{"$gt": cursorValue}}
		if filter != nil {
			queryFilter = bson.M{
				"$and": []bson.M{
					filter,
					cursorFilter,
				},
			}
		} else {
			queryFilter = cursorFilter
		}
	}

	// Fetch one more than limit to check if there are more results
	findOptions := options.Find().
		SetLimit(params.Limit + 1)

	if len(params.Sort) > 0 {
		findOptions.SetSort(params.Sort)
	} else {
		// Default sort by _id for cursor pagination
		findOptions.SetSort(bson.D{{Key: "_id", Value: 1}})
	}

	// Execute query
	cursor, err := collection.Find(ctx, queryFilter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to execute find query: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode results into a temporary slice
	var tempResults []bson.M
	if err := cursor.All(ctx, &tempResults); err != nil {
		return nil, fmt.Errorf("failed to decode results: %w", err)
	}

	// Check if there are more results
	hasNext := len(tempResults) > int(params.Limit)
	if hasNext {
		tempResults = tempResults[:params.Limit]
	}

	// Get next cursor from last result - properly encode based on type
	var nextCursor string
	if len(tempResults) > 0 && hasNext {
		lastDoc := tempResults[len(tempResults)-1]
		if id, ok := lastDoc["_id"]; ok {
			// Handle ObjectID type specially to get consistent hex representation
			if objectID, ok := id.(primitive.ObjectID); ok {
				nextCursor = objectID.Hex()
			} else {
				nextCursor = fmt.Sprintf("%v", id)
			}
		}
	}

	return &CursorResult{
		Data:       tempResults,
		NextCursor: nextCursor,
		HasNext:    hasNext,
		HasPrev:    params.Cursor != "",
	}, nil
}

// PaginateFast performs fast pagination without counting total documents
// Performance: Skip counting step for better performance when total count isn't needed
// Use this when you only need to know if there are more pages (HasNext)
func PaginateFast(ctx context.Context, collection *mongo.Collection, filter bson.M, params *PaginationParams, results interface{}) (*PaginationResult, error) {
	if params == nil {
		return nil, errors.New("pagination params cannot be nil")
	}

	// Fetch one more than pageSize to determine if there are more results
	findOptions := options.Find().
		SetSkip(params.Skip()).
		SetLimit(params.PageSize + 1)

	if len(params.Sort) > 0 {
		findOptions.SetSort(params.Sort)
	}

	// Execute query
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to execute find query: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode into temporary slice to check for extra result
	var tempResults []bson.M
	if err := cursor.All(ctx, &tempResults); err != nil {
		return nil, fmt.Errorf("failed to decode results: %w", err)
	}

	// Check if there are more results
	hasNext := len(tempResults) > int(params.PageSize)
	if hasNext {
		tempResults = tempResults[:params.PageSize]
	}

	return &PaginationResult{
		Data:        tempResults,
		Total:       -1, // Total is unknown in fast mode
		Page:        params.Page,
		PageSize:    params.PageSize,
		TotalPages:  -1, // Total pages is unknown in fast mode
		HasNext:     hasNext,
		HasPrevious: params.Page > 1,
	}, nil
}
