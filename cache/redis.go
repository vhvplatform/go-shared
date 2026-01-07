package cache

import (
	"context"
	"time"

	"github.com/vhvplatform/go-shared/redis"
)

// RedisCache wraps go-shared/redis.Cache to implement the cache.Cache interface
type RedisCache struct {
	cache *redis.Cache
}

// NewRedisCache creates a new Redis-based cache
func NewRedisCache(redisClient *redis.Client, config redis.CacheConfig) *RedisCache {
	return &RedisCache{
		cache: redis.NewCache(redisClient, config),
	}
}

// Get retrieves a value from Redis cache
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	if err := c.cache.Get(ctx, key, dest); err != nil {
		// Map Redis nil error to ErrCacheMiss
		if err.Error() == "redis: nil" {
			return ErrCacheMiss
		}
		return err
	}
	return nil
}

// Set stores a value in Redis cache
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.cache.Set(ctx, key, value, ttl)
}

// Delete removes a key from Redis cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.cache.Delete(ctx, key)
}

// Exists checks if a key exists in Redis cache
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.cache.Exists(ctx, key)
	return count > 0, err
}

// GetCache returns the underlying redis.Cache for advanced operations
func (c *RedisCache) GetCache() *redis.Cache {
	return c.cache
}
