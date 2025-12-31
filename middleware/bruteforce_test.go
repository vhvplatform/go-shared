package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestBruteForceProtection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client, err := setupTestRedis()
	if err != nil {
		t.Skip("Redis not available, skipping test:", err)
		return
	}
	defer client.Close()

	config := BruteForceProtectionConfig{
		RedisClient:     client,
		MaxAttempts:     3,
		LockoutDuration: 1 * time.Minute,
		AttemptWindow:   5 * time.Minute,
		KeyPrefix:       "test:bf:",
	}

	router := gin.New()
	router.Use(BruteForceProtection(config))
	router.POST("/login", func(c *gin.Context) {
		// Simulate authentication
		username := c.PostForm("username")
		password := c.PostForm("password")

		if username == "admin" && password == "correct" {
			RecordSuccessfulAttempt(c)
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		} else {
			RecordFailedAttempt(c)
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid credentials"})
		}
	})

	t.Run("Successful login", func(t *testing.T) {
		// Clean up
		client.FlushDB(context.Background())

		w := httptest.NewRecorder()
		body := strings.NewReader("username=admin&password=correct")
		req, _ := http.NewRequest("POST", "/login", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Failed login attempts", func(t *testing.T) {
		// Clean up
		client.FlushDB(context.Background())

		// First failed attempt
		for i := 1; i <= 3; i++ {
			w := httptest.NewRecorder()
			body := strings.NewReader("username=admin&password=wrong")
			req, _ := http.NewRequest("POST", "/login", body)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("Attempt %d: Expected status 401, got %d", i, w.Code)
			}
		}

		// Fourth attempt should be blocked
		w := httptest.NewRecorder()
		body := strings.NewReader("username=admin&password=wrong")
		req, _ := http.NewRequest("POST", "/login", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		router.ServeHTTP(w, req)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status 429 (locked out), got %d", w.Code)
		}
	})

	t.Run("Lockout persists", func(t *testing.T) {
		// Clean up
		client.FlushDB(context.Background())

		// Exhaust attempts
		for i := 1; i <= 4; i++ {
			w := httptest.NewRecorder()
			body := strings.NewReader("username=testuser&password=wrong")
			req, _ := http.NewRequest("POST", "/login", body)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			router.ServeHTTP(w, req)
		}

		// Try again after lockout
		w := httptest.NewRecorder()
		body := strings.NewReader("username=testuser&password=correct")
		req, _ := http.NewRequest("POST", "/login", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		router.ServeHTTP(w, req)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status 429 (still locked), got %d", w.Code)
		}
	})

	t.Run("Successful login resets counter", func(t *testing.T) {
		// Clean up
		client.FlushDB(context.Background())

		// Make 2 failed attempts
		for i := 1; i <= 2; i++ {
			w := httptest.NewRecorder()
			body := strings.NewReader("username=admin&password=wrong")
			req, _ := http.NewRequest("POST", "/login", body)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			router.ServeHTTP(w, req)
		}

		// Successful login should reset counter
		w := httptest.NewRecorder()
		body := strings.NewReader("username=admin&password=correct")
		req, _ := http.NewRequest("POST", "/login", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Should be able to make attempts again
		w2 := httptest.NewRecorder()
		body2 := strings.NewReader("username=admin&password=wrong")
		req2, _ := http.NewRequest("POST", "/login", body2)
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		router.ServeHTTP(w2, req2)

		if w2.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w2.Code)
		}
	})
}

func TestBruteForcePerUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client, err := setupTestRedis()
	if err != nil {
		t.Skip("Redis not available, skipping test:", err)
		return
	}
	defer client.Close()

	config := BruteForceProtectionConfig{
		RedisClient:     client,
		MaxAttempts:     3,
		LockoutDuration: 1 * time.Minute,
		AttemptWindow:   5 * time.Minute,
		KeyPrefix:       "test:bf:user:",
	}

	router := gin.New()
	router.Use(BruteForcePerUser(config, "username"))
	router.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		if username == "admin" && password == "correct" {
			RecordSuccessfulAttempt(c)
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		} else {
			RecordFailedAttempt(c)
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid credentials"})
		}
	})

	t.Run("Different users have separate counters", func(t *testing.T) {
		// Clean up
		client.FlushDB(context.Background())

		// User1 makes 3 failed attempts
		for i := 1; i <= 3; i++ {
			w := httptest.NewRecorder()
			body := strings.NewReader("username=user1&password=wrong")
			req, _ := http.NewRequest("POST", "/login", body)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			router.ServeHTTP(w, req)
		}

		// User1 should be locked out
		w1 := httptest.NewRecorder()
		body1 := strings.NewReader("username=user1&password=wrong")
		req1, _ := http.NewRequest("POST", "/login", body1)
		req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		router.ServeHTTP(w1, req1)

		if w1.Code != http.StatusTooManyRequests {
			t.Errorf("User1: Expected status 429, got %d", w1.Code)
		}

		// User2 should still be able to try
		w2 := httptest.NewRecorder()
		body2 := strings.NewReader("username=user2&password=wrong")
		req2, _ := http.NewRequest("POST", "/login", body2)
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		router.ServeHTTP(w2, req2)

		if w2.Code != http.StatusUnauthorized {
			t.Errorf("User2: Expected status 401, got %d", w2.Code)
		}
	})
}

func TestBruteForcePerEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client, err := setupTestRedis()
	if err != nil {
		t.Skip("Redis not available, skipping test:", err)
		return
	}
	defer client.Close()

	config := BruteForceProtectionConfig{
		RedisClient:     client,
		MaxAttempts:     3,
		LockoutDuration: 1 * time.Minute,
		AttemptWindow:   5 * time.Minute,
		KeyPrefix:       "test:bf:email:",
	}

	router := gin.New()
	router.Use(BruteForcePerEmail(config))
	router.POST("/login", func(c *gin.Context) {
		email := c.PostForm("email")
		password := c.PostForm("password")

		if email == "admin@example.com" && password == "correct" {
			RecordSuccessfulAttempt(c)
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		} else {
			RecordFailedAttempt(c)
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid credentials"})
		}
	})

	t.Run("Email-based tracking", func(t *testing.T) {
		// Clean up
		client.FlushDB(context.Background())

		// Make 3 failed attempts
		for i := 1; i <= 3; i++ {
			w := httptest.NewRecorder()
			body := strings.NewReader("email=test@example.com&password=wrong")
			req, _ := http.NewRequest("POST", "/login", body)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			router.ServeHTTP(w, req)
		}

		// Should be locked out
		w := httptest.NewRecorder()
		body := strings.NewReader("email=test@example.com&password=wrong")
		req, _ := http.NewRequest("POST", "/login", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		router.ServeHTTP(w, req)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status 429, got %d", w.Code)
		}
	})
}

func TestGetRemainingAttempts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client, err := setupTestRedis()
	if err != nil {
		t.Skip("Redis not available, skipping test:", err)
		return
	}
	defer client.Close()

	config := BruteForceProtectionConfig{
		RedisClient:     client,
		MaxAttempts:     5,
		LockoutDuration: 1 * time.Minute,
		AttemptWindow:   5 * time.Minute,
		KeyPrefix:       "test:bf:",
	}

	router := gin.New()
	router.Use(BruteForceProtection(config))
	router.POST("/login", func(c *gin.Context) {
		remaining, _ := GetRemainingAttempts(c)

		password := c.PostForm("password")
		if password == "correct" {
			RecordSuccessfulAttempt(c)
			c.JSON(http.StatusOK, gin.H{"message": "success", "remaining": remaining})
		} else {
			RecordFailedAttempt(c)
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid", "remaining": remaining})
		}
	})

	t.Run("Remaining attempts decreases", func(t *testing.T) {
		// Clean up
		client.FlushDB(context.Background())

		// First attempt
		w := httptest.NewRecorder()
		body := strings.NewReader("password=wrong")
		req, _ := http.NewRequest("POST", "/login", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		router.ServeHTTP(w, req)

		if !strings.Contains(w.Body.String(), "\"remaining\":5") {
			t.Errorf("Expected remaining: 5, got %s", w.Body.String())
		}
	})
}

func TestResetBruteForceProtection(t *testing.T) {
	client, err := setupTestRedis()
	if err != nil {
		t.Skip("Redis not available, skipping test:", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	// Set up some test data
	client.Set(ctx, "test:bf:lock:user123", "1", 1*time.Hour)
	client.Set(ctx, "test:bf:attempts:user123", "5", 1*time.Hour)

	// Reset
	err = ResetBruteForceProtection(ctx, client, "test:bf:", "user123")
	if err != nil {
		t.Fatalf("Failed to reset: %v", err)
	}

	// Verify data is removed
	lockExists, _ := client.Exists(ctx, "test:bf:lock:user123").Result()
	attemptsExists, _ := client.Exists(ctx, "test:bf:attempts:user123").Result()

	if lockExists != 0 {
		t.Error("Lock key should be deleted")
	}
	if attemptsExists != 0 {
		t.Error("Attempts key should be deleted")
	}
}

func TestGetBruteForceStatus(t *testing.T) {
	client, err := setupTestRedis()
	if err != nil {
		t.Skip("Redis not available, skipping test:", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("Not locked, no attempts", func(t *testing.T) {
		client.FlushDB(ctx)

		locked, attempts, ttl, err := GetBruteForceStatus(ctx, client, "test:bf:", "user123")
		if err != nil {
			t.Fatalf("Failed to get status: %v", err)
		}

		if locked {
			t.Error("Should not be locked")
		}
		if attempts != 0 {
			t.Errorf("Expected 0 attempts, got %d", attempts)
		}
		if ttl != 0 {
			t.Errorf("Expected 0 TTL, got %v", ttl)
		}
	})

	t.Run("Locked with attempts", func(t *testing.T) {
		client.FlushDB(ctx)

		client.Set(ctx, "test:bf:lock:user456", "1", 1*time.Minute)
		client.Set(ctx, "test:bf:attempts:user456", "5", 1*time.Hour)

		locked, attempts, ttl, err := GetBruteForceStatus(ctx, client, "test:bf:", "user456")
		if err != nil {
			t.Fatalf("Failed to get status: %v", err)
		}

		if !locked {
			t.Error("Should be locked")
		}
		if attempts != 5 {
			t.Errorf("Expected 5 attempts, got %d", attempts)
		}
		if ttl <= 0 {
			t.Errorf("Expected positive TTL, got %v", ttl)
		}
	})
}

func TestExponentialBackoff(t *testing.T) {
	tests := []struct {
		name            string
		baseDuration    time.Duration
		attempts        int
		maxAttempts     int
		multiplier      float64
		expectedMin     time.Duration
		expectedMax     time.Duration
	}{
		{
			name:            "First overage (attempts=6, max=5, overage=2)",
			baseDuration:    1 * time.Minute,
			attempts:        6,
			maxAttempts:     5,
			multiplier:      2,
			expectedMin:     4 * time.Minute,
			expectedMax:     4 * time.Minute,
		},
		{
			name:            "Second overage (attempts=7, max=5, overage=3)",
			baseDuration:    1 * time.Minute,
			attempts:        7,
			maxAttempts:     5,
			multiplier:      2,
			expectedMin:     8 * time.Minute,
			expectedMax:     8 * time.Minute,
		},
		{
			name:            "Capped at max",
			baseDuration:    1 * time.Hour,
			attempts:        50,
			maxAttempts:     5,
			multiplier:      2,
			expectedMin:     24 * time.Hour,
			expectedMax:     24 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateExponentialBackoff(tt.baseDuration, tt.attempts, tt.maxAttempts, tt.multiplier)

			if result < tt.expectedMin || result > tt.expectedMax {
				t.Errorf("Expected duration between %v and %v, got %v", tt.expectedMin, tt.expectedMax, result)
			}
		})
	}
}
