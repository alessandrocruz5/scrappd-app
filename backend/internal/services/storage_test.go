package services

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewR2Storage tests the creation of R2Storage client
func TestNewR2Storage(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard) // Suppress logs during testing

	tests := []struct {
		name        string
		config      *config.StorageConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid configuration",
			config: &config.StorageConfig{
				Endpoint:        "https://test-account.r2.cloudflarestorage.com",
				AccessKeyID:     "test-access-key",
				SecretAccessKey: "test-secret-key",
				BucketName:      "test-bucket",
				Region:          "auto",
			},
			expectError: false,
		},
		{
			name: "missing endpoint",
			config: &config.StorageConfig{
				Endpoint:        "",
				AccessKeyID:     "test-access-key",
				SecretAccessKey: "test-secret-key",
				BucketName:      "test-bucket",
				Region:          "auto",
			},
			expectError: true,
			errorMsg:    "storage endpoint is required",
		},
		{
			name: "missing access key ID",
			config: &config.StorageConfig{
				Endpoint:        "https://test-account.r2.cloudflarestorage.com",
				AccessKeyID:     "",
				SecretAccessKey: "test-secret-key",
				BucketName:      "test-bucket",
				Region:          "auto",
			},
			expectError: true,
			errorMsg:    "storage access key ID is required",
		},
		{
			name: "missing secret access key",
			config: &config.StorageConfig{
				Endpoint:        "https://test-account.r2.cloudflarestorage.com",
				AccessKeyID:     "test-access-key",
				SecretAccessKey: "",
				BucketName:      "test-bucket",
				Region:          "auto",
			},
			expectError: true,
			errorMsg:    "storage secret access key is required",
		},
		{
			name: "missing bucket name",
			config: &config.StorageConfig{
				Endpoint:        "https://test-account.r2.cloudflarestorage.com",
				AccessKeyID:     "test-access-key",
				SecretAccessKey: "test-secret-key",
				BucketName:      "",
				Region:          "auto",
			},
			expectError: true,
			errorMsg:    "storage bucket name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewR2Storage(tt.config, logger)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, storage)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, storage)
				assert.Equal(t, tt.config.BucketName, storage.bucketName)
				assert.NotNil(t, storage.client)
				assert.NotNil(t, storage.logger)
			}
		})
	}
}

// MockStorage is a mock implementation of the Storage interface for testing
type MockStorage struct {
	UploadFunc        func(ctx context.Context, file io.Reader, filename string, contentType string) (string, error)
	UploadWithKeyFunc func(ctx context.Context, file io.Reader, key string, contentType string) error
	DownloadFunc      func(ctx context.Context, key string) ([]byte, error)
	DeleteFunc        func(ctx context.Context, key string) error
	GetURLFunc        func(ctx context.Context, key string, expiry time.Duration) (string, error)
	ExistsFunc        func(ctx context.Context, key string) (bool, error)
	ListFunc          func(ctx context.Context, prefix string) ([]string, error)
	HealthCheckFunc   func(ctx context.Context) error
}

func (m *MockStorage) Upload(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, file, filename, contentType)
	}
	return "", nil
}

func (m *MockStorage) UploadWithKey(ctx context.Context, file io.Reader, key string, contentType string) error {
	if m.UploadWithKeyFunc != nil {
		return m.UploadWithKeyFunc(ctx, file, key, contentType)
	}
	return nil
}

func (m *MockStorage) Download(ctx context.Context, key string) ([]byte, error) {
	if m.DownloadFunc != nil {
		return m.DownloadFunc(ctx, key)
	}
	return nil, nil
}

func (m *MockStorage) Delete(ctx context.Context, key string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, key)
	}
	return nil
}

func (m *MockStorage) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if m.GetURLFunc != nil {
		return m.GetURLFunc(ctx, key, expiry)
	}
	return "", nil
}

func (m *MockStorage) Exists(ctx context.Context, key string) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, key)
	}
	return false, nil
}

func (m *MockStorage) List(ctx context.Context, prefix string) ([]string, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, prefix)
	}
	return nil, nil
}

func (m *MockStorage) HealthCheck(ctx context.Context) error {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc(ctx)
	}
	return nil
}

// TestMockStorage tests the mock storage implementation
func TestMockStorage(t *testing.T) {
	ctx := context.Background()

	t.Run("Upload", func(t *testing.T) {
		mock := &MockStorage{
			UploadFunc: func(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
				assert.Equal(t, "test.jpg", filename)
				assert.Equal(t, "image/jpeg", contentType)
				return "uploads/2024/01/15/test-uuid.jpg", nil
			},
		}

		file := strings.NewReader("test content")
		key, err := mock.Upload(ctx, file, "test.jpg", "image/jpeg")

		assert.NoError(t, err)
		assert.Equal(t, "uploads/2024/01/15/test-uuid.jpg", key)
	})

	t.Run("UploadWithKey", func(t *testing.T) {
		mock := &MockStorage{
			UploadWithKeyFunc: func(ctx context.Context, file io.Reader, key string, contentType string) error {
				assert.Equal(t, "custom/path/file.jpg", key)
				assert.Equal(t, "image/jpeg", contentType)
				return nil
			},
		}

		file := strings.NewReader("test content")
		err := mock.UploadWithKey(ctx, file, "custom/path/file.jpg", "image/jpeg")

		assert.NoError(t, err)
	})

	t.Run("Download", func(t *testing.T) {
		expectedContent := []byte("test file content")
		mock := &MockStorage{
			DownloadFunc: func(ctx context.Context, key string) ([]byte, error) {
				assert.Equal(t, "uploads/test.jpg", key)
				return expectedContent, nil
			},
		}

		content, err := mock.Download(ctx, "uploads/test.jpg")

		assert.NoError(t, err)
		assert.Equal(t, expectedContent, content)
	})

	t.Run("Delete", func(t *testing.T) {
		mock := &MockStorage{
			DeleteFunc: func(ctx context.Context, key string) error {
				assert.Equal(t, "uploads/test.jpg", key)
				return nil
			},
		}

		err := mock.Delete(ctx, "uploads/test.jpg")

		assert.NoError(t, err)
	})

	t.Run("GetURL", func(t *testing.T) {
		mock := &MockStorage{
			GetURLFunc: func(ctx context.Context, key string, expiry time.Duration) (string, error) {
				assert.Equal(t, "uploads/test.jpg", key)
				assert.Equal(t, 1*time.Hour, expiry)
				return "https://example.com/presigned-url", nil
			},
		}

		url, err := mock.GetURL(ctx, "uploads/test.jpg", 1*time.Hour)

		assert.NoError(t, err)
		assert.Equal(t, "https://example.com/presigned-url", url)
	})

	t.Run("Exists", func(t *testing.T) {
		mock := &MockStorage{
			ExistsFunc: func(ctx context.Context, key string) (bool, error) {
				assert.Equal(t, "uploads/test.jpg", key)
				return true, nil
			},
		}

		exists, err := mock.Exists(ctx, "uploads/test.jpg")

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("List", func(t *testing.T) {
		expectedKeys := []string{"uploads/file1.jpg", "uploads/file2.jpg"}
		mock := &MockStorage{
			ListFunc: func(ctx context.Context, prefix string) ([]string, error) {
				assert.Equal(t, "uploads/", prefix)
				return expectedKeys, nil
			},
		}

		keys, err := mock.List(ctx, "uploads/")

		assert.NoError(t, err)
		assert.Equal(t, expectedKeys, keys)
	})
}

// TestStorageInterface verifies that R2Storage implements Storage interface
func TestStorageInterface(t *testing.T) {
	var _ Storage = (*R2Storage)(nil)
	var _ Storage = (*MockStorage)(nil)
}

// TestUploadKeyGeneration tests that Upload generates proper keys
func TestUploadKeyGeneration(t *testing.T) {
	// This test would require actual R2 connection, so we'll test the key format pattern
	// The key should match: uploads/YYYY/MM/DD/UUID.ext
	t.Run("key format validation", func(t *testing.T) {
		// We can't test actual upload without R2, but we can test the pattern
		// This is covered in integration tests
		t.Skip("Requires integration test with actual R2 connection")
	})
}

// BenchmarkMockStorage benchmarks the mock storage operations
func BenchmarkMockStorage(b *testing.B) {
	ctx := context.Background()
	mock := &MockStorage{
		UploadFunc: func(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
			return "uploads/test.jpg", nil
		},
		DownloadFunc: func(ctx context.Context, key string) ([]byte, error) {
			return []byte("test content"), nil
		},
	}

	b.Run("Upload", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			file := bytes.NewReader([]byte("test content"))
			_, _ = mock.Upload(ctx, file, "test.jpg", "image/jpeg")
		}
	})

	b.Run("Download", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = mock.Download(ctx, "uploads/test.jpg")
		}
	})
}

// TestContextCancellation tests context cancellation handling
func TestContextCancellation(t *testing.T) {
	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		mock := &MockStorage{
			UploadFunc: func(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
				// Check if context is cancelled
				if ctx.Err() != nil {
					return "", ctx.Err()
				}
				return "uploads/test.jpg", nil
			},
		}

		file := strings.NewReader("test content")
		_, err := mock.Upload(ctx, file, "test.jpg", "image/jpeg")

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("timeout context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Wait for timeout

		mock := &MockStorage{
			DownloadFunc: func(ctx context.Context, key string) ([]byte, error) {
				if ctx.Err() != nil {
					return nil, ctx.Err()
				}
				return []byte("content"), nil
			},
		}

		_, err := mock.Download(ctx, "uploads/test.jpg")

		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
	})
}

// TestFileOperations tests various file operation scenarios
func TestFileOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("upload and download same file", func(t *testing.T) {
		originalContent := []byte("test file content")
		var uploadedKey string

		mock := &MockStorage{
			UploadFunc: func(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
				uploadedKey = "uploads/2024/01/15/test.jpg"
				return uploadedKey, nil
			},
			DownloadFunc: func(ctx context.Context, key string) ([]byte, error) {
				require.Equal(t, uploadedKey, key)
				return originalContent, nil
			},
		}

		// Upload
		file := bytes.NewReader(originalContent)
		key, err := mock.Upload(ctx, file, "test.jpg", "image/jpeg")
		require.NoError(t, err)

		// Download
		downloaded, err := mock.Download(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, originalContent, downloaded)
	})

	t.Run("delete uploaded file", func(t *testing.T) {
		deletedKeys := make(map[string]bool)

		mock := &MockStorage{
			UploadFunc: func(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
				return "uploads/test.jpg", nil
			},
			DeleteFunc: func(ctx context.Context, key string) error {
				deletedKeys[key] = true
				return nil
			},
			ExistsFunc: func(ctx context.Context, key string) (bool, error) {
				return !deletedKeys[key], nil
			},
		}

		// Upload
		file := strings.NewReader("test content")
		key, err := mock.Upload(ctx, file, "test.jpg", "image/jpeg")
		require.NoError(t, err)

		// Check exists
		exists, err := mock.Exists(ctx, key)
		require.NoError(t, err)
		assert.True(t, exists)

		// Delete
		err = mock.Delete(ctx, key)
		require.NoError(t, err)

		// Check not exists
		exists, err = mock.Exists(ctx, key)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("list files with prefix", func(t *testing.T) {
		files := map[string]bool{
			"uploads/user1/file1.jpg": true,
			"uploads/user1/file2.jpg": true,
			"uploads/user2/file1.jpg": true,
		}

		mock := &MockStorage{
			ListFunc: func(ctx context.Context, prefix string) ([]string, error) {
				var keys []string
				for key := range files {
					if strings.HasPrefix(key, prefix) {
						keys = append(keys, key)
					}
				}
				return keys, nil
			},
		}

		// List user1 files
		keys, err := mock.List(ctx, "uploads/user1/")
		require.NoError(t, err)
		assert.Len(t, keys, 2)

		// List all uploads
		keys, err = mock.List(ctx, "uploads/")
		require.NoError(t, err)
		assert.Len(t, keys, 3)
	})
}

// TestContentTypes tests various content type scenarios
func TestContentTypes(t *testing.T) {
	ctx := context.Background()

	contentTypes := []struct {
		name        string
		contentType string
		filename    string
	}{
		{"JPEG image", "image/jpeg", "test.jpg"},
		{"PNG image", "image/png", "test.png"},
		{"WebP image", "image/webp", "test.webp"},
		{"GIF image", "image/gif", "test.gif"},
		{"PDF document", "application/pdf", "document.pdf"},
		{"Text file", "text/plain", "notes.txt"},
	}

	for _, tc := range contentTypes {
		t.Run(tc.name, func(t *testing.T) {
			mock := &MockStorage{
				UploadFunc: func(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
					assert.Equal(t, tc.filename, filename)
					assert.Equal(t, tc.contentType, contentType)
					return "uploads/" + filename, nil
				},
			}

			file := strings.NewReader("test content")
			key, err := mock.Upload(ctx, file, tc.filename, tc.contentType)

			assert.NoError(t, err)
			assert.Contains(t, key, tc.filename)
		})
	}
}

// TestPresignedURLExpiry tests presigned URL expiry durations
func TestPresignedURLExpiry(t *testing.T) {
	ctx := context.Background()

	expiryDurations := []time.Duration{
		1 * time.Minute,
		15 * time.Minute,
		1 * time.Hour,
		24 * time.Hour,
		7 * 24 * time.Hour,
	}

	for _, expiry := range expiryDurations {
		t.Run(expiry.String(), func(t *testing.T) {
			mock := &MockStorage{
				GetURLFunc: func(ctx context.Context, key string, exp time.Duration) (string, error) {
					assert.Equal(t, expiry, exp)
					return "https://example.com/presigned", nil
				},
			}

			url, err := mock.GetURL(ctx, "uploads/test.jpg", expiry)

			assert.NoError(t, err)
			assert.NotEmpty(t, url)
		})
	}
}
