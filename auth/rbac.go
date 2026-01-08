package auth

import (
	"fmt"
	"strings"
)

// Permission represents a granular permission in the system
type Permission struct {
	Resource string // e.g., "user", "tenant", "document"
	Action   string // e.g., "read", "write", "delete", "create"
	Scope    string // e.g., "own", "tenant", "all", "*"
}

// String returns the permission as a string (resource:action:scope or resource.action)
func (p Permission) String() string {
	if p.Scope != "" && p.Scope != "*" {
		return fmt.Sprintf("%s:%s:%s", p.Resource, p.Action, p.Scope)
	}
	return fmt.Sprintf("%s.%s", p.Resource, p.Action)
}

// ParsePermission parses a permission string into a Permission struct
// Supports formats:
// - "resource.action" (e.g., "user.read")
// - "resource:action:scope" (e.g., "user:read:own")
// - "resource.*" (wildcard action)
// - "*" (super admin - all permissions)
func ParsePermission(permStr string) (Permission, error) {
	if permStr == "*" {
		return Permission{Resource: "*", Action: "*", Scope: "*"}, nil
	}

	// Check for colon format first (resource:action:scope)
	if strings.Contains(permStr, ":") {
		parts := strings.Split(permStr, ":")
		if len(parts) == 2 {
			return Permission{Resource: parts[0], Action: parts[1], Scope: "*"}, nil
		}
		if len(parts) == 3 {
			return Permission{Resource: parts[0], Action: parts[1], Scope: parts[2]}, nil
		}
		return Permission{}, fmt.Errorf("invalid permission format: %s", permStr)
	}

	// Check for dot format (resource.action)
	if strings.Contains(permStr, ".") {
		parts := strings.Split(permStr, ".")
		if len(parts) != 2 {
			return Permission{}, fmt.Errorf("invalid permission format: %s", permStr)
		}
		return Permission{Resource: parts[0], Action: parts[1], Scope: "*"}, nil
	}

	return Permission{}, fmt.Errorf("invalid permission format: %s", permStr)
}

// Matches checks if this permission matches another permission
// Supports wildcard matching
func (p Permission) Matches(other Permission) bool {
	// Super admin wildcard
	if p.Resource == "*" {
		return true
	}

	// Resource match
	if p.Resource != other.Resource && p.Resource != "*" {
		return false
	}

	// Action match (wildcard or exact)
	if p.Action != other.Action && p.Action != "*" {
		return false
	}

	// Scope match
	// "*" scope matches everything
	// "all" scope matches everything
	// "tenant" scope matches tenant and own
	// "own" scope only matches own
	if p.Scope == "*" || p.Scope == "all" {
		return true
	}

	if p.Scope == "tenant" && (other.Scope == "tenant" || other.Scope == "own") {
		return true
	}

	return p.Scope == other.Scope
}

// PermissionSet represents a collection of permissions
type PermissionSet struct {
	permissions map[string]Permission
}

// NewPermissionSet creates a new permission set
func NewPermissionSet(permStrings []string) (*PermissionSet, error) {
	ps := &PermissionSet{
		permissions: make(map[string]Permission),
	}

	for _, permStr := range permStrings {
		perm, err := ParsePermission(permStr)
		if err != nil {
			return nil, err
		}
		ps.permissions[perm.String()] = perm
	}

	return ps, nil
}

// Has checks if the permission set has a specific permission
func (ps *PermissionSet) Has(permStr string) bool {
	required, err := ParsePermission(permStr)
	if err != nil {
		return false
	}

	// Check for exact match first
	if _, exists := ps.permissions[required.String()]; exists {
		return true
	}

	// Check for wildcard matches
	for _, perm := range ps.permissions {
		if perm.Matches(required) {
			return true
		}
	}

	return false
}

// HasAll checks if the permission set has all specified permissions
func (ps *PermissionSet) HasAll(permStrings ...string) bool {
	for _, permStr := range permStrings {
		if !ps.Has(permStr) {
			return false
		}
	}
	return true
}

// HasAny checks if the permission set has any of the specified permissions
func (ps *PermissionSet) HasAny(permStrings ...string) bool {
	for _, permStr := range permStrings {
		if ps.Has(permStr) {
			return true
		}
	}
	return false
}

// Add adds a permission to the set
func (ps *PermissionSet) Add(permStr string) error {
	perm, err := ParsePermission(permStr)
	if err != nil {
		return err
	}
	ps.permissions[perm.String()] = perm
	return nil
}

// Remove removes a permission from the set
func (ps *PermissionSet) Remove(permStr string) {
	perm, err := ParsePermission(permStr)
	if err != nil {
		return
	}
	delete(ps.permissions, perm.String())
}

// List returns all permissions as strings
func (ps *PermissionSet) List() []string {
	result := make([]string, 0, len(ps.permissions))
	for key := range ps.permissions {
		result = append(result, key)
	}
	return result
}

// IsEmpty checks if the permission set is empty
func (ps *PermissionSet) IsEmpty() bool {
	return len(ps.permissions) == 0
}

// Count returns the number of permissions
func (ps *PermissionSet) Count() int {
	return len(ps.permissions)
}

// RBACChecker provides role-based access control checking
type RBACChecker struct {
	permissions *PermissionSet
	roles       []string
}

// NewRBACChecker creates a new RBAC checker
func NewRBACChecker(roles []string, permissions []string) (*RBACChecker, error) {
	permSet, err := NewPermissionSet(permissions)
	if err != nil {
		return nil, err
	}

	return &RBACChecker{
		permissions: permSet,
		roles:       roles,
	}, nil
}

// HasPermission checks if user has a specific permission
func (r *RBACChecker) HasPermission(permission string) bool {
	return r.permissions.Has(permission)
}

// HasAllPermissions checks if user has all specified permissions
func (r *RBACChecker) HasAllPermissions(permissions ...string) bool {
	return r.permissions.HasAll(permissions...)
}

// HasAnyPermission checks if user has any of the specified permissions
func (r *RBACChecker) HasAnyPermission(permissions ...string) bool {
	return r.permissions.HasAny(permissions...)
}

// HasRole checks if user has a specific role
func (r *RBACChecker) HasRole(role string) bool {
	for _, r := range r.roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if user has any of the specified roles
func (r *RBACChecker) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if r.HasRole(role) {
			return true
		}
	}
	return false
}

// HasAllRoles checks if user has all specified roles
func (r *RBACChecker) HasAllRoles(roles ...string) bool {
	for _, role := range roles {
		if !r.HasRole(role) {
			return false
		}
	}
	return true
}

// IsAdmin checks if user has admin role
func (r *RBACChecker) IsAdmin() bool {
	return r.HasAnyRole("admin", "administrator", "super_admin")
}

// IsSuperAdmin checks if user has super admin role or wildcard permission
func (r *RBACChecker) IsSuperAdmin() bool {
	return r.HasRole("super_admin") || r.HasPermission("*")
}

// CanAccessResource checks if user can perform action on resource with specific scope
func (r *RBACChecker) CanAccessResource(resource, action, scope string) bool {
	perm := Permission{Resource: resource, Action: action, Scope: scope}
	return r.permissions.Has(perm.String())
}

// GetRoles returns user's roles
func (r *RBACChecker) GetRoles() []string {
	return r.roles
}

// GetPermissions returns user's permissions
func (r *RBACChecker) GetPermissions() []string {
	return r.permissions.List()
}

// Common permission constants
const (
	// Wildcard
	PermissionWildcard = "*"

	// User permissions
	PermissionUserRead      = "user.read"
	PermissionUserReadOwn   = "user:read:own"
	PermissionUserWrite     = "user.write"
	PermissionUserWriteOwn  = "user:write:own"
	PermissionUserDelete    = "user.delete"
	PermissionUserDeleteOwn = "user:delete:own"
	PermissionUserCreate    = "user.create"
	PermissionUserManage    = "user.manage"

	// Tenant permissions
	PermissionTenantRead   = "tenant.read"
	PermissionTenantWrite  = "tenant.write"
	PermissionTenantDelete = "tenant.delete"
	PermissionTenantCreate = "tenant.create"
	PermissionTenantManage = "tenant.manage"

	// Role permissions
	PermissionRoleRead   = "role.read"
	PermissionRoleWrite  = "role.write"
	PermissionRoleDelete = "role.delete"
	PermissionRoleCreate = "role.create"
	PermissionRoleManage = "role.manage"

	// System permissions
	PermissionSystemRead   = "system.read"
	PermissionSystemWrite  = "system.write"
	PermissionSystemManage = "system.manage"
)

// Common role constants
const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleUser       = "user"
	RoleGuest      = "guest"
	RoleModerator  = "moderator"
	RoleEditor     = "editor"
	RoleViewer     = "viewer"
)
