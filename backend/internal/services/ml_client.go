package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/config"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"google.golang.org/api/idtoken"
)

// MLClient is the interface for the ML service client
type MLClient interface {
	RemoveBackground(ctx context.Context, imageData string, format string) (*models.RemoveBackgroundResponse, error)
	HealthCheck(ctx context.Context) (*models.HealthCheckResponse, error)
}

// mlClient implements the MLClient interface
type mlClient struct {
	httpClient  *http.Client
	baseURL     string
	maxRetries  int
	retryDelay  time.Duration
	useGCPAuth  bool
	tokenSource func(ctx context.Context) (string, error)
}

// NewMLClient creates a new ML service client
func NewMLClient(cfg *config.MLServiceConfig) MLClient {
	client := &mlClient{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		baseURL:    cfg.BaseURL,
		maxRetries: cfg.MaxRetries,
		retryDelay: cfg.RetryDelay,
		useGCPAuth: os.Getenv("ENVIRONMENT") == "production",
	}

	// Set up GCP authentication for production
	if client.useGCPAuth {
		client.tokenSource = func(ctx context.Context) (string, error) {
			// Create an ID token for the target service
			ts, err := idtoken.NewTokenSource(ctx, cfg.BaseURL)
			if err != nil {
				return "", fmt.Errorf("failed to create token source: %w", err)
			}
			token, err := ts.Token()
			if err != nil {
				return "", fmt.Errorf("failed to get token: %w", err)
			}
			return token.AccessToken, nil
		}
	}

	return client
}

// addAuthHeader adds GCP authentication header if in production
func (c *mlClient) addAuthHeader(ctx context.Context, req *http.Request) error {
	if !c.useGCPAuth || c.tokenSource == nil {
		return nil
	}

	token, err := c.tokenSource(ctx)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

// RemoveBackground sends an image to the ML service for background removal
func (c *mlClient) RemoveBackground(ctx context.Context, imageData string, format string) (*models.RemoveBackgroundResponse, error) {
	if imageData == "" {
		return nil, utils.ErrBadRequest("Image data is required", nil)
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, utils.ErrMLService("Request cancelled", ctx.Err())
			case <-time.After(c.retryDelay):
			}
		}

		response, err := c.doRemoveBackground(ctx, imageData, format)
		if err == nil {
			return response, nil
		}

		lastErr = err

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
func (c *mlClient) doRemoveBackground(ctx context.Context, imageData string, format string) (*models.RemoveBackgroundResponse, error) {
	// Decode base64 image to bytes
	imageBytes, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return nil, utils.ErrBadRequest("Failed to decode base64 image", err)
	}

	// Detect content type
	contentType := http.DetectContentType(imageBytes)

	var filename string
	switch contentType {
	case "image/jpeg":
		filename = "image.jpg"
	case "image/png":
		filename = "image.png"
	case "image/webp":
		filename = "image.webp"
	default:
		filename = "image.jpg"
		contentType = "image/jpeg"
	}

	// Create multipart form-data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	partHeader.Set("Content-Type", contentType)

	part, err := writer.CreatePart(partHeader)
	if err != nil {
		return nil, utils.ErrInternalServer("Failed to create form file", err)
	}

	if _, err := part.Write(imageBytes); err != nil {
		return nil, utils.ErrInternalServer("Failed to write image data", err)
	}

	if err := writer.Close(); err != nil {
		return nil, utils.ErrInternalServer("Failed to close multipart writer", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/process", c.baseURL),
		body,
	)
	if err != nil {
		return nil, utils.ErrInternalServer("Failed to create request", err)
	}

	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	// Add GCP authentication header
	if err := c.addAuthHeader(ctx, httpReq); err != nil {
		return nil, utils.ErrInternalServer("Failed to add auth header", err)
	}

	// Send request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, utils.ErrServiceUnavailable("ML", err)
	}
	defer httpResp.Body.Close()

	// Handle non-2xx status codes
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		var errResp models.ErrorResponse
		if err := json.Unmarshal(bodyBytes, &errResp); err == nil && errResp.Detail != "" {
			return nil, utils.NewAppError(
				utils.ErrCodeMLServiceError,
				errResp.Detail,
				httpResp.StatusCode,
				nil,
			)
		}
		return nil, utils.NewAppError(
			utils.ErrCodeMLServiceError,
			string(bodyBytes),
			httpResp.StatusCode,
			nil,
		)
	}

	// Read the PNG image bytes from response
	processedImageBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, utils.ErrMLService("Failed to read response", err)
	}

	processedImageBase64 := base64.StdEncoding.EncodeToString(processedImageBytes)

	processingTime := 0.0
	if timeStr := httpResp.Header.Get("X-Processing-Time"); timeStr != "" {
		fmt.Sscanf(timeStr, "%f", &processingTime)
	}

	return &models.RemoveBackgroundResponse{
		ProcessedImage: processedImageBase64,
		Metadata: models.BackgroundRemovalMeta{
			ProcessingTime: processingTime,
			Model:          "birefnet-general",
			OriginalSize:   models.Size{Width: 0, Height: 0},
			ProcessedSize:  models.Size{Width: 0, Height: 0},
		},
	}, nil
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

	// Add GCP authentication header
	if err := c.addAuthHeader(ctx, httpReq); err != nil {
		return nil, utils.ErrInternalServer("Failed to add auth header", err)
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
