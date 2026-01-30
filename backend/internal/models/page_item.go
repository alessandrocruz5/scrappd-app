package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PageItem represents an item positioned on a page.
type PageItem struct {
	ID        uuid.UUID       `json:"id"`
	PageID    uuid.UUID       `json:"page_id"`
	ItemID    uuid.UUID       `json:"item_id"`
	PositionX float64         `json:"position_x"`
	PositionY float64         `json:"position_y"`
	Width     float64         `json:"width"`
	Height    float64         `json:"height"`
	Rotation  float64         `json:"rotation"`
	ZIndex    int             `json:"z_index"`
	Opacity   float64         `json:"opacity"`
	Filters   json.RawMessage `json:"filters,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// PageItemUpdate holds optional fields to update.
type PageItemUpdate struct {
	ID        uuid.UUID       `json:"-"`
	PageID    uuid.UUID       `json:"-"`
	PositionX *float64        `json:"position_x,omitempty"`
	PositionY *float64        `json:"position_y,omitempty"`
	Width     *float64        `json:"width,omitempty"`
	Height    *float64        `json:"height,omitempty"`
	Rotation  *float64        `json:"rotation,omitempty"`
	ZIndex    *int            `json:"z_index,omitempty"`
	Opacity   *float64        `json:"opacity,omitempty"`
	Filters   json.RawMessage `json:"filters,omitempty"`
}
