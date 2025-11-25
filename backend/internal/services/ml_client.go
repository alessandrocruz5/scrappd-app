package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/config"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
)

// MLClient is the interface for the ML service client
type MLClient interface {
	RemoveBackground(ctx context.Context, imageData string, format string) (*models.RemoveBackgroundResponse, error)
	HealthCheck(ctx context.Context) (*models.HealthCheckResponse, error)
}

// mlClient implements the MLClient interface
type mlClient struct {
	httpClient *http.Client
	baseURL    string
	maxRetries int
	retryDelay time.Duration
}

// NewMLClient creates a new ML service client
func NewMLClient(cfg *config.MLServiceConfig) MLClient {
	return &mlClient{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		baseURL:    cfg.BaseURL,
		maxRetries: cfg.MaxRetries,
		retryDelay: cfg.RetryDelay,
	}
}

// RemoveBackground sends an image to the ML service for background removal
func (c *mlClient) RemoveBackground(ctx context.Context, imageData string, format string) (*models.RemoveBackgroundResponse, error) {
	if imageData == "" {
		return nil, utils.ErrBadRequest("Image data is required", nil)
	}

	req := models.RemoveBackgroundRequest{
		ImageData: imageData,
		Format:    format,
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retrying
			select {
			case <-ctx.Done():
				return nil, utils.ErrMLService("Request cancelled", ctx.Err())
			case <-time.After(c.retryDelay):
			}
		}

		response, err := c.doRemoveBackground(ctx, req)
		if err == nil {
			return response, nil
		}

		lastErr = err

		// Don't retry on client errors (4xx)
		if appErr, ok := err.(*utils.AppError); ok {
			if appErr.StatusCode >= 400 && appErr.StatusCode < 500 {
				return nil, err
			}
		}
	}

	return nil, utils.ErrMLService(
		fmt.Sprintf("Failed after %d attempts", c.maxRetries+1),
		lastErr,
	)
}

// doRemoveBackground performs the actual HTTP request
func (c *mlClient) doRemoveBackground(ctx context.Context, req models.RemoveBackgroundRequest) (*models.RemoveBackgroundResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, utils.ErrBadRequest("Failed to encode request", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/remove-background", c.baseURL),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, utils.ErrInternalServer("Failed to create request", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, utils.ErrServiceUnavailable("ML", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, utils.ErrMLService("Failed to read response", err)
	}

	// Handle non-2xx status codes
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		var errResp models.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Detail != "" {
			return nil, utils.NewAppError(
				utils.ErrCodeMLServiceError,
				errResp.Detail,
				httpResp.StatusCode,
				nil,
			)
		}
		return nil, utils.NewAppError(
			utils.ErrCodeMLServiceError,
			fmt.Sprintf("ML service returned status %d", httpResp.StatusCode),
			httpResp.StatusCode,
			nil,
		)
	}

	var response models.RemoveBackgroundResponse
	if err := json.Unmarshal(body, &response); err != nil {
		// Treat JSON decode errors as client errors (don't retry)
		return nil, utils.NewAppError(
			utils.ErrCodeMLServiceError,
			"Failed to decode response",
			http.StatusBadRequest, // Changed from 500 to 400 to prevent retries
			err,
		)
	}

	return &response, nil
}

// HealthCheck checks if the ML service is healthy
func (c *mlClient) HealthCheck(ctx context.Context) (*models.HealthCheckResponse, error) {
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/health", c.baseURL),
		nil,
	)
	if err != nil {
		return nil, utils.ErrInternalServer("Failed to create health check request", err)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, utils.ErrServiceUnavailable("ML", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, utils.ErrServiceUnavailable(
			"ML",
			fmt.Errorf("health check returned status %d", httpResp.StatusCode),
		)
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, utils.ErrMLService("Failed to read health check response", err)
	}

	var response models.HealthCheckResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, utils.ErrMLService("Failed to decode health check response", err)
	}

	return &response, nil
}
