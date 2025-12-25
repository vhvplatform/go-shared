package storage

import (
	"fmt"
)

// Placeholder implementations for different providers
// TODO: Implement actual provider logic

func newS3Client(config Config) (Client, error) {
	return nil, fmt.Errorf("S3 client not yet implemented")
}

func newMinIOClient(config Config) (Client, error) {
	return nil, fmt.Errorf("MinIO client not yet implemented")
}

func newGCSClient(config Config) (Client, error) {
	return nil, fmt.Errorf("GCS client not yet implemented")
}

func newAzureBlobClient(config Config) (Client, error) {
	return nil, fmt.Errorf("Azure Blob client not yet implemented")
}

func newLocalClient(config Config) (Client, error) {
	return nil, fmt.Errorf("Local storage client not yet implemented")
}
