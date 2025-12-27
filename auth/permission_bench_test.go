package auth

import (
	"context"
	"testing"

	pkgctx "github.com/vhvplatform/go-shared/context"
)

func BenchmarkHasPermission(b *testing.B) {
	permissions := []string{"users.read", "users.write", "posts.*", "comments.read"}
	ctx := pkgctx.WithPermissions(context.Background(), permissions)
	pc := NewPermissionChecker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pc.HasPermission(ctx, "users.read")
	}
}

func BenchmarkHasPermissionWildcard(b *testing.B) {
	permissions := []string{"users.read", "users.write", "posts.*", "comments.read"}
	ctx := pkgctx.WithPermissions(context.Background(), permissions)
	pc := NewPermissionChecker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pc.HasPermission(ctx, "posts.create")
	}
}

func BenchmarkHasPermissionMiss(b *testing.B) {
	permissions := []string{"users.read", "users.write", "posts.*", "comments.read"}
	ctx := pkgctx.WithPermissions(context.Background(), permissions)
	pc := NewPermissionChecker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pc.HasPermission(ctx, "admin.delete")
	}
}

func BenchmarkHasAnyPermission(b *testing.B) {
	permissions := []string{"users.read", "users.write", "posts.*", "comments.read"}
	ctx := pkgctx.WithPermissions(context.Background(), permissions)
	pc := NewPermissionChecker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pc.HasAnyPermission(ctx, "admin.delete", "users.read", "logs.view")
	}
}

func BenchmarkHasRole(b *testing.B) {
	roles := []string{"admin", "user", "moderator"}
	ctx := pkgctx.WithRoles(context.Background(), roles)
	pc := NewPermissionChecker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pc.HasRole(ctx, "admin")
	}
}

func BenchmarkIsSuperAdmin(b *testing.B) {
	roles := []string{"admin", "user", "super_admin"}
	ctx := pkgctx.WithRoles(context.Background(), roles)
	pc := NewPermissionChecker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pc.IsSuperAdmin(ctx)
	}
}
