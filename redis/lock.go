package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	// ErrLockNotAcquired is returned when a lock cannot be acquired
	ErrLockNotAcquired = errors.New("lock not acquired")

	// ErrLockNotHeld is returned when trying to release a lock that is not held
	ErrLockNotHeld = errors.New("lock not held")
)

// Lua script for releasing a lock (atomic compare-and-delete)
var releaseLockScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
else
    return 0
end
`)

// Lua script for extending a lock (atomic compare-and-expire)
var extendLockScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("expire", KEYS[1], ARGV[2])
else
    return 0
end
`)

// Lock defines the interface for distributed locking
type Lock interface {
	Acquire(ctx context.Context, timeout time.Duration) error
	Release(ctx context.Context) error
	Extend(ctx context.Context, duration time.Duration) error
	IsLocked(ctx context.Context) (bool, error)
}

// RedisLock implements distributed locking using Redis
type RedisLock struct {
	client *redis.Client
	key    string
	token  string
	ttl    time.Duration
}

// NewRedisLock creates a new distributed lock
func NewRedisLock(client *redis.Client, key string, ttl time.Duration) *RedisLock {
	return &RedisLock{
		client: client,
		key:    key,
		token:  uuid.New().String(),
		ttl:    ttl,
	}
}

// Acquire attempts to acquire the lock with retry logic
func (l *RedisLock) Acquire(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		// Try to acquire lock using SetNX
		acquired, err := l.client.SetNX(ctx, l.key, l.token, l.ttl).Result()
		if err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}

		if acquired {
			return nil
		}

		// Check if we've exceeded the timeout
		if time.Now().After(deadline) {
			return ErrLockNotAcquired
		}

		// Wait before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Continue to next iteration
		}
	}
}

// Release releases the lock using a Lua script for atomic compare-and-delete
func (l *RedisLock) Release(ctx context.Context) error {
	result, err := releaseLockScript.Run(ctx, l.client, []string{l.key}, l.token).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	// Check if the lock was actually deleted
	deleted, ok := result.(int64)
	if !ok || deleted == 0 {
		return ErrLockNotHeld
	}

	return nil
}

// Extend extends the lock TTL using a Lua script for atomic operation
func (l *RedisLock) Extend(ctx context.Context, duration time.Duration) error {
	ttlSeconds := int64(duration.Seconds())
	result, err := extendLockScript.Run(ctx, l.client, []string{l.key}, l.token, ttlSeconds).Result()
	if err != nil {
		return fmt.Errorf("failed to extend lock: %w", err)
	}

	// Check if the lock was actually extended
	extended, ok := result.(int64)
	if !ok || extended == 0 {
		return ErrLockNotHeld
	}

	l.ttl = duration
	return nil
}

// IsLocked checks if the lock is currently held by this instance
func (l *RedisLock) IsLocked(ctx context.Context) (bool, error) {
	value, err := l.client.Get(ctx, l.key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check lock status: %w", err)
	}

	return value == l.token, nil
}

// RefreshLoop automatically refreshes the lock in the background
// It returns a channel that will be closed when the refresh loop stops.
// The caller should cancel the context to stop the refresh loop.
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	errChan := lock.RefreshLoop(ctx, 10*time.Second)
func (l *RedisLock) RefreshLoop(ctx context.Context, interval time.Duration) <-chan error {
	errChan := make(chan error, 1)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		defer close(errChan)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := l.Extend(ctx, l.ttl); err != nil {
					errChan <- err
					return
				}
			}
		}
	}()

	return errChan
}
