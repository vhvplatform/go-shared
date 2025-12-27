package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// MGet gets multiple keys in a single operation (performance optimization)
func (c *Client) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	return c.Client.MGet(ctx, keys...).Result()
}

// MSet sets multiple key-value pairs in a single operation (performance optimization)
func (c *Client) MSet(ctx context.Context, pairs ...interface{}) error {
	return c.Client.MSet(ctx, pairs...).Err()
}

// Pipeline creates a pipeline for batch operations
// Performance: Reduces network round trips by batching Redis commands
func (c *Client) Pipeline() redis.Pipeliner {
	return c.Client.Pipeline()
}

// TxPipeline creates a transaction pipeline
// Performance: Executes multiple commands atomically in a single round trip
func (c *Client) TxPipeline() redis.Pipeliner {
	return c.Client.TxPipeline()
}

// SetNX sets a key only if it doesn't exist (useful for distributed locks)
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.Client.SetNX(ctx, key, value, expiration).Result()
}

// GetSet gets old value and sets new value atomically
func (c *Client) GetSet(ctx context.Context, key string, value interface{}) (string, error) {
	return c.Client.GetSet(ctx, key, value).Result()
}

// Incr increments a key's value by 1
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.Client.Incr(ctx, key).Result()
}

// IncrBy increments a key's value by a specified amount
func (c *Client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.Client.IncrBy(ctx, key, value).Result()
}

// Expire sets a timeout on a key
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.Client.Expire(ctx, key, expiration).Result()
}

// TTL returns the remaining time to live of a key
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.Client.TTL(ctx, key).Result()
}
