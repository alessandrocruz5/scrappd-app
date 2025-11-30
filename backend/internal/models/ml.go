package models

import "time"

// RemoveBackgroundRequest represents a request to remove background from an image
type RemoveBackgroundRequest struct {
	ImageData string `json:"image_data" binding:"required"` // Base64 encoded image
	Format    string `json:"format,omitempty"`              // Output format (png, webp)
}

// RemoveBackgroundResponse represents the response from the ML service
type RemoveBackgroundResponse struct {
	ProcessedImage string                `json:"processed_image"` // Base64 encoded result
	Metadata       BackgroundRemovalMeta `json:"metadata"`
}

// BackgroundRemovalMeta contains metadata about the background removal process
type BackgroundRemovalMeta struct {
	ProcessingTime float64 `json:"processing_time"` // Time in seconds
	Model          string  `json:"model"`           // Model used (e.g., "BiRefNet")
	OriginalSize   Size    `json:"original_size"`
	ProcessedSize  Size    `json:"processed_size"`
}

// Size represents image dimensions
type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// HealthCheckResponse represents the ML service health check response
type HealthCheckResponse struct {
	Status  string    `json:"status"`
	Version string    `json:"version"`
	Model   string    `json:"model"`
	Time    time.Time `json:"time"`
}

// ErrorResponse represents an error response from the ML service
type ErrorResponse struct {
	Detail string `json:"detail"`
}
