package cache

import (
	"context"
	"time"
)

// TieredCache implements a two-tier caching strategy with L1 (local) and L2 (distributed) caches
type TieredCache struct {
	l1 Cache // Local cache (e.g., Ristretto)
	l2 Cache // Distributed cache (e.g., Redis)
}

// TieredCacheConfig holds configuration for tiered cache
type TieredCacheConfig struct {
	L1TTLCap time.Duration // Maximum TTL for L1 cache (default: 5 minutes)
}

// NewTieredCache creates a new two-tier cache
func NewTieredCache(l1, l2 Cache, config *TieredCacheConfig) *TieredCache {
	if config == nil {
		config = &TieredCacheConfig{
			L1TTLCap: 5 * time.Minute,
		}
	}
	if config.L1TTLCap == 0 {
		config.L1TTLCap = 5 * time.Minute
	}

	return &TieredCache{
		l1: l1,
		l2: l2,
	}
}

// Get tries L1 first, then L2, then backfills L1 on L2 hit
func (c *TieredCache) Get(ctx context.Context, key string, dest interface{}) error {
	// Try L1 first (fast path)
	if err := c.l1.Get(ctx, key, dest); err == nil {
		return nil // L1 hit
	}

	// L1 miss, try L2
	if err := c.l2.Get(ctx, key, dest); err != nil {
		return err // L2 miss
	}

	// L2 hit, backfill L1 asynchronously (fire-and-forget)
	go func() {
		// Use background context to avoid cancellation
		_ = c.l1.Set(context.Background(), key, dest, 5*time.Minute)
	}()

	return nil
}

// Set writes to both L1 and L2
func (c *TieredCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Write to L2 first (source of truth)
	if err := c.l2.Set(ctx, key, value, ttl); err != nil {
		return err
	}

	// Write to L1 asynchronously with capped TTL
	go func() {
		l1TTL := ttl
		if l1TTL > 5*time.Minute {
			l1TTL = 5 * time.Minute // Cap L1 TTL to avoid stale data
		}
		_ = c.l1.Set(context.Background(), key, value, l1TTL)
	}()

	return nil
}

// Delete removes from both caches
func (c *TieredCache) Delete(ctx context.Context, key string) error {
	// Delete from L1 (best effort)
	_ = c.l1.Delete(ctx, key)

	// Delete from L2 (source of truth)
	return c.l2.Delete(ctx, key)
}

// Exists checks both caches
func (c *TieredCache) Exists(ctx context.Context, key string) (bool, error) {
	// Check L1 first
	if exists, err := c.l1.Exists(ctx, key); err == nil && exists {
		return true, nil
	}

	// Check L2
	return c.l2.Exists(ctx, key)
}

// LocalOnly returns the L1 cache for local-only operations
func (c *TieredCache) LocalOnly() Cache {
	return c.l1
}

// DistributedOnly returns the L2 cache for distributed operations
func (c *TieredCache) DistributedOnly() Cache {
	return c.l2
}
