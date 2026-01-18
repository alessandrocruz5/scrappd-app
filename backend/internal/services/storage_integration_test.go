// +build integration

package services

import (
	"bytes"
	"context"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/config"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests require actual R2 credentials
// Run with: go test -tags=integration ./internal/services/...
//
// Required environment variables:
// - STORAGE_ENDPOINT
// - STORAGE_ACCESS_KEY_ID
// - STORAGE_SECRET_ACCESS_KEY
// - STORAGE_BUCKET_NAME
// - STORAGE_REGION (optional, defaults to "auto")

func loadTestConfig(t *testing.T) *config.StorageConfig {
	// Try to load .env file (optional)
	_ = godotenv.Load("../../.env")

	cfg := &config.StorageConfig{
		Endpoint:        os.Getenv("STORAGE_ENDPOINT"),
		AccessKeyID:     os.Getenv("STORAGE_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("STORAGE_SECRET_ACCESS_KEY"),
		BucketName:      os.Getenv("STORAGE_BUCKET_NAME"),
		Region:          os.Getenv("STORAGE_REGION"),
	}

	// Default region to "auto" if not set
	if cfg.Region == "" {
		cfg.Region = "auto"
	}

	// Check if all required variables are set
	if cfg.Endpoint == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" || cfg.BucketName == "" {
		t.Skip("Skipping integration tests: R2 credentials not configured. Set STORAGE_ENDPOINT, STORAGE_ACCESS_KEY_ID, STORAGE_SECRET_ACCESS_KEY, and STORAGE_BUCKET_NAME environment variables.")
	}

	return cfg
}

func TestIntegrationR2StorageUpload(t *testing.T) {
	cfg := loadTestConfig(t)
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	storage, err := NewR2Storage(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, storage)

	ctx := context.Background()

	t.Run("upload small file", func(t *testing.T) {
		content := []byte("Hello from integration test!")
		file := bytes.NewReader(content)

		key, err := storage.Upload(ctx, file, "test.txt", "text/plain")
		require.NoError(t, err)
		assert.NotEmpty(t, key)

		// Verify key format: uploads/YYYY/MM/DD/UUID.ext
		pattern := `^uploads/\d{4}/\d{2}/\d{2}/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\.txt$`
		matched, _ := regexp.MatchString(pattern, key)
		assert.True(t, matched, "Key should match pattern: %s", pattern)

		// Cleanup
		defer func() {
			_ = storage.Delete(ctx, key)
		}()
	})

	t.Run("upload image file", func(t *testing.T) {
		// Create a simple 1x1 PNG image (smallest valid PNG)
		pngData := []byte{
			0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
			0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
			0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
			0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
			0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41,
			0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
			0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
			0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
			0x42, 0x60, 0x82,
		}

		file := bytes.NewReader(pngData)

		key, err := storage.Upload(ctx, file, "test.png", "image/png")
		require.NoError(t, err)
		assert.NotEmpty(t, key)
		assert.Contains(t, key, ".png")

		// Cleanup
		defer func() {
			_ = storage.Delete(ctx, key)
		}()
	})

	t.Run("upload with custom key", func(t *testing.T) {
		content := []byte("Custom key test")
		file := bytes.NewReader(content)
		customKey := "test/custom/path/file.txt"

		err := storage.UploadWithKey(ctx, file, customKey, "text/plain")
		require.NoError(t, err)

		// Cleanup
		defer func() {
			_ = storage.Delete(ctx, customKey)
		}()
	})
}

func TestIntegrationR2StorageDownload(t *testing.T) {
	cfg := loadTestConfig(t)
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	storage, err := NewR2Storage(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("download uploaded file", func(t *testing.T) {
		originalContent := []byte("Download test content")
		file := bytes.NewReader(originalContent)

		// Upload
		key, err := storage.Upload(ctx, file, "download-test.txt", "text/plain")
		require.NoError(t, err)

		// Download
		downloadedContent, err := storage.Download(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, originalContent, downloadedContent)

		// Cleanup
		defer func() {
			_ = storage.Delete(ctx, key)
		}()
	})

	t.Run("download non-existent file", func(t *testing.T) {
		_, err := storage.Download(ctx, "non/existent/file.txt")
		assert.Error(t, err)
	})
}

func TestIntegrationR2StorageDelete(t *testing.T) {
	cfg := loadTestConfig(t)
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	storage, err := NewR2Storage(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("delete existing file", func(t *testing.T) {
		content := []byte("Delete test")
		file := bytes.NewReader(content)

		// Upload
		key, err := storage.Upload(ctx, file, "delete-test.txt", "text/plain")
		require.NoError(t, err)

		// Verify exists
		exists, err := storage.Exists(ctx, key)
		require.NoError(t, err)
		assert.True(t, exists)

		// Delete
		err = storage.Delete(ctx, key)
		require.NoError(t, err)

		// Verify deleted (may still exist briefly due to eventual consistency)
		// We don't assert here as R2 might take a moment to propagate deletion
	})

	t.Run("delete non-existent file", func(t *testing.T) {
		// Deleting non-existent file should not error in S3
		err := storage.Delete(ctx, "non/existent/file.txt")
		// S3 returns success even if file doesn't exist
		assert.NoError(t, err)
	})
}

func TestIntegrationR2StorageGetURL(t *testing.T) {
	cfg := loadTestConfig(t)
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	storage, err := NewR2Storage(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("generate presigned URL", func(t *testing.T) {
		content := []byte("Presigned URL test")
		file := bytes.NewReader(content)

		// Upload
		key, err := storage.Upload(ctx, file, "presigned-test.txt", "text/plain")
		require.NoError(t, err)

		// Generate presigned URL
		url, err := storage.GetURL(ctx, key, 1*time.Hour)
		require.NoError(t, err)
		assert.NotEmpty(t, url)
		assert.Contains(t, url, "https://")

		// Cleanup
		defer func() {
			_ = storage.Delete(ctx, key)
		}()
	})

	t.Run("generate URL with different expiry", func(t *testing.T) {
		content := []byte("Expiry test")
		file := bytes.NewReader(content)

		key, err := storage.Upload(ctx, file, "expiry-test.txt", "text/plain")
		require.NoError(t, err)

		// Test different expiry durations
		durations := []time.Duration{
			15 * time.Minute,
			1 * time.Hour,
			24 * time.Hour,
		}

		for _, duration := range durations {
			url, err := storage.GetURL(ctx, key, duration)
			require.NoError(t, err)
			assert.NotEmpty(t, url)
		}

		// Cleanup
		defer func() {
			_ = storage.Delete(ctx, key)
		}()
	})
}

func TestIntegrationR2StorageExists(t *testing.T) {
	cfg := loadTestConfig(t)
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	storage, err := NewR2Storage(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("check existing file", func(t *testing.T) {
		content := []byte("Exists test")
		file := bytes.NewReader(content)

		key, err := storage.Upload(ctx, file, "exists-test.txt", "text/plain")
		require.NoError(t, err)

		exists, err := storage.Exists(ctx, key)
		require.NoError(t, err)
		assert.True(t, exists)

		// Cleanup
		defer func() {
			_ = storage.Delete(ctx, key)
		}()
	})

	t.Run("check non-existent file", func(t *testing.T) {
		exists, err := storage.Exists(ctx, "non/existent/file.txt")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestIntegrationR2StorageList(t *testing.T) {
	cfg := loadTestConfig(t)
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	storage, err := NewR2Storage(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("list files with prefix", func(t *testing.T) {
		// Upload multiple files with same prefix
		prefix := "test-list/"
		filenames := []string{"file1.txt", "file2.txt", "file3.txt"}
		uploadedKeys := make([]string, 0, len(filenames))

		for _, filename := range filenames {
			content := []byte("List test: " + filename)
			file := bytes.NewReader(content)

			key := prefix + filename
			err := storage.UploadWithKey(ctx, file, key, "text/plain")
			require.NoError(t, err)
			uploadedKeys = append(uploadedKeys, key)
		}

		// List files
		keys, err := storage.List(ctx, prefix)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(keys), len(filenames), "Should find at least the uploaded files")

		// Verify our files are in the list
		for _, uploadedKey := range uploadedKeys {
			assert.Contains(t, keys, uploadedKey)
		}

		// Cleanup
		defer func() {
			for _, key := range uploadedKeys {
				_ = storage.Delete(ctx, key)
			}
		}()
	})

	t.Run("list empty prefix", func(t *testing.T) {
		keys, err := storage.List(ctx, "non-existent-prefix-12345/")
		require.NoError(t, err)
		assert.Empty(t, keys)
	})
}

func TestIntegrationR2StorageCompleteWorkflow(t *testing.T) {
	cfg := loadTestConfig(t)
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	storage, err := NewR2Storage(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("complete upload-download-delete workflow", func(t *testing.T) {
		originalContent := []byte("Complete workflow test content")

		// 1. Upload
		file := bytes.NewReader(originalContent)
		key, err := storage.Upload(ctx, file, "workflow-test.txt", "text/plain")
		require.NoError(t, err)
		assert.NotEmpty(t, key)

		// 2. Verify exists
		exists, err := storage.Exists(ctx, key)
		require.NoError(t, err)
		assert.True(t, exists)

		// 3. Download and verify content
		downloadedContent, err := storage.Download(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, originalContent, downloadedContent)

		// 4. Generate presigned URL
		url, err := storage.GetURL(ctx, key, 1*time.Hour)
		require.NoError(t, err)
		assert.NotEmpty(t, url)

		// 5. Delete
		err = storage.Delete(ctx, key)
		require.NoError(t, err)

		// 6. Verify deletion (note: may take time due to eventual consistency)
		// We skip this check as it might fail due to R2's eventual consistency
	})
}

func TestIntegrationR2StorageLargeFile(t *testing.T) {
	cfg := loadTestConfig(t)
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	storage, err := NewR2Storage(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("upload and download 1MB file", func(t *testing.T) {
		// Create 1MB of data
		size := 1024 * 1024 // 1MB
		content := make([]byte, size)
		for i := range content {
			content[i] = byte(i % 256)
		}

		file := bytes.NewReader(content)

		// Upload
		key, err := storage.Upload(ctx, file, "large-file.bin", "application/octet-stream")
		require.NoError(t, err)

		// Download
		downloadedContent, err := storage.Download(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, len(content), len(downloadedContent))
		assert.Equal(t, content, downloadedContent)

		// Cleanup
		defer func() {
			_ = storage.Delete(ctx, key)
		}()
	})
}

func TestIntegrationR2StorageConcurrency(t *testing.T) {
	cfg := loadTestConfig(t)
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	storage, err := NewR2Storage(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("concurrent uploads", func(t *testing.T) {
		numUploads := 5
		uploadedKeys := make(chan string, numUploads)
		errors := make(chan error, numUploads)

		// Upload files concurrently
		for i := 0; i < numUploads; i++ {
			go func(index int) {
				content := []byte(string(rune('A' + index)))
				file := bytes.NewReader(content)

				key, err := storage.Upload(ctx, file, string(rune('A'+index))+".txt", "text/plain")
				if err != nil {
					errors <- err
					return
				}
				uploadedKeys <- key
			}(i)
		}

		// Collect results
		keys := make([]string, 0, numUploads)
		for i := 0; i < numUploads; i++ {
			select {
			case key := <-uploadedKeys:
				keys = append(keys, key)
			case err := <-errors:
				t.Fatalf("Upload failed: %v", err)
			case <-time.After(30 * time.Second):
				t.Fatal("Upload timeout")
			}
		}

		assert.Len(t, keys, numUploads)

		// Cleanup
		defer func() {
			for _, key := range keys {
				_ = storage.Delete(ctx, key)
			}
		}()
	})
}
