package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// Client wraps the ClickHouse client
type Client struct {
	conn   driver.Conn
	config Config
}

// Config holds ClickHouse configuration
type Config struct {
	Addr            []string      // List of ClickHouse addresses
	Database        string        // Database name
	Username        string        // Username for authentication
	Password        string        // Password for authentication
	MaxOpenConns    int           // Maximum open connections (default: 10)
	MaxIdleConns    int           // Maximum idle connections (default: 5)
	ConnMaxLifetime time.Duration // Connection maximum lifetime (default: 1 hour)
	ConnMaxIdleTime time.Duration // Connection maximum idle time (default: 10 minutes)
	DialTimeout     time.Duration // Dial timeout (default: 10 seconds)
	Debug           bool          // Enable debug mode
	Secure          bool          // Enable TLS
	Compression     bool          // Enable compression (default: true)
}

// NewClient creates a new ClickHouse client
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	// Set default values
	if len(cfg.Addr) == 0 {
		return nil, fmt.Errorf("at least one address must be provided")
	}
	if cfg.Database == "" {
		return nil, fmt.Errorf("database name must be provided")
	}
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 10
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 5
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = 1 * time.Hour
	}
	if cfg.ConnMaxIdleTime == 0 {
		cfg.ConnMaxIdleTime = 10 * time.Minute
	}
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 10 * time.Second
	}

	// Create ClickHouse connection options
	opts := &clickhouse.Options{
		Addr: cfg.Addr,
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		DialTimeout: cfg.DialTimeout,
		Debug:       cfg.Debug,
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	}

	// Disable compression if configured
	if !cfg.Compression {
		opts.Compression = &clickhouse.Compression{
			Method: clickhouse.CompressionNone,
		}
	}

	// Open connection
	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Ping to verify connection
	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	return &Client{
		conn:   conn,
		config: cfg,
	}, nil
}

// Close closes the ClickHouse connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// HealthCheck performs a health check on the ClickHouse connection
func (c *Client) HealthCheck(ctx context.Context) error {
	return c.conn.Ping(ctx)
}

// GetConn returns the underlying ClickHouse connection for advanced operations
func (c *Client) GetConn() driver.Conn {
	return c.conn
}

// Exec executes a query without returning any rows
// Useful for INSERT, UPDATE, DELETE, CREATE, DROP, etc.
func (c *Client) Exec(ctx context.Context, query string, args ...interface{}) error {
	return c.conn.Exec(ctx, query, args...)
}

// Query executes a query that returns rows
func (c *Client) Query(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return c.conn.Select(ctx, dest, query, args...)
}

// QueryRow executes a query that is expected to return at most one row
func (c *Client) QueryRow(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return c.conn.QueryRow(ctx, query, args...).Scan(dest)
}

// PrepareBatch prepares a batch for efficient bulk inserts
func (c *Client) PrepareBatch(ctx context.Context, query string) (driver.Batch, error) {
	return c.conn.PrepareBatch(ctx, query)
}

// Stats returns database statistics
func (c *Client) Stats() driver.Stats {
	return c.conn.Stats()
}

// GetDatabase returns the configured database name
func (c *Client) GetDatabase() string {
	return c.config.Database
}
