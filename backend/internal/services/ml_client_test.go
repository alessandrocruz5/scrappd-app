package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	expectedResponse := models.RemoveBackgroundResponse{
		ProcessedImage: "base64processedimage",
		Metadata: models.BackgroundRemovalMeta{
			ProcessingTime: 14.5,
			Model:          "BiRefNet",
			OriginalSize:   models.Size{Width: 1920, Height: 1080},
			ProcessedSize:  models.Size{Width: 1920, Height: 1080},
		},
	}

	client, server := setupTestMLClient(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/remove-background", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify request body
		var req models.RemoveBackgroundRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "base64imagedata", req.ImageData)
		assert.Equal(t, "png", req.Format)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResponse)
	})
	defer server.Close()

	ctx := context.Background()
	response, err := client.RemoveBackground(ctx, "base64imagedata", "png")

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, expectedResponse.ProcessedImage, response.ProcessedImage)
	assert.Equal(t, expectedResponse.Metadata.ProcessingTime, response.Metadata.ProcessingTime)
	assert.Equal(t, expectedResponse.Metadata.Model, response.Metadata.Model)
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
	client, server := setupTestMLClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Detail: "Invalid image format",
		})
	})
	defer server.Close()

	ctx := context.Background()
	response, err := client.RemoveBackground(ctx, "base64imagedata", "png")

	assert.Nil(t, response)
	require.Error(t, err)

	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrCodeMLServiceError, appErr.Code)
	assert.Contains(t, appErr.Message, "Invalid image format")
}

func TestRemoveBackground_ServerError(t *testing.T) {
	client, server := setupTestMLClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Detail: "Internal server error",
		})
	})
	defer server.Close()

	ctx := context.Background()
	response, err := client.RemoveBackground(ctx, "base64imagedata", "png")

	assert.Nil(t, response)
	require.Error(t, err)

	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrCodeMLServiceError, appErr.Code)
}

func TestRemoveBackground_RetryOnServerError(t *testing.T) {
	attemptCount := 0
	expectedResponse := models.RemoveBackgroundResponse{
		ProcessedImage: "base64processedimage",
		Metadata: models.BackgroundRemovalMeta{
			ProcessingTime: 14.5,
			Model:          "BiRefNet",
		},
	}

	client, server := setupTestMLClient(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++

		// Fail on first attempt, succeed on second
		if attemptCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResponse)
	})
	defer server.Close()

	ctx := context.Background()
	response, err := client.RemoveBackground(ctx, "base64imagedata", "png")

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 2, attemptCount) // Should have retried once
	assert.Equal(t, expectedResponse.ProcessedImage, response.ProcessedImage)
}

func TestRemoveBackground_NoRetryOnClientError(t *testing.T) {
	attemptCount := 0

	client, server := setupTestMLClient(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{
			Detail: "Invalid image",
		})
	})
	defer server.Close()

	ctx := context.Background()
	response, err := client.RemoveBackground(ctx, "base64imagedata", "png")

	assert.Nil(t, response)
	require.Error(t, err)
	assert.Equal(t, 1, attemptCount) // Should NOT retry on client errors
}

func TestRemoveBackground_AllRetriesFail(t *testing.T) {
	attemptCount := 0

	client, server := setupTestMLClient(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	defer server.Close()

	ctx := context.Background()
	response, err := client.RemoveBackground(ctx, "base64imagedata", "png")

	assert.Nil(t, response)
	require.Error(t, err)
	assert.Equal(t, 3, attemptCount) // Initial attempt + 2 retries

	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrCodeMLServiceError, appErr.Code)
	assert.Contains(t, appErr.Message, "Failed after 3 attempts")
}

func TestRemoveBackground_ContextCancellation(t *testing.T) {
	client, server := setupTestMLClient(func(w http.ResponseWriter, r *http.Request) {
		// Delay to allow context cancellation
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	response, err := client.RemoveBackground(ctx, "base64imagedata", "png")

	assert.Nil(t, response)
	require.Error(t, err)
}

func TestRemoveBackground_InvalidJSONResponse(t *testing.T) {
	client, server := setupTestMLClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	})
	defer server.Close()

	ctx := context.Background()
	response, err := client.RemoveBackground(ctx, "base64imagedata", "png")

	assert.Nil(t, response)
	require.Error(t, err)

	appErr, ok := err.(*utils.AppError)
	require.True(t, ok)
	assert.Equal(t, utils.ErrCodeMLServiceError, appErr.Code)
	assert.Contains(t, appErr.Message, "Failed to decode response")
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
