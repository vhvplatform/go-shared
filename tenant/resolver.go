package tenant

import (
	"context"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	pkgctx "github.com/vhvplatform/go-shared/context"
)

var (
	ErrTenantNotResolved = errors.New("tenant could not be resolved")
	ErrInvalidDomain     = errors.New("invalid domain")
)

// ResolutionStrategy defines how to resolve tenant
type ResolutionStrategy string

const (
	StrategyHeader    ResolutionStrategy = "header"
	StrategySubdomain ResolutionStrategy = "subdomain"
	StrategyDomain    ResolutionStrategy = "domain"
	StrategyParam     ResolutionStrategy = "param"
)

// Resolver resolves tenant from request
type Resolver struct {
	strategies []ResolutionStrategy
	headerName string
	paramName  string
}

// ResolverConfig configures tenant resolver
type ResolverConfig struct {
	Strategies []ResolutionStrategy
	HeaderName string // default: "X-Tenant-ID"
	ParamName  string // default: "tenant_id"
}

// NewResolver creates a new tenant resolver
func NewResolver(config ResolverConfig) *Resolver {
	if len(config.Strategies) == 0 {
		config.Strategies = []ResolutionStrategy{StrategyHeader, StrategySubdomain}
	}
	if config.HeaderName == "" {
		config.HeaderName = "X-Tenant-ID"
	}
	if config.ParamName == "" {
		config.ParamName = "tenant_id"
	}

	return &Resolver{
		strategies: config.Strategies,
		headerName: config.HeaderName,
		paramName:  config.ParamName,
	}
}

// Resolve resolves tenant from gin context
func (r *Resolver) Resolve(c *gin.Context) (string, string, error) {
	for _, strategy := range r.strategies {
		tenantID, domain, err := r.resolveByStrategy(c, strategy)
		if err == nil && tenantID != "" {
			return tenantID, domain, nil
		}
	}
	return "", "", ErrTenantNotResolved
}

func (r *Resolver) resolveByStrategy(c *gin.Context, strategy ResolutionStrategy) (string, string, error) {
	switch strategy {
	case StrategyHeader:
		return r.resolveFromHeader(c)
	case StrategySubdomain:
		return r.resolveFromSubdomain(c)
	case StrategyDomain:
		return r.resolveFromDomain(c)
	case StrategyParam:
		return r.resolveFromParam(c)
	default:
		return "", "", ErrTenantNotResolved
	}
}

func (r *Resolver) resolveFromHeader(c *gin.Context) (string, string, error) {
	tenantID := c.GetHeader(r.headerName)
	if tenantID == "" {
		return "", "", ErrTenantNotResolved
	}
	return tenantID, "", nil
}

func (r *Resolver) resolveFromSubdomain(c *gin.Context) (string, string, error) {
	host := c.Request.Host
	
	// Performance: Use strings.Index instead of Split for faster parsing
	firstDot := strings.IndexByte(host, '.')
	if firstDot == -1 {
		return "", "", ErrInvalidDomain
	}
	
	// Check if there's at least one more dot (domain.com)
	if strings.IndexByte(host[firstDot+1:], '.') == -1 {
		return "", "", ErrInvalidDomain
	}
	
	subdomain := host[:firstDot]
	if subdomain == "www" || subdomain == "api" {
		return "", "", ErrTenantNotResolved
	}

	return subdomain, host, nil
}

func (r *Resolver) resolveFromDomain(c *gin.Context) (string, string, error) {
	domain := c.Request.Host
	if domain == "" {
		return "", "", ErrInvalidDomain
	}

	// TODO: Lookup tenant by custom domain in database
	// This requires database access, implement in service layer
	return "", domain, ErrTenantNotResolved
}

func (r *Resolver) resolveFromParam(c *gin.Context) (string, string, error) {
	tenantID := c.Query(r.paramName)
	if tenantID == "" {
		tenantID = c.Param(r.paramName)
	}

	if tenantID == "" {
		return "", "", ErrTenantNotResolved
	}

	return tenantID, "", nil
}

// Middleware creates a Gin middleware for tenant resolution
func (r *Resolver) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, domain, err := r.Resolve(c)
		if err != nil {
			c.JSON(400, gin.H{"error": "Tenant could not be resolved"})
			c.Abort()
			return
		}

		c.Set("tenant_id", tenantID)
		c.Set("tenant_domain", domain)
		c.Next()
	}
}

// GetTenantFromContext retrieves tenant info from context
func GetTenantFromContext(ctx context.Context) (tenantID string, domain string) {
	tenantID, _ = pkgctx.GetTenantID(ctx)
	domain = pkgctx.GetTenantDomain(ctx)
	return
}
