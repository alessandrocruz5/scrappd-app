# R2 Storage Setup Guide

This guide explains how to set up and use Cloudflare R2 storage in the Scrappd application.

## Overview

The application uses a storage interface that supports Cloudflare R2 (S3-compatible) object storage. This allows you to store user-uploaded images, processed images, and other files in a scalable, cost-effective manner.

## Architecture

### Storage Interface

The `Storage` interface (`internal/services/storage.go`) defines the following operations:

- **Upload**: Upload a file with auto-generated key
- **UploadWithKey**: Upload a file with a specific key/path
- **Download**: Retrieve a file from storage
- **Delete**: Remove a file from storage
- **GetURL**: Generate a presigned URL for temporary access
- **Exists**: Check if a file exists
- **List**: List files with a given prefix

### R2Storage Implementation

The `R2Storage` struct implements the `Storage` interface using AWS SDK v2 for S3-compatible operations. It automatically handles:

- Authentication with R2 endpoints
- File organization by date (uploads/YYYY/MM/DD/)
- Unique file naming using UUIDs
- Comprehensive logging
- Error handling

## Setup Instructions

### 1. Create Cloudflare R2 Bucket

1. Log in to your [Cloudflare Dashboard](https://dash.cloudflare.com)
2. Navigate to **R2 Object Storage**
3. Click **Create bucket**
4. Enter bucket name (e.g., `scrappd-images`)
5. Choose your preferred location
6. Click **Create bucket**

### 2. Generate R2 API Tokens

1. In the R2 dashboard, click **Manage R2 API Tokens**
2. Click **Create API token**
3. Configure permissions:
   - **Object Read & Write** for full access
   - Or specific permissions based on your needs
4. Click **Create API Token**
5. Copy the following values:
   - Access Key ID
   - Secret Access Key
   - Endpoint URL (format: `https://<account-id>.r2.cloudflarestorage.com`)

### 3. Configure Environment Variables

Update your `.env` file with the R2 credentials:

```env
# Storage Configuration
STORAGE_ENDPOINT=https://<account-id>.r2.cloudflarestorage.com
STORAGE_ACCESS_KEY_ID=<your-access-key-id>
STORAGE_SECRET_ACCESS_KEY=<your-secret-access-key>
STORAGE_BUCKET_NAME=scrappd-images
STORAGE_REGION=auto
```

**Note**: Never commit credentials to version control. Use environment variables or secret management systems.

### 4. Install Dependencies

The required AWS SDK dependencies are already added to `go.mod`:

```bash
go mod tidy
go mod download
```

Dependencies:
- `github.com/aws/aws-sdk-go-v2`
- `github.com/aws/aws-sdk-go-v2/config`
- `github.com/aws/aws-sdk-go-v2/credentials`
- `github.com/aws/aws-sdk-go-v2/service/s3`

### 5. Initialize Storage in Your Application

In your main application setup (e.g., `cmd/api/main.go`):

```go
import (
    "github.com/alessandrocruz5/scrappd-app/backend/internal/config"
    "github.com/alessandrocruz5/scrappd-app/backend/internal/services"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    // Initialize logger
    logger := logrus.New()

    // Initialize storage
    storage, err := services.NewR2Storage(&cfg.Storage, logger)
    if err != nil {
        log.Fatalf("Failed to initialize storage: %v", err)
    }

    // Pass storage to your handlers
    // handler := handlers.NewHandler(storage)
}
```

## Usage Examples

### Upload a File

```go
func (h *Handler) uploadFile(c *gin.Context) {
    file, header, err := c.Request.FormFile("file")
    if err != nil {
        c.JSON(400, gin.H{"error": "No file provided"})
        return
    }
    defer file.Close()

    key, err := h.storage.Upload(
        c.Request.Context(),
        file,
        header.Filename,
        header.Header.Get("Content-Type"),
    )
    if err != nil {
        c.JSON(500, gin.H{"error": "Upload failed"})
        return
    }

    c.JSON(200, gin.H{"key": key})
}
```

### Generate Presigned URL

```go
// Generate a URL valid for 1 hour
url, err := storage.GetURL(ctx, fileKey, 1*time.Hour)
if err != nil {
    return err
}
// Share this URL with clients
```

### Delete a File

```go
err := storage.Delete(ctx, fileKey)
if err != nil {
    return err
}
```

### List User Files

```go
files, err := storage.List(ctx, "uploads/user-123/")
if err != nil {
    return err
}
```

## File Organization

Files are automatically organized using the following structure:

```
bucket/
└── uploads/
    └── YYYY/
        └── MM/
            └── DD/
                ├── <uuid1>.jpg
                ├── <uuid2>.png
                └── <uuid3>.webp
```

This organization:
- Makes it easy to implement retention policies
- Allows efficient listing by date
- Prevents naming conflicts with UUIDs
- Enables date-based analytics

## Security Considerations

### 1. Access Control

- Use separate buckets for different environments (dev, staging, prod)
- Implement role-based access control in your application
- Never expose storage keys directly to clients

### 2. Presigned URLs

- Set appropriate expiration times (default: 1 hour)
- Consider implementing URL signing in your application
- Monitor presigned URL generation for abuse

### 3. File Validation

Always validate files before uploading:

```go
import "github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"

// Validate image
err := utils.ValidateImageFile(fileHeader)
if err != nil {
    return err
}
```

### 4. Rate Limiting

Implement rate limiting on upload endpoints to prevent abuse:

```go
// Example using middleware
router.POST("/upload", rateLimitMiddleware, uploadHandler)
```

## Cost Optimization

### R2 Pricing (as of 2024)
- **Storage**: $0.015 per GB/month
- **Class A operations** (write, list): $4.50 per million requests
- **Class B operations** (read): $0.36 per million requests
- **Data egress**: Free (no egress charges!)

### Best Practices

1. **Lifecycle Policies**: Implement automatic deletion of temporary files
2. **Compression**: Compress images before uploading when possible
3. **Caching**: Use CDN or edge caching for frequently accessed files
4. **Monitoring**: Track storage usage and request patterns

## Monitoring and Logging

The R2Storage implementation includes comprehensive logging:

```go
logger.WithFields(logrus.Fields{
    "key": key,
    "content_type": contentType,
}).Info("Successfully uploaded file to R2")
```

Monitor these logs for:
- Upload/download patterns
- Error rates
- Performance metrics
- Security events

## Troubleshooting

### Common Issues

#### 1. Authentication Errors

```
Error: failed to upload file: AccessDenied
```

**Solution**: Verify your access key ID and secret access key are correct.

#### 2. Endpoint Connection Errors

```
Error: failed to connect to endpoint
```

**Solution**: Check that your endpoint URL is in the correct format:
`https://<account-id>.r2.cloudflarestorage.com`

#### 3. Bucket Not Found

```
Error: NoSuchBucket: The specified bucket does not exist
```

**Solution**: Verify the bucket name matches exactly (case-sensitive).

#### 4. Region Errors

```
Error: region mismatch
```

**Solution**: For R2, always use `STORAGE_REGION=auto`.

## Testing

### Unit Tests

Create unit tests using mocks:

```go
type MockStorage struct {
    mock.Mock
}

func (m *MockStorage) Upload(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
    args := m.Called(ctx, file, filename, contentType)
    return args.String(0), args.Error(1)
}
```

### Integration Tests

Test with actual R2 bucket (use separate test bucket):

```bash
STORAGE_BUCKET_NAME=scrappd-images-test go test ./internal/services/...
```

## Migration from Other Storage

If migrating from another storage solution:

1. Implement the `Storage` interface for your current provider
2. Run both implementations in parallel
3. Gradually migrate files to R2
4. Update references to use new keys
5. Deprecate old storage

## Additional Resources

- [Cloudflare R2 Documentation](https://developers.cloudflare.com/r2/)
- [AWS SDK for Go v2 Documentation](https://aws.github.io/aws-sdk-go-v2/docs/)
- [S3 API Reference](https://docs.aws.amazon.com/AmazonS3/latest/API/Welcome.html)

## Support

For issues or questions:
1. Check the troubleshooting section above
2. Review application logs
3. Consult Cloudflare R2 documentation
4. Open an issue in the project repository
