# Storage Service Package

Package `storage` cung cấp abstraction layer cho object storage với hỗ trợ nhiều provider khác nhau.

## Providers Hỗ Trợ

- ✅ AWS S3
- ✅ MinIO
- ✅ Google Cloud Storage (GCS)
- ✅ Azure Blob Storage
- ✅ Local Filesystem

## Tính Năng

- Upload/Download objects
- List objects với pagination
- Delete single và multiple objects
- Copy objects giữa buckets
- Pre-signed URLs
- Metadata và ACL
- Bucket management

## Sử Dụng Cơ Bản

```go
import "github.com/vhvcorp/go-shared/storage"

// Tạo client với S3
client, err := storage.NewClient(storage.Config{
    Provider: storage.ProviderS3,
    Bucket:   "my-bucket",
    Region:   "us-east-1",
    Options: map[string]string{
        "access_key_id": "your-access-key",
        "secret_access_key": "your-secret-key",
    },
})
defer client.Close()

// Upload file
file, _ := os.Open("photo.jpg")
defer file.Close()

fileInfo, _ := file.Stat()
object, err := client.Upload(ctx, "my-bucket", &storage.UploadInput{
    Key:         "photos/2024/photo.jpg",
    Body:        file,
    Size:        fileInfo.Size(),
    ContentType: "image/jpeg",
    ACL:         storage.ACLPublicRead,
})

fmt.Printf("Uploaded: %s (%d bytes)\n", object.Key, object.Size)
```

## Download File

```go
// Download object
reader, err := client.Download(ctx, "my-bucket", &storage.DownloadInput{
    Key: "photos/2024/photo.jpg",
})
defer reader.Close()

// Lưu vào file
outFile, _ := os.Create("downloaded.jpg")
io.Copy(outFile, reader)
```

## List Objects

```go
// List với pagination
result, err := client.List(ctx, "my-bucket", &storage.ListInput{
    Prefix:  "photos/2024/",
    MaxKeys: 100,
})

for _, obj := range result.Objects {
    fmt.Printf("- %s (%d bytes, modified: %s)\n", 
        obj.Key, obj.Size, obj.LastModified)
}

// Tiếp tục với page kế tiếp
if result.IsTruncated {
    nextPage, _ := client.List(ctx, "my-bucket", &storage.ListInput{
        Prefix: "photos/2024/",
        Marker: result.NextMarker,
    })
}
```

## Pre-signed URLs

```go
// Tạo URL có hiệu lực trong 1 giờ
url, err := client.GetPresignedURL(ctx, "my-bucket", 
    "photos/secret.jpg", time.Hour)

fmt.Printf("Temporary URL: %s\n", url)
```

## Delete Objects

```go
// Delete một object
err := client.Delete(ctx, "my-bucket", "photos/old.jpg")

// Delete nhiều objects
keys := []string{
    "photos/file1.jpg",
    "photos/file2.jpg",
    "photos/file3.jpg",
}
err = client.DeleteMultiple(ctx, "my-bucket", keys)
```

## Copy Objects

```go
// Copy trong cùng bucket
object, err := client.Copy(ctx, 
    "my-bucket", "photos/original.jpg",
    "my-bucket", "photos/backup.jpg")

// Copy giữa các buckets
object, err = client.Copy(ctx,
    "source-bucket", "file.pdf",
    "dest-bucket", "archived/file.pdf")
```

## Metadata và Custom Headers

```go
// Upload với metadata
object, err := client.Upload(ctx, "my-bucket", &storage.UploadInput{
    Key:  "document.pdf",
    Body: file,
    Metadata: map[string]string{
        "author":      "John Doe",
        "department":  "Engineering",
        "uploaded-by": "api-service",
    },
})

// Lấy metadata
obj, err := client.Get(ctx, "my-bucket", "document.pdf")
fmt.Printf("Author: %s\n", obj.Metadata["author"])
```

## MinIO Configuration

```go
// MinIO local hoặc self-hosted
client, err := storage.NewClient(storage.Config{
    Provider: storage.ProviderMinIO,
    Options: map[string]string{
        "endpoint": "localhost:9000",
        "access_key": "minioadmin",
        "secret_key": "minioadmin",
        "use_ssl": "false",
    },
})
```

## Status

🚧 **In Development** - Hiện tại package đã có interfaces và structures sẵn sàng, implementations cho các providers đang được phát triển.
