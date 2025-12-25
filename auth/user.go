package auth

import (
	"context"

	pkgctx "github.com/vhvcorp/go-shared/context"
)

// UserInfo represents current user information
type UserInfo struct {
	ID          string
	Email       string
	TenantID    string
	Roles       []string
	Permissions []string
}

// GetCurrentUser retrieves current user info from context
func GetCurrentUser(ctx context.Context) (*UserInfo, error) {
	userID, err := pkgctx.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	tenantID, err := pkgctx.GetTenantID(ctx)
	if err != nil {
		return nil, err
	}

	permissions, _ := pkgctx.GetPermissions(ctx)

	return &UserInfo{
		ID:          userID,
		Email:       pkgctx.GetEmail(ctx),
		TenantID:    tenantID,
		Roles:       pkgctx.GetRoles(ctx),
		Permissions: permissions,
	}, nil
}

// MustGetCurrentUser retrieves current user or panics
func MustGetCurrentUser(ctx context.Context) *UserInfo {
	user, err := GetCurrentUser(ctx)
	if err != nil {
		panic(err)
	}
	return user
}

// GetCurrentUserID retrieves current user ID
func GetCurrentUserID(ctx context.Context) (string, error) {
	return pkgctx.GetUserID(ctx)
}

// GetCurrentTenantID retrieves current tenant ID
func GetCurrentTenantID(ctx context.Context) (string, error) {
	return pkgctx.GetTenantID(ctx)
}

// IsAuthenticated checks if user is authenticated
func IsAuthenticated(ctx context.Context) bool {
	_, err := pkgctx.GetUserID(ctx)
	return err == nil
}
