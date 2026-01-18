package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/config"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestMLClient creates a test ML client with a mock server
func setupTestMLClient(handler http.HandlerFunc) (*mlClient, *httptest.Server) {
	server := httptest.NewServer(handler)

	cfg := &config.MLServiceConfig{
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 2,
		RetryDelay: 100 * time.Millisecond,
	}

	client := NewMLClient(cfg).(*mlClient)
	return client, server
}

func TestRemoveBackground_Success(t *testing.T) {
	inputBase64 := base64.StdEncoding.EncodeToString(testPNGBytes())
	processedBytes := []byte("processed-image")
	expectedBase64 := base64.StdEncoding.EncodeToString(processedBytes)

	client, server := setupTestMLClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/process", r.URL.Path)
		assert.True(t, strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data"))

		err := r.ParseMultipartForm(10 << 20)
		require.NoError(t, err)

		file, header, err := r.FormFile("file")
		require.NoError(t, err)
		defer file.Close()

		assert.Equal(t, "image.png", header.Filename)
		assert.Equal(t, "image/png", header.Header.Get("Content-Type"))
		_, err = io.ReadAll(file)
		require.NoError(t, err)

		w.Header().Set("X-Processing-Time", "14.5")
		w.WriteHeader(http.StatusOK)
		w.Write(processedBytes)
	})
	defer server.Close()

	ctx := context.Background()
	response, err := client.RemoveBackground(ctx, inputBase64, "png")

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, expectedBase64, response.ProcessedImage)
	assert.Equal(t, 14.5, response.Metadata.ProcessingTime)
	assert.Equal(t, "birefnet-general", response.Metadata.Model)
}

func TestRemoveBackground_EmptyImageData(t *testing.T) {
	client, server := setupTestMLClient(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Should not make HTTP request with empty image data")
	})
	defer server.Close()

	ctx := context.Background()
	response, err := client.RemoveBackground(ctx, "", "png")

	assert.Nil(t, response)
	require.Error(t, err)

	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrCodeBadRequest, appErr.Code)
	assert.Contains(t, appErr.Message, "Image data is required")
}

func TestRemoveBackground_MLServiceError(t *testing.T) {
	inputBase64 := base64.StdEncoding.EncodeToString(testPNGBytes())
	client, server := setupTestMLClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Detail: "Invalid image format",
		})
	})
	defer server.Close()

	ctx := context.Background()
	response, err := client.RemoveBackground(ctx, inputBase64, "png")

	assert.Nil(t, response)
	require.Error(t, err)

	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrCodeMLServiceError, appErr.Code)
	assert.Contains(t, appErr.Message, "Invalid image format")
}

func TestRemoveBackground_ServerError(t *testing.T) {
	inputBase64 := base64.StdEncoding.EncodeToString(testPNGBytes())
	client, server := setupTestMLClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Detail: "Internal server error",
		})
	})
	defer server.Close()

	ctx := context.Background()
	response, err := client.RemoveBackground(ctx, inputBase64, "png")

	assert.Nil(t, response)
	require.Error(t, err)

	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrCodeMLServiceError, appErr.Code)
}

func TestRemoveBackground_RetryOnServerError(t *testing.T) {
	attemptCount := 0
	inputBase64 := base64.StdEncoding.EncodeToString(testPNGBytes())
	processedBytes := []byte("processed-image")
	expectedBase64 := base64.StdEncoding.EncodeToString(processedBytes)

	client, server := setupTestMLClient(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++

		// Fail on first attempt, succeed on second
		if attemptCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(processedBytes)
	})
	defer server.Close()

	ctx := context.Background()
	response, err := client.RemoveBackground(ctx, inputBase64, "png")

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 2, attemptCount) // Should have retried once
	assert.Equal(t, expectedBase64, response.ProcessedImage)
}

func TestRemoveBackground_NoRetryOnClientError(t *testing.T) {
	attemptCount := 0
	inputBase64 := base64.StdEncoding.EncodeToString(testPNGBytes())

	client, server := setupTestMLClient(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Detail: "Invalid image",
		})
	})
	defer server.Close()

	ctx := context.Background()
	response, err := client.RemoveBackground(ctx, inputBase64, "png")

	assert.Nil(t, response)
	require.Error(t, err)
	assert.Equal(t, 1, attemptCount) // Should NOT retry on client errors
}

func TestRemoveBackground_AllRetriesFail(t *testing.T) {
	attemptCount := 0
	inputBase64 := base64.StdEncoding.EncodeToString(testPNGBytes())

	client, server := setupTestMLClient(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	defer server.Close()

	ctx := context.Background()
	response, err := client.RemoveBackground(ctx, inputBase64, "png")

	assert.Nil(t, response)
	require.Error(t, err)
	assert.Equal(t, 3, attemptCount) // Initial attempt + 2 retries

	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrCodeMLServiceError, appErr.Code)
	assert.Contains(t, appErr.Message, "Failed after 3 attempts")
}

func TestRemoveBackground_ContextCancellation(t *testing.T) {
	inputBase64 := base64.StdEncoding.EncodeToString(testPNGBytes())
	client, server := setupTestMLClient(func(w http.ResponseWriter, r *http.Request) {
		// Delay to allow context cancellation
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	response, err := client.RemoveBackground(ctx, inputBase64, "png")

	assert.Nil(t, response)
	require.Error(t, err)
}

func TestRemoveBackground_InvalidJSONResponse(t *testing.T) {
	inputBase64 := base64.StdEncoding.EncodeToString(testPNGBytes())
	client, server := setupTestMLClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid json"))
	})
	defer server.Close()

	ctx := context.Background()
	response, err := client.RemoveBackground(ctx, inputBase64, "png")

	assert.Nil(t, response)
	require.Error(t, err)

	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrCodeMLServiceError, appErr.Code)
	assert.Contains(t, appErr.Message, "invalid json")
}

func testPNGBytes() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}
}

func TestHealthCheck_Success(t *testing.T) {
	expectedResponse := models.HealthCheckResponse{
		Status:  "healthy",
		Version: "1.0.0",
		Model:   "BiRefNet",
		Time:    time.Now(),
	}

	client, server := setupTestMLClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/health", r.URL.Path)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResponse)
	})
	defer server.Close()

	ctx := context.Background()
	response, err := client.HealthCheck(ctx)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, expectedResponse.Status, response.Status)
	assert.Equal(t, expectedResponse.Version, response.Version)
	assert.Equal(t, expectedResponse.Model, response.Model)
}

func TestHealthCheck_ServiceUnavailable(t *testing.T) {
	client, server := setupTestMLClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	defer server.Close()

	ctx := context.Background()
	response, err := client.HealthCheck(ctx)

	assert.Nil(t, response)
	require.Error(t, err)

	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrCodeServiceUnavailable, appErr.Code)
}

func TestHealthCheck_InvalidResponse(t *testing.T) {
	client, server := setupTestMLClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	})
	defer server.Close()

	ctx := context.Background()
	response, err := client.HealthCheck(ctx)

	assert.Nil(t, response)
	require.Error(t, err)

	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrCodeMLServiceError, appErr.Code)
}
