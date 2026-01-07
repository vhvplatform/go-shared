package grpc

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GatewayWrapper provides a helper to register gRPC Gateway onto Gin router
type GatewayWrapper struct {
	Mux *runtime.ServeMux
}

// NewGatewayWrapper creates a new GatewayWrapper
func NewGatewayWrapper() *GatewayWrapper {
	return &GatewayWrapper{
		Mux: runtime.NewServeMux(
			runtime.WithIncomingHeaderMatcher(CustomHeaderMatcher),
		),
	}
}

// CustomHeaderMatcher passes through custom headers from HTTP to gRPC
func CustomHeaderMatcher(key string) (string, bool) {
	switch key {
	case "X-Tenant-Id", "X-Internal-Token", "X-Correlation-Id":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

// Handler returns a Gin HandlerFunc that proxies to the gRPC Gateway Mux
func (w *GatewayWrapper) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		w.Mux.ServeHTTP(c.Writer, c.Request)
	}
}

// DialService connects to a gRPC service with the given credentials
func DialService(ctx context.Context, target string, creds grpc.DialOption) (*grpc.ClientConn, error) {
	if creds == nil {
		creds = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	conn, err := grpc.DialContext(ctx, target, creds)
	if err != nil {
		return nil, fmt.Errorf("failed to dial service %s: %w", target, err)
	}
	return conn, nil
}
