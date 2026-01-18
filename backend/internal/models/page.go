package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Page represents a scrapbook page with layout data.
type Page struct {
	ID                uuid.UUID       `json:"id"`
	ProjectID         uuid.UUID       `json:"project_id"`
	PageNumber        int             `json:"page_number"`
	Title             *string         `json:"title,omitempty"`
	CanvasWidth       int             `json:"canvas_width"`
	CanvasHeight      int             `json:"canvas_height"`
	BackgroundColor   string          `json:"background_color"`
	BackgroundImageURL *string         `json:"background_image_url,omitempty"`
	BackgroundPattern *string         `json:"background_pattern,omitempty"`
	LayoutTemplate    json.RawMessage `json:"layout_template,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// PageUpdate represents mutable page fields.
type PageUpdate struct {
	ID                uuid.UUID       `json:"-"`
	PageNumber        *int            `json:"page_number,omitempty"`
	Title             *string         `json:"title,omitempty"`
	CanvasWidth       *int            `json:"canvas_width,omitempty"`
	CanvasHeight      *int            `json:"canvas_height,omitempty"`
	BackgroundColor   *string         `json:"background_color,omitempty"`
	BackgroundImageURL *string         `json:"background_image_url,omitempty"`
	BackgroundPattern *string         `json:"background_pattern,omitempty"`
	LayoutTemplate    json.RawMessage `json:"layout_template,omitempty"`
}
