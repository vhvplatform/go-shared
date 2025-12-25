package auth

import (
	"context"
	"errors"
	"strings"

	pkgctx "github.com/vhvcorp/go-shared/context"
)

var (
	ErrPermissionDenied = errors.New("permission denied")
	ErrNoPermissions    = errors.New("no permissions found")
)

// PermissionChecker provides permission checking functionality
type PermissionChecker struct{}

// NewPermissionChecker creates a new permission checker
func NewPermissionChecker() *PermissionChecker {
	return &PermissionChecker{}
}

// HasPermission checks if user has a specific permission
func (pc *PermissionChecker) HasPermission(ctx context.Context, permission string) bool {
	permissions, err := pkgctx.GetPermissions(ctx)
	if err != nil {
		return false
	}

	for _, p := range permissions {
		if p == permission || p == "*" {
			return true
		}

		// Check wildcard permissions (e.g., "users.*" matches "users.read")
		if strings.HasSuffix(p, ".*") {
			prefix := strings.TrimSuffix(p, ".*")
			if strings.HasPrefix(permission, prefix+".") {
				return true
			}
		}
	}

	return false
}

// HasAnyPermission checks if user has any of the specified permissions
func (pc *PermissionChecker) HasAnyPermission(ctx context.Context, permissions ...string) bool {
	for _, permission := range permissions {
		if pc.HasPermission(ctx, permission) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if user has all of the specified permissions
func (pc *PermissionChecker) HasAllPermissions(ctx context.Context, permissions ...string) bool {
	for _, permission := range permissions {
		if !pc.HasPermission(ctx, permission) {
			return false
		}
	}
	return true
}

// RequirePermission checks permission and returns error if not authorized
func (pc *PermissionChecker) RequirePermission(ctx context.Context, permission string) error {
	if !pc.HasPermission(ctx, permission) {
		return ErrPermissionDenied
	}
	return nil
}

// HasRole checks if user has a specific role
func (pc *PermissionChecker) HasRole(ctx context.Context, role string) bool {
	roles := pkgctx.GetRoles(ctx)
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if user has any of the specified roles
func (pc *PermissionChecker) HasAnyRole(ctx context.Context, roles ...string) bool {
	for _, role := range roles {
		if pc.HasRole(ctx, role) {
			return true
		}
	}
	return false
}

// IsSuperAdmin checks if user is super admin
func (pc *PermissionChecker) IsSuperAdmin(ctx context.Context) bool {
	return pc.HasRole(ctx, "super_admin")
}

// IsTenantAdmin checks if user is tenant admin
func (pc *PermissionChecker) IsTenantAdmin(ctx context.Context) bool {
	return pc.HasRole(ctx, "tenant_admin") || pc.IsSuperAdmin(ctx)
}

// Global permission checker instance
var GlobalPermissionChecker = NewPermissionChecker()

// Helper functions using global checker
func HasPermission(ctx context.Context, permission string) bool {
	return GlobalPermissionChecker.HasPermission(ctx, permission)
}

func RequirePermission(ctx context.Context, permission string) error {
	return GlobalPermissionChecker.RequirePermission(ctx, permission)
}

func HasRole(ctx context.Context, role string) bool {
	return GlobalPermissionChecker.HasRole(ctx, role)
}

func IsSuperAdmin(ctx context.Context) bool {
	return GlobalPermissionChecker.IsSuperAdmin(ctx)
}

func IsTenantAdmin(ctx context.Context) bool {
	return GlobalPermissionChecker.IsTenantAdmin(ctx)
}
