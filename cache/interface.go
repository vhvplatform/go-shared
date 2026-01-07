package cache

import (
	"context"
	"time"
)

// Cache is a generic cache interface that can be implemented by different backends
type Cache interface {
	// Get retrieves a value from the cache and deserializes it into dest
	Get(ctx context.Context, key string, dest interface{}) error

	// Set stores a value in the cache with the specified TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a key from the cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in the cache
	Exists(ctx context.Context, key string) (bool, error)
}
