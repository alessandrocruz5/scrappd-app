package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveBackgroundRequest_JSON(t *testing.T) {
	req := RemoveBackgroundRequest{
		ImageData: "base64encodedimage",
		Format:    "png",
	}

	jsonData, err := json.Marshal(req)
	require.NoError(t, err)

	var decoded RemoveBackgroundRequest
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, req.ImageData, decoded.ImageData)
	assert.Equal(t, req.Format, decoded.Format)
}

func TestRemoveBackgroundResponse_JSON(t *testing.T) {
	resp := RemoveBackgroundResponse{
		ProcessedImage: "base64processedimage",
		Metadata: BackgroundRemovalMeta{
			ProcessingTime: 14.5,
			Model:          "BiRefNet",
			OriginalSize:   Size{Width: 1920, Height: 1080},
			ProcessedSize:  Size{Width: 1920, Height: 1080},
		},
	}

	jsonData, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded RemoveBackgroundResponse
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, resp.ProcessedImage, decoded.ProcessedImage)
	assert.Equal(t, resp.Metadata.ProcessingTime, decoded.Metadata.ProcessingTime)
	assert.Equal(t, resp.Metadata.Model, decoded.Metadata.Model)
	assert.Equal(t, 1920, decoded.Metadata.OriginalSize.Width)
}

func TestHealthCheckResponse_JSON(t *testing.T) {
	now := time.Now()
	resp := HealthCheckResponse{
		Status:  "healthy",
		Version: "1.0.0",
		Model:   "BiRefNet",
		Time:    now,
	}

	jsonData, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded HealthCheckResponse
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, resp.Status, decoded.Status)
	assert.Equal(t, resp.Version, decoded.Version)
	assert.Equal(t, resp.Model, decoded.Model)
}

func TestErrorResponse_JSON(t *testing.T) {
	resp := ErrorResponse{
		Detail: "Invalid image format",
	}

	jsonData, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded ErrorResponse
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, resp.Detail, decoded.Detail)
}
