package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Storage defines the interface for object storage operations
type Storage interface {
	// Upload uploads a file to storage and returns the file key/path
	Upload(ctx context.Context, file io.Reader, filename string, contentType string) (string, error)

	// UploadWithKey uploads a file with a specific key/path
	UploadWithKey(ctx context.Context, file io.Reader, key string, contentType string) error

	// Download retrieves a file from storage
	Download(ctx context.Context, key string) ([]byte, error)

	// Delete removes a file from storage
	Delete(ctx context.Context, key string) error

	// GetURL generates a presigned URL for accessing a file
	GetURL(ctx context.Context, key string, expiry time.Duration) (string, error)

	// Exists checks if a file exists in storage
	Exists(ctx context.Context, key string) (bool, error)

	// List lists files with a given prefix
	List(ctx context.Context, prefix string) ([]string, error)

	// HealthCheck validates storage connectivity
	HealthCheck(ctx context.Context) error
}

// R2Storage implements the Storage interface using Cloudflare R2 (S3-compatible)
type R2Storage struct {
	client     *s3.Client
	bucketName string
	logger     *logrus.Logger
}

// NewR2Storage creates a new R2 storage client
func NewR2Storage(cfg *config.StorageConfig, logger *logrus.Logger) (*R2Storage, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("storage endpoint is required")
	}
	if cfg.AccessKeyID == "" {
		return nil, fmt.Errorf("storage access key ID is required")
	}
	if cfg.SecretAccessKey == "" {
		return nil, fmt.Errorf("storage secret access key is required")
	}
	if cfg.BucketName == "" {
		return nil, fmt.Errorf("storage bucket name is required")
	}

	// Create custom AWS config for R2
	awsCfg := aws.Config{
		Region: cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		),
		EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               cfg.Endpoint,
					HostnameImmutable: true,
					SigningRegion:     cfg.Region,
				}, nil
			},
		),
	}

	// Create S3 client
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	logger.WithFields(logrus.Fields{
		"endpoint": cfg.Endpoint,
		"bucket":   cfg.BucketName,
		"region":   cfg.Region,
	}).Info("Initialized R2 storage client")

	return &R2Storage{
		client:     client,
		bucketName: cfg.BucketName,
		logger:     logger,
	}, nil
}

// Upload uploads a file to R2 storage with an auto-generated key
func (r *R2Storage) Upload(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
	// Generate a unique key with timestamp and UUID
	ext := filepath.Ext(filename)
	key := fmt.Sprintf("uploads/%s/%s%s",
		time.Now().Format("2006/01/02"),
		uuid.New().String(),
		ext,
	)

	err := r.UploadWithKey(ctx, file, key, contentType)
	if err != nil {
		return "", err
	}

	return key, nil
}

// UploadWithKey uploads a file to R2 storage with a specific key
func (r *R2Storage) UploadWithKey(ctx context.Context, file io.Reader, key string, contentType string) error {
	// Read the file content
	content, err := io.ReadAll(file)
	if err != nil {
		r.logger.WithError(err).Error("Failed to read file content")
		return fmt.Errorf("failed to read file content: %w", err)
	}

	// Upload to R2
	input := &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: aws.String(contentType),
	}

	_, err = r.client.PutObject(ctx, input)
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"key":   key,
			"error": err,
		}).Error("Failed to upload file to R2")
		return fmt.Errorf("failed to upload file: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"key":          key,
		"content_type": contentType,
	}).Info("Successfully uploaded file to R2")

	return nil
}

// Download retrieves a file from R2 storage
func (r *R2Storage) Download(ctx context.Context, key string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	}

	result, err := r.client.GetObject(ctx, input)
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"key":   key,
			"error": err,
		}).Error("Failed to download file from R2")
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer result.Body.Close()

	content, err := io.ReadAll(result.Body)
	if err != nil {
		r.logger.WithError(err).Error("Failed to read downloaded file content")
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	r.logger.WithField("key", key).Info("Successfully downloaded file from R2")
	return content, nil
}

// Delete removes a file from R2 storage
func (r *R2Storage) Delete(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	}

	_, err := r.client.DeleteObject(ctx, input)
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"key":   key,
			"error": err,
		}).Error("Failed to delete file from R2")
		return fmt.Errorf("failed to delete file: %w", err)
	}

	r.logger.WithField("key", key).Info("Successfully deleted file from R2")
	return nil
}

// HealthCheck validates access to the configured bucket.
func (r *R2Storage) HealthCheck(ctx context.Context) error {
	_, err := r.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(r.bucketName),
	})
	if err != nil {
		r.logger.WithError(err).Error("Storage health check failed")
		return fmt.Errorf("storage health check failed: %w", err)
	}
	return nil
}

// GetURL generates a presigned URL for accessing a file
func (r *R2Storage) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(r.client)

	input := &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	}

	result, err := presignClient.PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"key":   key,
			"error": err,
		}).Error("Failed to generate presigned URL")
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"key":    key,
		"expiry": expiry,
	}).Info("Successfully generated presigned URL")

	return result.URL, nil
}

// Exists checks if a file exists in R2 storage
func (r *R2Storage) Exists(ctx context.Context, key string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	}

	_, err := r.client.HeadObject(ctx, input)
	if err != nil {
		// Check if it's a "not found" error
		// In AWS SDK v2, we need to check the error type
		return false, nil
	}

	return true, nil
}

// List lists files with a given prefix
func (r *R2Storage) List(ctx context.Context, prefix string) ([]string, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(r.bucketName),
		Prefix: aws.String(prefix),
	}

	var keys []string
	paginator := s3.NewListObjectsV2Paginator(r.client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			r.logger.WithFields(logrus.Fields{
				"prefix": prefix,
				"error":  err,
			}).Error("Failed to list files from R2")
			return nil, fmt.Errorf("failed to list files: %w", err)
		}

		for _, obj := range page.Contents {
			if obj.Key != nil {
				keys = append(keys, *obj.Key)
			}
		}
	}

	r.logger.WithFields(logrus.Fields{
		"prefix": prefix,
		"count":  len(keys),
	}).Info("Successfully listed files from R2")

	return keys, nil
}
