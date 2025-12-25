package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Iterator provides safe iteration over Redis keys using SCAN
type Iterator interface {
	Next(ctx context.Context) bool
	Key() string
	Err() error
	Close() error
}

// RedisIterator implements Iterator using Redis SCAN command
type RedisIterator struct {
	client  *redis.Client
	pattern string
	count   int64
	cursor  uint64
	keys    []string
	current string
	err     error
	pos     int
	done    bool
}

// NewIterator creates a new iterator for keys matching the pattern
func NewIterator(client *redis.Client, pattern string, count int64) *RedisIterator {
	if count <= 0 {
		count = 10 // Default batch size
	}

	return &RedisIterator{
		client:  client,
		pattern: pattern,
		count:   count,
		cursor:  0,
		keys:    make([]string, 0),
		pos:     0,
		done:    false,
	}
}

// Next advances to the next key. Returns true if a key is available.
func (it *RedisIterator) Next(ctx context.Context) bool {
	if it.err != nil {
		return false
	}

	// If we have keys in buffer, return the next one
	if it.pos < len(it.keys) {
		it.current = it.keys[it.pos]
		it.pos++
		return true
	}

	// If we've already completed scanning, no more keys
	if it.done {
		return false
	}

	// Fetch next batch
	var keys []string
	var newCursor uint64
	var err error

	keys, newCursor, err = it.client.Scan(ctx, it.cursor, it.pattern, it.count).Result()
	if err != nil {
		it.err = fmt.Errorf("scan failed: %w", err)
		return false
	}

	it.cursor = newCursor
	it.keys = keys
	it.pos = 0

	// Check if scan is complete (cursor returns to 0)
	if it.cursor == 0 {
		it.done = true
	}

	// If no keys in this batch, try next batch if not done
	if len(it.keys) == 0 {
		if it.done {
			return false
		}
		// Recursively try next batch
		return it.Next(ctx)
	}

	// Return first key from the batch
	it.current = it.keys[it.pos]
	it.pos++
	return true
}

// Key returns the current key
func (it *RedisIterator) Key() string {
	return it.current
}

// Err returns any error that occurred during iteration
func (it *RedisIterator) Err() error {
	return it.err
}

// Close cleans up resources (no-op for Redis iterator)
func (it *RedisIterator) Close() error {
	it.done = true
	it.keys = nil
	return nil
}
