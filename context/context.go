package context

import (
	"context"
	"errors"
)

type contextKey string

const (
	UserIDKey        contextKey = "user_id"
	TenantIDKey      contextKey = "tenant_id"
	AppIDKey         contextKey = "app_id"
	RolesKey         contextKey = "roles"
	PermissionsKey   contextKey = "permissions"
	EmailKey         contextKey = "email"
	CorrelationIDKey contextKey = "correlation_id"
	TenantDomainKey  contextKey = "tenant_domain"
	// RequestCtxKey caches the full request context to avoid repeated field lookups during retrieval
	RequestCtxKey contextKey = "request_context"
)

var (
	ErrUserNotFound        = errors.New("user not found in context")
	ErrTenantNotFound      = errors.New("tenant not found in context")
	ErrPermissionsNotFound = errors.New("permissions not found in context")
)

// RequestContext holds all context information for a request
type RequestContext struct {
	UserID        string
	TenantID      string
	AppID         string
	Email         string
	Roles         []string
	Permissions   []string
	CorrelationID string
	TenantDomain  string
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserID retrieves user ID from context
func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return "", ErrUserNotFound
	}
	return userID, nil
}

// MustGetUserID retrieves user ID or panics
func MustGetUserID(ctx context.Context) string {
	userID, err := GetUserID(ctx)
	if err != nil {
		panic(err)
	}
	return userID
}

// WithTenantID adds tenant ID to context
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// GetTenantID retrieves tenant ID from context
func GetTenantID(ctx context.Context) (string, error) {
	tenantID, ok := ctx.Value(TenantIDKey).(string)
	if !ok || tenantID == "" {
		return "", ErrTenantNotFound
	}
	return tenantID, nil
}

// MustGetTenantID retrieves tenant ID or panics
func MustGetTenantID(ctx context.Context) string {
	tenantID, err := GetTenantID(ctx)
	if err != nil {
		panic(err)
	}
	return tenantID
}

// WithAppID adds application ID to context
func WithAppID(ctx context.Context, appID string) context.Context {
	return context.WithValue(ctx, AppIDKey, appID)
}

// GetAppID retrieves application ID from context
func GetAppID(ctx context.Context) string {
	appID, _ := ctx.Value(AppIDKey).(string)
	return appID
}

// WithRoles adds roles to context
func WithRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, RolesKey, roles)
}

// GetRoles retrieves roles from context
func GetRoles(ctx context.Context) []string {
	roles, ok := ctx.Value(RolesKey).([]string)
	if !ok {
		return []string{}
	}
	return roles
}

// WithPermissions adds permissions to context
func WithPermissions(ctx context.Context, permissions []string) context.Context {
	return context.WithValue(ctx, PermissionsKey, permissions)
}

// GetPermissions retrieves permissions from context
func GetPermissions(ctx context.Context) ([]string, error) {
	permissions, ok := ctx.Value(PermissionsKey).([]string)
	if !ok {
		return nil, ErrPermissionsNotFound
	}
	return permissions, nil
}

// WithEmail adds email to context
func WithEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, EmailKey, email)
}

// GetEmail retrieves email from context
func GetEmail(ctx context.Context) string {
	email, _ := ctx.Value(EmailKey).(string)
	return email
}

// WithCorrelationID adds correlation ID to context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// GetCorrelationID retrieves correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	correlationID, _ := ctx.Value(CorrelationIDKey).(string)
	return correlationID
}

// WithTenantDomain adds tenant domain to context
func WithTenantDomain(ctx context.Context, domain string) context.Context {
	return context.WithValue(ctx, TenantDomainKey, domain)
}

// GetTenantDomain retrieves tenant domain from context
func GetTenantDomain(ctx context.Context) string {
	domain, _ := ctx.Value(TenantDomainKey).(string)
	return domain
}

// WithRequestContext adds full request context
// Performance: Caches the full RequestContext to avoid repeated field lookups
func WithRequestContext(ctx context.Context, rc *RequestContext) context.Context {
	ctx = WithUserID(ctx, rc.UserID)
	ctx = WithTenantID(ctx, rc.TenantID)
	ctx = WithAppID(ctx, rc.AppID)
	ctx = WithEmail(ctx, rc.Email)
	ctx = WithRoles(ctx, rc.Roles)
	ctx = WithPermissions(ctx, rc.Permissions)
	ctx = WithCorrelationID(ctx, rc.CorrelationID)
	ctx = WithTenantDomain(ctx, rc.TenantDomain)
	// Store complete context for faster retrieval
	ctx = context.WithValue(ctx, RequestCtxKey, rc)
	return ctx
}

// GetRequestContext retrieves full request context
// Performance: Returns cached RequestContext if available, avoiding multiple lookups
func GetRequestContext(ctx context.Context) *RequestContext {
	// Try to get cached context first (performance optimization)
	if rc, ok := ctx.Value(RequestCtxKey).(*RequestContext); ok && rc != nil {
		return rc
	}

	// Fallback to building from individual values
	userID, _ := GetUserID(ctx)
	tenantID, _ := GetTenantID(ctx)
	permissions, _ := GetPermissions(ctx)

	return &RequestContext{
		UserID:        userID,
		TenantID:      tenantID,
		AppID:         GetAppID(ctx),
		Email:         GetEmail(ctx),
		Roles:         GetRoles(ctx),
		Permissions:   permissions,
		CorrelationID: GetCorrelationID(ctx),
		TenantDomain:  GetTenantDomain(ctx),
	}
}
