package cache

import "errors"

var (
	// ErrCacheMiss is returned when a key is not found in the cache
	ErrCacheMiss = errors.New("cache: key not found")

	// ErrInvalidCacheValue is returned when the cached value cannot be deserialized
	ErrInvalidCacheValue = errors.New("cache: invalid cache value")

	// ErrCacheUnavailable is returned when the cache backend is unavailable
	ErrCacheUnavailable = errors.New("cache: backend unavailable")
)
