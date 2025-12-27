package middleware

import (
	"context"
	"testing"
)

func BenchmarkRateLimiterGetLimiter(b *testing.B) {
	rl := NewRateLimiter(100, 10)
	ctx := context.Background()
	go rl.CleanupLimiters(ctx)
	defer rl.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rl.GetLimiter("test-key")
	}
}

func BenchmarkRateLimiterGetLimiterParallel(b *testing.B) {
	rl := NewRateLimiter(100, 10)
	ctx := context.Background()
	go rl.CleanupLimiters(ctx)
	defer rl.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Use strconv.Itoa to generate proper string keys
			key := "test-key-" + string(rune('0'+i%10))
			_ = rl.GetLimiter(key)
			i++
		}
	})
}

func BenchmarkRateLimiterAllow(b *testing.B) {
	rl := NewRateLimiter(1000, 100)
	ctx := context.Background()
	go rl.CleanupLimiters(ctx)
	defer rl.Stop()

	limiter := rl.GetLimiter("test-key")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = limiter.Allow()
	}
}
