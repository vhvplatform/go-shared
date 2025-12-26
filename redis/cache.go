package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheConfig holds configuration for the cache client
type CacheConfig struct {
	DefaultTTL time.Duration // Default expiration time for cache entries
	KeyPrefix  string        // Namespace prefix for all keys
	Serializer Serializer    // Serializer for encoding/decoding values
}

// Cache wraps Redis client with enhanced caching features
type Cache struct {
	client     *redis.Client
	config     CacheConfig
	serializer Serializer
}

// NewCache creates a new cache instance with the given configuration
func NewCache(client *Client, config CacheConfig) *Cache {
	// Set defaults
	if config.DefaultTTL == 0 {
		config.DefaultTTL = 1 * time.Hour
	}
	if config.Serializer == nil {
		config.Serializer = NewJSONSerializer()
	}

	return &Cache{
		client:     client.GetClient(),
		config:     config,
		serializer: config.Serializer,
	}
}

// buildKey adds the configured prefix to the key
func (c *Cache) buildKey(key string) string {
	if c.config.KeyPrefix == "" {
		return key
	}
	return c.config.KeyPrefix + ":" + key
}

// stripPrefix removes the configured prefix from the key
func (c *Cache) stripPrefix(key string) string {
	if c.config.KeyPrefix == "" {
		return key
	}
	prefix := c.config.KeyPrefix + ":"
	return strings.TrimPrefix(key, prefix)
}

// WithPrefix creates a new cache instance with an additional prefix
func (c *Cache) WithPrefix(prefix string) *Cache {
	newConfig := c.config
	if c.config.KeyPrefix == "" {
		newConfig.KeyPrefix = prefix
	} else {
		newConfig.KeyPrefix = c.config.KeyPrefix + ":" + prefix
	}

	return &Cache{
		client:     c.client,
		config:     newConfig,
		serializer: c.serializer,
	}
}

// Set stores a value in the cache with the specified TTL
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := c.serializer.Serialize(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %w", err)
	}

	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}

	return c.client.Set(ctx, c.buildKey(key), data, ttl).Err()
}

// Get retrieves a value from the cache
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, c.buildKey(key)).Bytes()
	if err != nil {
		return err
	}

	return c.serializer.Deserialize(data, dest)
}

// Delete removes a key from the cache
func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.buildKey(key)).Err()
}

// MGet retrieves multiple keys at once
func (c *Cache) MGet(ctx context.Context, keys ...string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	// Build prefixed keys
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.buildKey(key)
	}

	// Use pipeline for efficiency
	pipe := c.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(prefixedKeys))
	for i, prefixedKey := range prefixedKeys {
		cmds[i] = pipe.Get(ctx, prefixedKey)
	}

	_, _ = pipe.Exec(ctx) // Ignore pipeline errors, check individual commands

	// Collect results
	result := make(map[string]interface{})
	var lastErr error
	for i, cmd := range cmds {
		val, err := cmd.Result()
		if err == redis.Nil {
			// Key doesn't exist, skip
			continue
		}
		if err != nil {
			// Store error but continue processing other keys
			lastErr = fmt.Errorf("failed to get key %s: %w", keys[i], err)
			continue
		}

		// Store as string directly since we can't know the original type
		result[keys[i]] = val
	}

	return result, lastErr
}

// MSet sets multiple key-value pairs at once
func (c *Cache) MSet(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}

	// Use pipeline for efficiency
	pipe := c.client.Pipeline()

	for key, value := range items {
		data, err := c.serializer.Serialize(value)
		if err != nil {
			return fmt.Errorf("failed to serialize value for key %s: %w", key, err)
		}
		pipe.Set(ctx, c.buildKey(key), data, ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// MDelete deletes multiple keys at once
func (c *Cache) MDelete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.buildKey(key)
	}

	return c.client.Del(ctx, prefixedKeys...).Err()
}

// Remember executes the function and caches the result if cache miss occurs
func (c *Cache) Remember(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	data, err := c.client.Get(ctx, c.buildKey(key)).Result()
	if err == nil {
		// Cache hit - return the data as string (or deserialize if needed)
		return data, nil
	}
	if err != redis.Nil {
		// Return error if it's not a cache miss
		return nil, fmt.Errorf("cache get error: %w", err)
	}

	// Cache miss - execute function
	value, err := fn()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := c.Set(ctx, key, value, ttl); err != nil {
		// Log error but don't fail the operation
		// The value was computed successfully
		return value, fmt.Errorf("cache set error (value returned): %w", err)
	}

	return value, nil
}

// RememberForever is like Remember but without expiration
func (c *Cache) RememberForever(ctx context.Context, key string, fn func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	data, err := c.client.Get(ctx, c.buildKey(key)).Result()
	if err == nil {
		// Cache hit - return the data
		return data, nil
	}
	if err != redis.Nil {
		// Return error if it's not a cache miss
		return nil, fmt.Errorf("cache get error: %w", err)
	}

	// Cache miss - execute function
	value, err := fn()
	if err != nil {
		return nil, err
	}

	// Store in cache without expiration (0 TTL means no expiration for Redis)
	serialized, err := c.serializer.Serialize(value)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize value: %w", err)
	}

	if err := c.client.Set(ctx, c.buildKey(key), serialized, 0).Err(); err != nil {
		return value, fmt.Errorf("cache set error (value returned): %w", err)
	}

	return value, nil
}

// DeleteByPattern deletes all keys matching the pattern using SCAN
func (c *Cache) DeleteByPattern(ctx context.Context, pattern string) (int, error) {
	prefixedPattern := c.buildKey(pattern)
	iterator := NewIterator(c.client, prefixedPattern, 100)

	count := 0
	batch := make([]string, 0, 100)

	for iterator.Next(ctx) {
		batch = append(batch, iterator.Key())

		// Delete in batches of 100
		if len(batch) >= 100 {
			if err := c.client.Del(ctx, batch...).Err(); err != nil {
				return count, fmt.Errorf("failed to delete batch: %w", err)
			}
			count += len(batch)
			batch = batch[:0]
		}
	}

	// Delete remaining keys
	if len(batch) > 0 {
		if err := c.client.Del(ctx, batch...).Err(); err != nil {
			return count, fmt.Errorf("failed to delete final batch: %w", err)
		}
		count += len(batch)
	}

	if err := iterator.Err(); err != nil {
		return count, err
	}

	return count, nil
}

// FlushPrefix deletes all keys with the cache prefix
func (c *Cache) FlushPrefix(ctx context.Context) error {
	pattern := c.buildKey("*")
	_, err := c.DeleteByPattern(ctx, "*")
	if err != nil {
		return fmt.Errorf("failed to flush prefix %s: %w", pattern, err)
	}
	return nil
}

// Increment increments a key by 1
func (c *Cache) Increment(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, c.buildKey(key)).Result()
}

// IncrementBy increments a key by the specified value
func (c *Cache) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, c.buildKey(key), value).Result()
}

// Decrement decrements a key by 1
func (c *Cache) Decrement(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, c.buildKey(key)).Result()
}

// DecrementBy decrements a key by the specified value
func (c *Cache) DecrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.DecrBy(ctx, c.buildKey(key), value).Result()
}

// IncrementFloat increments a key by a float value
func (c *Cache) IncrementFloat(ctx context.Context, key string, value float64) (float64, error) {
	return c.client.IncrByFloat(ctx, c.buildKey(key), value).Result()
}

// GetTTL returns the remaining TTL for a key
func (c *Cache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, c.buildKey(key)).Result()
}

// Expire sets an expiration time on a key
func (c *Cache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.client.Expire(ctx, c.buildKey(key), ttl).Err()
}

// ExpireAt sets an expiration time at a specific timestamp
func (c *Cache) ExpireAt(ctx context.Context, key string, tm time.Time) error {
	return c.client.ExpireAt(ctx, c.buildKey(key), tm).Err()
}

// Persist removes the expiration from a key
func (c *Cache) Persist(ctx context.Context, key string) error {
	return c.client.Persist(ctx, c.buildKey(key)).Err()
}

// Touch refreshes the TTL to the default value
func (c *Cache) Touch(ctx context.Context, key string) error {
	return c.Expire(ctx, key, c.config.DefaultTTL)
}

// Exists checks if one or more keys exist
func (c *Cache) Exists(ctx context.Context, keys ...string) (int, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.buildKey(key)
	}

	count, err := c.client.Exists(ctx, prefixedKeys...).Result()
	return int(count), err
}

// Keys returns all keys matching the pattern (uses SCAN internally)
func (c *Cache) Keys(ctx context.Context, pattern string) ([]string, error) {
	prefixedPattern := c.buildKey(pattern)
	iterator := NewIterator(c.client, prefixedPattern, 100)

	keys := make([]string, 0)
	for iterator.Next(ctx) {
		keys = append(keys, c.stripPrefix(iterator.Key()))
	}

	if err := iterator.Err(); err != nil {
		return keys, err
	}

	return keys, nil
}

// Scan returns an iterator for keys matching the pattern
func (c *Cache) Scan(ctx context.Context, pattern string, count int64) Iterator {
	prefixedPattern := c.buildKey(pattern)
	return NewIterator(c.client, prefixedPattern, count)
}

// Rename renames a key
func (c *Cache) Rename(ctx context.Context, oldKey, newKey string) error {
	return c.client.Rename(ctx, c.buildKey(oldKey), c.buildKey(newKey)).Err()
}

// Lock creates a distributed lock
func (c *Cache) Lock(key string, ttl time.Duration) Lock {
	return NewRedisLock(c.client, c.buildKey(key), ttl)
}

// WithLock executes a function while holding a distributed lock
func (c *Cache) WithLock(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	lock := c.Lock(key, ttl)

	if err := lock.Acquire(ctx, 10*time.Second); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer func() {
		// Release lock and ignore errors since we're already returning
		// Any error here would be due to the lock expiring or being released by another process
		_ = lock.Release(ctx)
	}()

	return fn()
}
