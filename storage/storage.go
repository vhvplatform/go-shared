// Package storage provides an abstraction layer for object storage
// with support for multiple providers (S3, MinIO, GCS, Azure Blob, etc.)
package storage

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Provider represents a storage service provider
type Provider string

const (
	// ProviderS3 uses AWS S3
	ProviderS3 Provider = "s3"
	// ProviderMinIO uses MinIO
	ProviderMinIO Provider = "minio"
	// ProviderGCS uses Google Cloud Storage
	ProviderGCS Provider = "gcs"
	// ProviderAzureBlob uses Azure Blob Storage
	ProviderAzureBlob Provider = "azure_blob"
	// ProviderLocal uses local filesystem
	ProviderLocal Provider = "local"
)

// Object represents a stored object
type Object struct {
	Key          string            // Object key/path
	Size         int64             // Object size in bytes
	ContentType  string            // MIME type
	LastModified time.Time         // Last modification time
	ETag         string            // Entity tag
	Metadata     map[string]string // Custom metadata
	URL          string            // Public URL (if available)
}

// UploadInput contains parameters for uploading an object
type UploadInput struct {
	Key         string            // Object key/path
	Body        io.Reader         // Object content
	Size        int64             // Content size (optional, but recommended)
	ContentType string            // MIME type
	ACL         ACL               // Access control list
	Metadata    map[string]string // Custom metadata
}

// DownloadInput contains parameters for downloading an object
type DownloadInput struct {
	Key   string // Object key/path
	Range string // Byte range (optional, e.g., "bytes=0-1023")
}

// ListInput contains parameters for listing objects
type ListInput struct {
	Prefix    string // Filter by key prefix
	Delimiter string // Delimiter for grouping
	MaxKeys   int    // Maximum number of keys to return
	Marker    string // Pagination marker
}

// ListOutput contains the result of a list operation
type ListOutput struct {
	Objects     []Object // List of objects
	Prefixes    []string // Common prefixes (directories)
	NextMarker  string   // Marker for next page
	IsTruncated bool     // Whether there are more results
	TotalCount  int      // Total number of objects (if available)
}

// ACL represents access control list
type ACL string

const (
	// ACLPrivate makes object private
	ACLPrivate ACL = "private"
	// ACLPublicRead makes object publicly readable
	ACLPublicRead ACL = "public-read"
	// ACLPublicReadWrite makes object publicly readable and writable
	ACLPublicReadWrite ACL = "public-read-write"
	// ACLAuthenticatedRead makes object readable by authenticated users
	ACLAuthenticatedRead ACL = "authenticated-read"
)

// Client is the interface that all storage providers must implement
type Client interface {
	// Upload uploads an object to storage
	Upload(ctx context.Context, bucket string, input *UploadInput) (*Object, error)

	// Download downloads an object from storage
	Download(ctx context.Context, bucket string, input *DownloadInput) (io.ReadCloser, error)

	// Delete deletes an object from storage
	Delete(ctx context.Context, bucket, key string) error

	// DeleteMultiple deletes multiple objects from storage
	DeleteMultiple(ctx context.Context, bucket string, keys []string) error

	// Get retrieves object metadata without downloading content
	Get(ctx context.Context, bucket, key string) (*Object, error)

	// List lists objects in a bucket
	List(ctx context.Context, bucket string, input *ListInput) (*ListOutput, error)

	// Exists checks if an object exists
	Exists(ctx context.Context, bucket, key string) (bool, error)

	// Copy copies an object within or between buckets
	Copy(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey string) (*Object, error)

	// GetPresignedURL generates a pre-signed URL for temporary access
	GetPresignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)

	// CreateBucket creates a new bucket
	CreateBucket(ctx context.Context, bucket string) error

	// DeleteBucket deletes a bucket
	DeleteBucket(ctx context.Context, bucket string) error

	// BucketExists checks if a bucket exists
	BucketExists(ctx context.Context, bucket string) (bool, error)

	// Close closes the storage client and releases resources
	Close() error
}

// Config contains configuration for storage client
// Note: Use provider-specific config structs (S3Config, MinIOConfig, etc.)
// for type-safe configuration, or use Options map for dynamic configuration
type Config struct {
	Provider Provider          // Storage provider to use
	Bucket   string            // Default bucket name
	Region   string            // Region (for cloud providers)
	Options  map[string]string // Provider-specific options
}

// S3Config contains AWS S3-specific configuration
type S3Config struct {
	Region          string // AWS region
	AccessKeyID     string // AWS access key ID
	SecretAccessKey string // AWS secret access key
	Endpoint        string // Custom endpoint (optional, for S3-compatible services)
	UsePathStyle    bool   // Use path-style addressing
	DisableSSL      bool   // Disable SSL (for local dev only)
}

// MinIOConfig contains MinIO-specific configuration
type MinIOConfig struct {
	Endpoint        string // MinIO endpoint
	AccessKeyID     string // MinIO access key
	SecretAccessKey string // MinIO secret key
	UseSSL          bool   // Use SSL
}

// GCSConfig contains Google Cloud Storage-specific configuration
type GCSConfig struct {
	ProjectID       string // GCP project ID
	CredentialsFile string // Path to credentials JSON file
	CredentialsJSON []byte // Credentials JSON content
}

// AzureBlobConfig contains Azure Blob Storage-specific configuration
type AzureBlobConfig struct {
	AccountName   string // Storage account name
	AccountKey    string // Storage account key
	ContainerName string // Default container name
}

// LocalConfig contains local filesystem-specific configuration
type LocalConfig struct {
	BasePath string // Base directory path
}

// NewClient creates a new storage client based on the provider
func NewClient(config Config) (Client, error) {
	switch config.Provider {
	case ProviderS3:
		return newS3Client(config)
	case ProviderMinIO:
		return newMinIOClient(config)
	case ProviderGCS:
		return newGCSClient(config)
	case ProviderAzureBlob:
		return newAzureBlobClient(config)
	case ProviderLocal:
		return newLocalClient(config)
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", config.Provider)
	}
}

// Validate checks if upload input is valid
func (u *UploadInput) Validate() error {
	if u.Key == "" {
		return fmt.Errorf("key is required")
	}
	if u.Body == nil {
		return fmt.Errorf("body is required")
	}
	return nil
}

// Validate checks if download input is valid
func (d *DownloadInput) Validate() error {
	if d.Key == "" {
		return fmt.Errorf("key is required")
	}
	return nil
}
