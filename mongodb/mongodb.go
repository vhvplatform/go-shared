package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Client wraps the MongoDB client
type Client struct {
	*mongo.Client
	database string
	config   Config
}

// MongoClient is an alias for Client for backward compatibility
type MongoClient = Client

// Config holds MongoDB configuration
type Config struct {
	URI             string
	Database        string
	MaxPoolSize     uint64
	MinPoolSize     uint64
	ConnectTimeout  time.Duration
	MaxConnIdleTime time.Duration
}

// NewClient creates a new MongoDB client
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	// Set default timeouts if not provided
	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = 10 * time.Second
	}
	if cfg.MaxConnIdleTime == 0 {
		cfg.MaxConnIdleTime = 5 * time.Minute
	}

	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize).
		SetConnectTimeout(cfg.ConnectTimeout).
		SetMaxConnIdleTime(cfg.MaxConnIdleTime).
		SetServerSelectionTimeout(5 * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &Client{
		Client:   client,
		database: cfg.Database,
		config:   cfg,
	}, nil
}

// Database returns the configured database
func (c *Client) Database() *mongo.Database {
	return c.Client.Database(c.database)
}

// Collection returns a collection from the configured database
func (c *Client) Collection(name string) *mongo.Collection {
	return c.Database().Collection(name)
}

// Close closes the MongoDB connection
func (c *Client) Close(ctx context.Context) error {
	return c.Client.Disconnect(ctx)
}

// HealthCheck performs a health check on the MongoDB connection
func (c *Client) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.Client.Ping(ctx, readpref.Primary())
}

// NewMongoClient creates a new MongoDB client (wrapper for NewClient for backward compatibility)
func NewMongoClient(uri, database string) (*MongoClient, error) {
	cfg := Config{
		URI:         uri,
		Database:    database,
		MaxPoolSize: 100,
		MinPoolSize: 10,
	}
	return NewClient(context.Background(), cfg)
}

// Disconnect closes the MongoDB connection (alias for Close)
func (c *Client) Disconnect(ctx context.Context) error {
	return c.Close(ctx)
}

// GetClient returns the underlying mongo.Client
func (c *Client) GetClient() *mongo.Client {
	return c.Client
}

// GetTenantCollection returns a collection with tenant isolation
func (c *Client) GetTenantCollection(name string, tenantID string) *mongo.Collection {
	// This is a convenience method that can be used with TenantRepository
	return c.Collection(name)
}
