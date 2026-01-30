package integration

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	defaultMLTimeout = 30 * time.Second
)

func TestMLServiceHealth(t *testing.T) {
	baseURL := os.Getenv("ML_SERVICE_URL")
	if baseURL == "" {
		t.Skip("ML_SERVICE_URL not set; skipping ML integration tests")
	}

	client := &http.Client{Timeout: defaultMLTimeout}
	req, err := http.NewRequest(http.MethodGet, baseURL+"/health", nil)
	if err != nil {
		t.Fatalf("failed to create health request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from /health, got %d", resp.StatusCode)
	}
}

func TestMLServiceProcess(t *testing.T) {
	baseURL := os.Getenv("ML_SERVICE_URL")
	if baseURL == "" {
		t.Skip("ML_SERVICE_URL not set; skipping ML integration tests")
	}

	// Minimal 1x1 PNG bytes.
	pngBytes := []byte{
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

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.png")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write(pngBytes); err != nil {
		t.Fatalf("failed to write PNG bytes: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultMLTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/process", body)
	if err != nil {
		t.Fatalf("failed to create process request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: defaultMLTimeout}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("process request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from /process, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "image/png" {
		t.Fatalf("expected image/png response, got %q", contentType)
	}
}
