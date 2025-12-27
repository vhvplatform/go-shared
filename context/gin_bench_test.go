package context

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func BenchmarkToGinContext(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	rc := &RequestContext{
		UserID:        "user123",
		TenantID:      "tenant456",
		AppID:         "app789",
		Email:         "user@example.com",
		Roles:         []string{"admin", "user"},
		Permissions:   []string{"users.read", "users.write"},
		CorrelationID: "corr-123",
		TenantDomain:  "example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		b.StartTimer()
		
		ToGinContext(c, rc)
	}
}

func BenchmarkFromGinContext(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	rc := &RequestContext{
		UserID:        "user123",
		TenantID:      "tenant456",
		AppID:         "app789",
		Email:         "user@example.com",
		Roles:         []string{"admin", "user"},
		Permissions:   []string{"users.read", "users.write"},
		CorrelationID: "corr-123",
		TenantDomain:  "example.com",
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	ToGinContext(c, rc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FromGinContext(c)
	}
}

func BenchmarkFromGinContextUncached(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", "user123")
	c.Set("tenant_id", "tenant456")
	c.Set("app_id", "app789")
	c.Set("email", "user@example.com")
	c.Set("roles", []string{"admin", "user"})
	c.Set("permissions", []string{"users.read", "users.write"})
	c.Set("correlation_id", "corr-123")
	c.Set("tenant_domain", "example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FromGinContext(c)
	}
}

func BenchmarkGetUserIDFromGin(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", "user123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetUserIDFromGin(c)
	}
}

func BenchmarkHasRoleFromGin(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("roles", []string{"admin", "user", "moderator"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = HasRoleFromGin(c, "admin")
	}
}
