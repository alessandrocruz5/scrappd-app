package models

import (
	"time"

	"github.com/google/uuid"
)

// Item represents a processed image and its metadata.
type Item struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`

	OriginalImageKey  string `json:"original_image_key"`
	OriginalImageURL  string `json:"original_image_url"`
	OriginalFileSize  *int64 `json:"original_file_size_bytes,omitempty"`
	OriginalWidth     *int   `json:"original_width,omitempty"`
	OriginalHeight    *int   `json:"original_height,omitempty"`
	ProcessedImageKey *string `json:"processed_image_key,omitempty"`
	ProcessedImageURL *string `json:"processed_image_url,omitempty"`
	ProcessedFileSize *int64  `json:"processed_file_size_bytes,omitempty"`

	ProcessingStatus     string     `json:"processing_status"`
	MLModelVersion       *string    `json:"ml_model_version,omitempty"`
	ProcessingStartedAt  *time.Time `json:"processing_started_at,omitempty"`
	ProcessingCompletedAt *time.Time `json:"processing_completed_at,omitempty"`
	ProcessingError      *string    `json:"processing_error,omitempty"`

	MimeType     *string  `json:"mime_type,omitempty"`
	ItemName     *string  `json:"item_name,omitempty"`
	ItemCategory *string  `json:"item_category,omitempty"`
	Tags         []string `json:"tags,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}
