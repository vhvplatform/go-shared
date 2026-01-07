package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dgraph-io/ristretto"
)

// RistrettoCache implements Cache interface using Ristretto in-memory cache
type RistrettoCache struct {
	client *ristretto.Cache[string, any]
}

// RistrettoConfig holds configuration for Ristretto cache
type RistrettoConfig struct {
	MaxCost     int64 // Maximum cost of cache (approx memory in bytes)
	NumCounters int64 // Should be ~10x the number of keys
	BufferItems int64 // Number of keys per Get buffer (default: 64)
}

// NewRistrettoCache creates a new Ristretto-based cache
func NewRistrettoCache(config RistrettoConfig) (*RistrettoCache, error) {
	if config.BufferItems == 0 {
		config.BufferItems = 64
	}

	cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: config.NumCounters,
		MaxCost:     config.MaxCost,
		BufferItems: config.BufferItems,
	})
	if err != nil {
		return nil, err
	}

	return &RistrettoCache{client: cache}, nil
}

// Get retrieves a value from the cache
func (c *RistrettoCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, found := c.client.Get(key)
	if !found {
		return ErrCacheMiss
	}

	data, ok := val.([]byte)
	if !ok {
		return ErrInvalidCacheValue
	}

	return json.Unmarshal(data, dest)
}

// Set stores a value in the cache
func (c *RistrettoCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	cost := int64(len(data))
	c.client.SetWithTTL(key, data, cost, ttl)
	return nil
}

// Delete removes a key from the cache
func (c *RistrettoCache) Delete(ctx context.Context, key string) error {
	c.client.Del(key)
	return nil
}

// Exists checks if a key exists in the cache
func (c *RistrettoCache) Exists(ctx context.Context, key string) (bool, error) {
	_, found := c.client.Get(key)
	return found, nil
}

// Close closes the cache and releases resources
func (c *RistrettoCache) Close() error {
	c.client.Close()
	return nil
}
