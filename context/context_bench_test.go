package context

import (
	"context"
	"testing"
)

func BenchmarkWithRequestContext(b *testing.B) {
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
		ctx := context.Background()
		_ = WithRequestContext(ctx, rc)
	}
}

func BenchmarkGetRequestContext(b *testing.B) {
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

	ctx := WithRequestContext(context.Background(), rc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetRequestContext(ctx)
	}
}

func BenchmarkGetUserID(b *testing.B) {
	ctx := WithUserID(context.Background(), "user123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetUserID(ctx)
	}
}

func BenchmarkGetPermissions(b *testing.B) {
	ctx := WithPermissions(context.Background(), []string{"users.read", "users.write", "admin.*"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetPermissions(ctx)
	}
}
