package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func setupTestRedis() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use a different DB for tests
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	// Clean up test database
	client.FlushDB(ctx)

	return client, nil
}

func TestReplayProtection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client, err := setupTestRedis()
	if err != nil {
		t.Skip("Redis not available, skipping test:", err)
		return
	}
	defer client.Close()

	config := ReplayProtectionConfig{
		RedisClient:     client,
		NonceHeader:     "X-Request-Nonce",
		TimestampHeader: "X-Request-Timestamp",
		TimeWindow:      5 * time.Minute,
		KeyPrefix:       "test:replay:",
	}

	router := gin.New()
	router.Use(ReplayProtection(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	t.Run("Valid request with nonce and timestamp", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-Nonce", "unique-nonce-1")
		req.Header.Set("X-Request-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Missing nonce header", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Missing timestamp header", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-Nonce", "unique-nonce-2")

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Invalid timestamp format", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-Nonce", "unique-nonce-3")
		req.Header.Set("X-Request-Timestamp", "invalid")

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Future timestamp", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-Nonce", "unique-nonce-4")
		futureTime := time.Now().Add(1 * time.Hour).Unix()
		req.Header.Set("X-Request-Timestamp", strconv.FormatInt(futureTime, 10))

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Expired timestamp", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-Nonce", "unique-nonce-5")
		expiredTime := time.Now().Add(-10 * time.Minute).Unix()
		req.Header.Set("X-Request-Timestamp", strconv.FormatInt(expiredTime, 10))

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Replay attack detection", func(t *testing.T) {
		nonce := "unique-nonce-6"
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)

		// First request should succeed
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("GET", "/test", nil)
		req1.Header.Set("X-Request-Nonce", nonce)
		req1.Header.Set("X-Request-Timestamp", timestamp)

		router.ServeHTTP(w1, req1)

		if w1.Code != http.StatusOK {
			t.Errorf("First request: Expected status 200, got %d", w1.Code)
		}

		// Second request with same nonce should fail
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/test", nil)
		req2.Header.Set("X-Request-Nonce", nonce)
		req2.Header.Set("X-Request-Timestamp", timestamp)

		router.ServeHTTP(w2, req2)

		if w2.Code != http.StatusBadRequest {
			t.Errorf("Second request: Expected status 400, got %d", w2.Code)
		}
	})

	t.Run("Different nonces allowed", func(t *testing.T) {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)

		// First request
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("GET", "/test", nil)
		req1.Header.Set("X-Request-Nonce", "unique-nonce-7")
		req1.Header.Set("X-Request-Timestamp", timestamp)

		router.ServeHTTP(w1, req1)

		if w1.Code != http.StatusOK {
			t.Errorf("First request: Expected status 200, got %d", w1.Code)
		}

		// Second request with different nonce
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/test", nil)
		req2.Header.Set("X-Request-Nonce", "unique-nonce-8")
		req2.Header.Set("X-Request-Timestamp", timestamp)

		router.ServeHTTP(w2, req2)

		if w2.Code != http.StatusOK {
			t.Errorf("Second request: Expected status 200, got %d", w2.Code)
		}
	})
}

func TestReplayProtectionWithHash(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client, err := setupTestRedis()
	if err != nil {
		t.Skip("Redis not available, skipping test:", err)
		return
	}
	defer client.Close()

	config := ReplayProtectionConfig{
		RedisClient: client,
		TimeWindow:  5 * time.Minute,
		KeyPrefix:   "test:replay:hash:",
	}

	router := gin.New()
	router.Use(ReplayProtectionWithHash(config, false))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	t.Run("Duplicate request detection", func(t *testing.T) {
		// First request should succeed
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("POST", "/test", nil)

		router.ServeHTTP(w1, req1)

		if w1.Code != http.StatusOK {
			t.Errorf("First request: Expected status 200, got %d", w1.Code)
		}

		// Second identical request should fail
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("POST", "/test", nil)

		router.ServeHTTP(w2, req2)

		if w2.Code != http.StatusBadRequest {
			t.Errorf("Second request: Expected status 400, got %d", w2.Code)
		}
	})
}

func TestCleanupExpiredNonces(t *testing.T) {
	client, err := setupTestRedis()
	if err != nil {
		t.Skip("Redis not available, skipping test:", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	// Create some test nonces
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("test:replay:nonce-%d", i)
		client.Set(ctx, key, "1", 1*time.Hour)
	}

	// Clean up nonces
	deleted, err := CleanupExpiredNonces(ctx, client, "test:replay:*")
	if err != nil {
		t.Fatalf("Failed to cleanup nonces: %v", err)
	}

	if deleted != 5 {
		t.Errorf("Expected to delete 5 nonces, deleted %d", deleted)
	}
}

func TestBuildNonceKey(t *testing.T) {
	tests := []struct {
		name      string
		prefix    string
		nonce     string
		timestamp string
		userID    string
		expected  string
	}{
		{
			name:      "With user ID",
			prefix:    "replay:",
			nonce:     "abc123",
			timestamp: "1234567890",
			userID:    "user123",
			expected:  "replay:user123:1234567890:abc123",
		},
		{
			name:      "Without user ID",
			prefix:    "replay:",
			nonce:     "abc123",
			timestamp: "1234567890",
			userID:    "",
			expected:  "replay:1234567890:abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildNonceKey(tt.prefix, tt.nonce, tt.timestamp, tt.userID)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
